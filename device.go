// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"io"
	"log"

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
	d.bTag = nextbTag(d.bTag)
	header := createDevDepMsgOutBulkOutHeader(d.bTag, uint32(len(p)), true)
	data := append(header[:], p...)
	if moduloFour := len(data) % 4; moduloFour > 0 {
		numAlignment := 4 - moduloFour
		alignment := bytes.Repeat([]byte{0x00}, numAlignment)
		data = append(data, alignment...)
	}
	return d.usbDevice.Write(data)
}

// Read creates and sends the header on the bulk out endpoint and then reads
// from the bulk in endpoint per USBTMC standard.
func (d *Device) Read(p []byte) (n int, err error) {
	// FIXME(mdr): Seems like I shouldn't use 1024 as a magic number or as a hard
	// size limit.
	temp := make([]byte, 1024)
	d.bTag = nextbTag(d.bTag)
	header := createRequestDevDepMsgInBulkOutHeader(d.bTag, uint32(len(p)), d.termCharEnabled, d.termChar)
	// log.Printf("RequestDevDepMsg Header to write = %v", header)
	n, err = d.usbDevice.Write(header[:])
	n, err = d.usbDevice.Read(temp)
	log.Println("Foo")
	// log.Printf("Read %d bytes on BulkInEndpoint", n)
	// Remove the USBMTC Bulk-IN Header from the data and the number of bytes
	if n < 12 {
		return 0, err
	}
	reader := bytes.NewReader(temp)
	n, err = reader.ReadAt(p, 12)
	if err != nil && err != io.EOF {
		return n, err
	}
	return n, nil
}

// Close closes the underlying USB device.
func (d *Device) Close() error {
	return d.usbDevice.Close()
}

// WriteString writes a string using the underlying USB device.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}
