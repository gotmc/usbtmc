// Copyright (c) 2015-2024 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gotmc/usbtmc/driver"
)

const (
	// This is a guess. The USB spec says the max value can be fetched from
	// the descriptor, but the libusb documentation says packets can be up
	// to 512 bytes.
	// Ref: https://libusb.sourceforge.io/api-1.0/libusb_packetoverflow.html
	maxPacketSize = 512

	usbtmcHeaderLen = 12
)

// Device models a USBTMC device, which includes a USB device and the required
// USBTMC attributes and methods.
type Device struct {
	usbDevice       driver.USBDevice
	bTag            byte
	termChar        byte
	termCharEnabled bool
}

// Write creates the appropriate USBMTC header, writes the header and data on
// the bulk out endpoint, and returns the number of bytes written and any
// errors.
func (d *Device) Write(p []byte) (n int, err error) {
	// FIXME(mdr): I need to change this so that I look at the size of the buf
	// being written to see if it can truly fit into one transfer, and if not
	// split it into multiple transfers.
	maxTransferSize := 512
	for pos := 0; pos < len(p); {
		d.bTag = nextbTag(d.bTag)
		thisLen := len(p[pos:])
		if thisLen > maxTransferSize-bulkOutHeaderSize {
			thisLen = maxTransferSize - bulkOutHeaderSize
		}
		header := encodeBulkOutHeader(d.bTag, uint32(thisLen), true)
		data := append(header[:], p[pos:pos+thisLen]...)
		if moduloFour := len(data) % 4; moduloFour > 0 {
			numAlignment := 4 - moduloFour
			alignment := bytes.Repeat([]byte{0x00}, numAlignment)
			data = append(data, alignment...)
		}
		_, err := d.usbDevice.Write(data)
		if err != nil {
			return pos, err
		}
		pos += thisLen
	}
	return len(p), nil
}

// Read creates and sends the header on the bulk out endpoint and then reads
// from the bulk in endpoint per USBTMC standard.
func (d *Device) doRead(p []byte, useTermChar bool) (n int, err error) {
	d.bTag = nextbTag(d.bTag)
	header := encodeMsgInBulkOutHeader(d.bTag, uint32(len(p)),
		useTermChar && d.termCharEnabled, d.termChar)
	if _, err = d.usbDevice.Write(header[:]); err != nil {
		return 0, err
	}
	debug.Printf("sent reqdevdepmsgin hdr %v (data len %v)\n",
		hex.EncodeToString(header[:]), len(p))

	// Per Figure 4 in the USBTMC spec, messages may be sent in multiple
	// transfers. The first will have a USBTMC header, the middle transfers
	// will only contain data bytes, and the final may end with alignment
	// bytes. Mixed in with this are three definitions of length:
	//
	//   1) the number of bytes the caller wants to receive (len(p))
	//   2) the number of bytes the device means to send ('transfer', from
	//      the USBTMC header)
	//   3) the number of bytes in the current transfer (resp).
	//
	// The header also includes an end-of-message (EOM) bit, but it's not
	// clear how this bit is used.
	//
	// We'll attempt to read the number of bytes the caller wants (1), but
	// will stop short if the number of bytes the device wants to send (2)
	// is reached or if it sends a transfer with zero non-header bytes.
	pos := 0
	var transfer int
	for pos < len(p) {
		var resp int
		var err error
		if pos == 0 {
			resp, transfer, _, err = d.readRemoveHeader(p[pos:])
		} else {
			resp, err = d.readKeepHeader(p[pos:])
		}
		debug.Printf("read: pos %d (buf left %d); got %d bytes",
			pos, len(p[pos:]), resp)

		dumpLen, dumpTrunc := 100, 1
		if resp < dumpLen {
			dumpLen, dumpTrunc = resp, 0
		}
		if left := len(p) - pos; left < dumpLen {
			dumpLen, dumpTrunc = left, 0
		}
		debug.Printf("data[%d:]=%s%s\n", pos,
			hex.EncodeToString(p[pos:dumpLen]),
			[]string{"", "..."}[dumpTrunc])

		if err != nil {
			return pos, err
		}
		if resp == 0 {
			debug.Print("zero-length read; giving up")
			break
		}
		pos += resp
		if pos >= transfer {
			break
		}
	}

	return min(pos, transfer), nil
}

// Read reads from the device respecting the termChar setting. Use for transfers
// of ASCII data.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.doRead(p, true)
}

// BulkRead reads from the device without allowing termChar to be set. Use for
// transfers of binary data.
func (d *Device) BulkRead(p []byte) (n int, err error) {
	return d.doRead(p, false)
}

func inHdrToString(buf []byte) string {
	id, bTag, bTagInverse := msgID(buf[0]), buf[1], buf[2]

	out := "type "
	switch id {
	case devDepMsgOut:
		out += "1???" // no response expected
	case devDepMsgIn:
		out += "dvdp"
	case vendorSpecificOut:
		out += "126?" // no response expected
	case vendorSpecificIn:
		out += "vnsp"
	default:
		out += fmt.Sprintf("R%03d", id)
	}

	out += fmt.Sprintf(" tag % 3d", bTag)
	if invertbTag(bTag) != bTagInverse {
		out += fmt.Sprintf(" bad inv % 3d", bTagInverse)
	}

	if msgID(id) == devDepMsgIn {
		out += fmt.Sprintf(" sz %d", binary.LittleEndian.Uint32(buf[4:8]))

		attr := buf[8]
		out += fmt.Sprintf(" D1=%d", (attr&2)>>1)
		out += fmt.Sprintf(" EOM?=%s", []string{"no", "yes"}[(attr&1)])

		out += " " + hex.EncodeToString(buf[9:12])
	} else {
		out += " " + hex.EncodeToString(buf[4:12])
	}

	return out
}

func (d *Device) readRemoveHeader(p []byte) (n int, transfer int, transferAttr byte, err error) {
	// Reading from the USB device triggers interactions with the hardware,
	// so we take care with the buffer size. The caller expects len(p)
	// bytes, but we also need to allow space for the USBTMC header. The
	// libusb documentation is full of dire warnings about what happens if
	// the incoming data exceeds the receiving buffer[^1]. It recommends
	// making sure the incoming buffer is a multiple of the maximum packet
	// size. We don't know the actual maximum packet size, but we think we
	// know the maximum packet size, so rounding the transfer size up to the
	// next multiple of the maximum packet size should make it difficult for
	// incoming data to overflow.
	//
	// [^1]: https://libusb.sourceforge.io/api-1.0/libusb_packetoverflow.html
	tempSz := len(p) + usbtmcHeaderLen
	if m := tempSz % 512; m != 0 {
		tempSz += 512 - m
	}

	debug.Printf("readRemoveHeader: len(p) %v, w/hdr %v -> buf size %v\n",
		len(p), len(p)+usbtmcHeaderLen, tempSz)
	temp := make([]byte, tempSz)

	n, err = d.usbDevice.Read(temp)
	if err != nil {
		return 0, 0, 0, err
	}
	if n < usbtmcHeaderLen {
		return 0, 0, 0, fmt.Errorf(
			"short %d-byte read: no space for header", n)
	}

	debug.Printf("readRemoveHeader: header %s\n", inHdrToString(temp))

	t32 := binary.LittleEndian.Uint32(temp[4:8])
	transfer = int(t32)
	transferAttr = temp[8]

	// Copy the bytes after the reader to the caller's buffer, but only as
	// many bytes as the USB device said it read. Let the caller deal with
	// any discrepancies between the USBTMC transfer size and the number of
	// bytes we got from the USB device.
	toCopy := min(len(temp)-usbtmcHeaderLen, n-usbtmcHeaderLen)
	if toCopy > 0 {
		copy(p, temp[usbtmcHeaderLen:usbtmcHeaderLen+toCopy])
	}
	return n - usbtmcHeaderLen, transfer, transferAttr, nil
}

func (d *Device) readKeepHeader(p []byte) (n int, err error) {
	return d.usbDevice.Read(p)
}

// Close closes the underlying USB device.
func (d *Device) Close() error {
	return d.usbDevice.Close()
}

// WriteString writes a string using the underlying USB device. A newline
// terminator is not automatically added.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Command sends the SCPI/ASCII command to the underlying USB device. A newline
// character is automatically added to the end of the string.
func (d *Device) Command(format string, a ...interface{}) error {
	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	_, err := d.WriteString(strings.TrimSpace(cmd) + "\n")
	return err
}

// Query writes the given string to the USBTMC device and returns the returned
// value as a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(s string) (string, error) {
	err := d.Command(s)
	if err != nil {
		return "", err
	}

	// Try to ensure a single-packet read
	p := make([]byte, maxPacketSize-usbtmcHeaderLen)
	n, err := d.Read(p)
	if err != nil {
		return "", err
	}
	s = fmt.Sprintf("%s", p[:n])
	p = nil
	return s, nil
}
