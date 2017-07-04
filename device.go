// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"fmt"
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
	// log.Println("device.go/Write()")
	d.bTag = nextbTag(d.bTag)
	header := encodeBulkOutHeader(d.bTag, uint32(len(p)), true)
	data := append(header[:], p...)
	if moduloFour := len(data) % 4; moduloFour > 0 {
		numAlignment := 4 - moduloFour
		alignment := bytes.Repeat([]byte{0x00}, numAlignment)
		data = append(data, alignment...)
	}
	// log.Println("About to run d.usbDevice.Write(data)")
	return d.usbDevice.Write(data)
}

// Read creates and sends the header on the bulk out endpoint and then reads
// from the bulk in endpoint per USBTMC standard.
func (d *Device) Read(p []byte) (n int, err error) {
	log.Println("device.go/Read()")
	return d.readKeepHeader(p)
}

func (d *Device) readRemoveHeader(p []byte) (n int, err error) {
	// FIXME(mdr): Seems like I shouldn't use 1024 as a magic number or as a hard
	// size limit.
	log.Println("In device.go/Read()")
	usbtmcHeaderLen := 12
	temp := make([]byte, 512)
	d.bTag = nextbTag(d.bTag)
	header := encodeMsgInBulkOutHeader(d.bTag, uint32(len(p)), d.termCharEnabled, d.termChar)
	n, err = d.usbDevice.Write(header[:])
	log.Println("Wrote the header.")
	n, err = d.usbDevice.Read(temp)
	log.Println("Read from USB device.")
	// Remove the USBMTC Bulk-IN Header from the data and the number of bytes
	if n < usbtmcHeaderLen {
		return 0, err
	}
	reader := bytes.NewReader(temp)
	_, err = reader.ReadAt(p, int64(usbtmcHeaderLen))
	if err != nil && err != io.EOF {
		return n - usbtmcHeaderLen, err
	}
	return n - usbtmcHeaderLen, nil
}

func (d *Device) readKeepHeader(p []byte) (n int, err error) {
	d.bTag = nextbTag(d.bTag)
	header := encodeMsgInBulkOutHeader(d.bTag, uint32(len(p)), d.termCharEnabled, d.termChar)
	log.Println("device.go/readKeepHeader()")
	_, err = d.usbDevice.Write(header[:])
	log.Println("Wrote header in device.go/readKeepHeader()")
	return d.usbDevice.Read(p)
}

// Close closes the underlying USB device.
func (d *Device) Close() error {
	return d.usbDevice.Close()
}

// WriteString writes a string using the underlying USB device.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Query writes the given string to the USBTMC device and returns the returned
// value as a string.
func (d *Device) Query(s string) (string, error) {
	// log.Println("Inside device.go/Query(s string)")
	_, err := d.WriteString(s)
	if err != nil {
		return "", err
	}
	p := make([]byte, 512)
	log.Println("Here")
	n, err := d.Read(p)
	if err != nil {
		return "", err
	}
	s = fmt.Sprintf("%s", p[:n])
	p = nil
	return s, nil
}
