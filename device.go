// Copyright (c) 2015-2024 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/gotmc/usbtmc/driver"
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
func (d *Device) Read(p []byte) (n int, err error) {
	d.bTag = nextbTag(d.bTag)
	header := encodeMsgInBulkOutHeader(d.bTag, uint32(len(p)), d.termCharEnabled, d.termChar)
	if _, err = d.usbDevice.Write(header[:]); err != nil {
		return 0, err
	}
	pos := 0
	var transfer int
	for pos < len(p) {
		var resp int
		var err error
		if pos == 0 {
			resp, transfer, err = d.readRemoveHeader(p[pos:])
		} else {
			resp, err = d.readKeepHeader(p[pos:])
		}
		if err != nil {
			return pos, err
		}
		if resp == 0 {
			break
		}
		pos += resp
		if pos >= transfer {
			break
		}
	}
	return pos, nil
}

func (d *Device) readRemoveHeader(p []byte) (n int, transfer int, err error) {
	// FIXME(mdr): Seems like I shouldn't use 512 as a magic number or as a hard
	// size limit. I should grab the max size of the bulk in endpoint.
	usbtmcHeaderLen := 12
	temp := make([]byte, 512)
	n, err = d.usbDevice.Read(temp)
	// Remove the USBMTC Bulk-IN Header from the data and the number of bytes
	if n < usbtmcHeaderLen {
		return 0, 0, err
	}
	t32 := binary.LittleEndian.Uint32(temp[4:8])
	transfer = int(t32)
	reader := bytes.NewReader(temp)
	_, err = reader.ReadAt(p, int64(usbtmcHeaderLen))

	if err != nil && err != io.EOF {
		return n - usbtmcHeaderLen, transfer, err
	}
	return n - usbtmcHeaderLen, transfer, nil
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
	p := make([]byte, 512)
	n, err := d.Read(p)
	if err != nil {
		return "", err
	}
	s = fmt.Sprintf("%s", p[:n])
	p = nil
	return s, nil
}
