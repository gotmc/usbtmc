// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"log"

	"github.com/gotmc/usbtmc/driver"
)

type Device struct {
	usbDevice       driver.USBDevice
	bTag            byte
	termChar        byte
	termCharEnabled bool
}

func (d *Device) nextbTag() {
	d.bTag = (d.bTag % 255) + 1
}

// Write creates the appropriate USBMTC header, writes the header and data on
// the bulk out endpoint, and returns the number of bytes written and any
// errors.
func (d *Device) Write(p []byte) (n int, err error) {
	// FIXME(mdr): I need to change this so that I look at the size of the buf
	// being written to see if it can truly fit into one transfer, and if not
	// split it into multiple transfers.
	d.nextbTag()
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
	d.nextbTag()
	header := createRequestDevDepMsgInBulkOutHeader(d.bTag, uint32(len(p)), d.termCharEnabled, d.termChar)
	log.Printf("RequestDevDepMsg Header to write = %v", header)
	n, err = d.usbDevice.Write(header[:])
	n, err = d.usbDevice.Read(p)
	log.Printf("Read %d bytes on BulkInEndpoint", n)
	return n, err
}

// Close closes the underlying USB device.
func (d *Device) Close() error {
	return d.usbDevice.Close()
}

func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}
