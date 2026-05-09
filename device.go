// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gotmc/usbtmc/driver"
)

const (
	// This is a guess. The USB spec says the max value can be fetched from
	// the descriptor, but the libusb documentation says packets can be up
	// to 512 bytes.
	// Ref: https://libusb.sourceforge.io/api-1.0/libusb_packetoverflow.html
	maxPacketSize = 512

	usbtmcHeaderLen = 12

	// trailingDrainTimeout caps how long doRead waits when draining bulk-IN
	// packets queued past the device-declared transfer length. Devices that
	// terminate with a zero-length packet return immediately; this timeout
	// is only reached by devices that go silent rather than sending a ZLP.
	trailingDrainTimeout = 20 * time.Millisecond
)

// Sentinel errors returned by readRemoveHeader for response framing problems.
// doRead matches these via errors.Is so it can recover from continuation
// reads on firmwares that echo a stale bTag rather than advancing it.
var (
	errMsgIDMismatch       = errors.New("unexpected MsgID")
	errBTagMismatch        = errors.New("bTag mismatch")
	errBTagInverseMismatch = errors.New("bTagInverse mismatch")
)

func isContinuationHeaderMismatch(err error) bool {
	return errors.Is(err, errMsgIDMismatch) ||
		errors.Is(err, errBTagMismatch) ||
		errors.Is(err, errBTagInverseMismatch)
}

// Device models a USBTMC device, which includes a USB device and the required
// USBTMC attributes and methods.
type Device struct {
	mu              sync.Mutex
	usbDevice       driver.USBDevice
	bTag            byte
	termChar        byte
	termCharEnabled bool
}

// Write creates the appropriate USBMTC header, writes the header and data on
// the bulk out endpoint, and returns the number of bytes written and any
// errors.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.WriteBinary(context.Background(), p)
}

// WriteBinary writes binary data without adding a terminator. It creates the
// appropriate USBTMC header, writes the header and data on the bulk out
// endpoint, and returns the number of bytes written and any errors.
func (d *Device) WriteBinary(ctx context.Context, p []byte) (n int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// FIXME(mdr): I need to change this so that I look at the size of the buf
	// being written to see if it can truly fit into one transfer, and if not
	// split it into multiple transfers.
	maxTransferSize := 512
	for pos := 0; pos < len(p); {
		if err := ctx.Err(); err != nil {
			return pos, err
		}
		d.bTag = nextbTag(d.bTag)
		thisLen := len(p[pos:])
		if thisLen > maxTransferSize-bulkOutHeaderSize {
			thisLen = maxTransferSize - bulkOutHeaderSize
		}
		isLastChunk := pos+thisLen >= len(p)
		header := encodeBulkOutHeader(d.bTag, uint32(thisLen), isLastChunk)
		data := append(header[:], p[pos:pos+thisLen]...)
		if moduloFour := len(data) % 4; moduloFour > 0 {
			numAlignment := 4 - moduloFour
			alignment := bytes.Repeat([]byte{0x00}, numAlignment)
			data = append(data, alignment...)
		}
		_, err := d.usbDevice.WriteContext(ctx, data)
		if err != nil {
			return pos, err
		}
		pos += thisLen
	}
	return len(p), nil
}

// doRead sends REQUEST_DEV_DEP_MSG_IN on the bulk-out endpoint and reads the
// response from the bulk-in endpoint per the USBTMC standard.
//
// USBTMC §3.3.1 allows a logical message-in transaction to span multiple
// DEV_DEP_MSG_IN responses, each driven by its own REQUEST. The outer loop
// continues until EOM=1, the caller's buffer is full, or the trailing
// drain returns zero bytes (ZLP or timeout).
func (d *Device) doRead(ctx context.Context, p []byte, useTermChar bool) (n int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	pos := 0
	for initial := true; ; initial = false {
		if err := ctx.Err(); err != nil {
			return pos, err
		}
		d.bTag = nextbTag(d.bTag)
		header := encodeMsgInBulkOutHeader(d.bTag, uint32(len(p)-pos), //nolint:gosec
			useTermChar && d.termCharEnabled, d.termChar)
		if _, err = d.usbDevice.WriteContext(ctx, header[:]); err != nil {
			return pos, err
		}
		debug.Printf("sent reqdevdepmsgin hdr %v (data len %v)\n",
			hex.EncodeToString(header[:]), len(p)-pos)

		msgStart := pos
		resp, transfer, transferAttr, rerr := d.readRemoveHeader(ctx, d.bTag, p[pos:])
		if rerr != nil {
			// Some firmwares answer a continuation REQUEST with a
			// bookkeeping packet that echoes the previous bTag and queue
			// the remaining payload as raw bulk-IN packets without a
			// USBTMC header. Tolerate the mismatch on continuation reads
			// by draining what's queued and returning what we have.
			if !initial && isContinuationHeaderMismatch(rerr) {
				debug.Printf("continuation header mismatch (tolerated): %v", rerr)
				pos += d.drainBulkIn(ctx, p[pos:])
				return pos, nil
			}
			return pos, rerr
		}
		pos += resp

		// Per Figure 4 of the USBTMC spec, subsequent transfers carry only
		// data plus alignment padding. Read until the device-declared
		// transfer is reached, the caller's buffer is full, or a read
		// returns zero bytes.
		for pos-msgStart < transfer && pos < len(p) {
			if err := ctx.Err(); err != nil {
				return pos, err
			}
			r, rerr := d.readKeepHeader(ctx, p[pos:])
			debug.Printf("read: pos %d (buf left %d); got %d bytes",
				pos, len(p[pos:]), r)
			dumpLen, dumpTrunc := 100, 1
			if r < dumpLen {
				dumpLen, dumpTrunc = r, 0
			}
			if left := len(p) - pos; left < dumpLen {
				dumpLen, dumpTrunc = left, 0
			}
			debug.Printf("data[%d:]=%s%s\n", pos,
				hex.EncodeToString(p[pos:pos+dumpLen]),
				[]string{"", "..."}[dumpTrunc])
			if rerr != nil {
				return pos, rerr
			}
			if r == 0 {
				debug.Print("zero-length read; giving up")
				break
			}
			pos += r
		}

		// On the first iteration, drain any bulk-IN packets queued past the
		// declared transfer length. Some devices (e.g. Rigol DS1000Z series)
		// advertise a smaller transfer_size than the actual payload and never
		// set EOM=1, relying on a ZLP to signal end-of-message. The timeout
		// is a safety net for devices that go silent instead of sending a ZLP.
		var trailing int
		if initial && pos < len(p) {
			drainCtx, cancel := context.WithTimeout(ctx, trailingDrainTimeout)
			trailing = d.drainBulkIn(drainCtx, p[pos:])
			cancel()
			pos += trailing
		}
		if trailing == 0 && pos-msgStart > transfer {
			pos = msgStart + transfer
		}
		eom := transferAttr&0x01 != 0
		if eom || pos >= len(p) || pos == msgStart || trailing > 0 {
			break
		}
	}
	return pos, nil
}

// drainBulkIn reads raw bulk-IN packets into p until the buffer is full, a
// zero-length read is returned, an error occurs, or ctx is cancelled.
func (d *Device) drainBulkIn(ctx context.Context, p []byte) int {
	n := 0
	for n < len(p) {
		if ctx.Err() != nil {
			break
		}
		r, err := d.readKeepHeader(ctx, p[n:])
		if err != nil || r == 0 {
			break
		}
		n += r
	}
	return n
}

// Read reads from the device respecting the termChar setting. Use for transfers
// of ASCII data.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.doRead(context.Background(), p, true)
}

// ReadBinary reads binary data without terminator interpretation.
func (d *Device) ReadBinary(ctx context.Context, p []byte) (n int, err error) {
	return d.doRead(ctx, p, false)
}

// ReadRaw reads from the device without allowing termChar to be set. Use for
// transfers of binary data.
func (d *Device) ReadRaw(p []byte) (n int, err error) {
	return d.ReadBinary(context.Background(), p)
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

func (d *Device) readRemoveHeader(
	ctx context.Context, expectedBTag byte, p []byte,
) (n int, transfer int, transferAttr byte, err error) {
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

	n, err = d.usbDevice.ReadContext(ctx, temp)
	if err != nil {
		return 0, 0, 0, err
	}
	if n < usbtmcHeaderLen {
		return 0, 0, 0, fmt.Errorf(
			"short %d-byte read: no space for header", n)
	}

	debug.Printf("readRemoveHeader: header %s\n", inHdrToString(temp))

	// Validate the response header per USBTMC Table 5.
	respMsgID := msgID(temp[0])
	if respMsgID != devDepMsgIn {
		return 0, 0, 0, fmt.Errorf(
			"%w: got %d, want %d (DEV_DEP_MSG_IN)",
			errMsgIDMismatch, respMsgID, devDepMsgIn)
	}
	respBTag := temp[1]
	if respBTag != expectedBTag {
		return 0, 0, 0, fmt.Errorf(
			"%w: got %d, want %d", errBTagMismatch, respBTag, expectedBTag)
	}
	if temp[2] != invertbTag(respBTag) {
		return 0, 0, 0, fmt.Errorf(
			"%w: got %d, want %d",
			errBTagInverseMismatch, temp[2], invertbTag(respBTag))
	}

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

func (d *Device) readKeepHeader(ctx context.Context, p []byte) (n int, err error) {
	return d.usbDevice.ReadContext(ctx, p)
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

// WriteStringContext is like WriteString but accepts a context.
func (d *Device) WriteStringContext(ctx context.Context, s string) (n int, err error) {
	return d.WriteBinary(ctx, []byte(s))
}

// Command sends the SCPI/ASCII command to the underlying USB device. A newline
// character is automatically added to the end of the string.
func (d *Device) Command(ctx context.Context, format string, a ...any) error {
	cmd := format
	if a != nil {
		cmd = fmt.Sprintf(format, a...)
	}
	_, err := d.WriteStringContext(ctx, strings.TrimSpace(cmd)+string(d.termChar))
	return err
}

// Query writes the given string to the USBTMC device and returns the returned
// value as a string. A newline character is automatically added to the query
// command sent to the instrument.
func (d *Device) Query(ctx context.Context, s string) (string, error) {
	err := d.Command(ctx, s)
	if err != nil {
		return "", err
	}

	// Try to ensure a single-packet read using ASCII mode (with termChar).
	p := make([]byte, maxPacketSize-usbtmcHeaderLen)
	n, err := d.doRead(ctx, p, true)
	if err != nil {
		return "", err
	}
	s = string(p[:n])
	p = nil
	return s, nil
}
