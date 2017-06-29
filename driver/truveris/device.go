// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package truveris

import (
	"bytes"

	"github.com/truveris/gousb/usb"
)

// Device represents a USB device not a USBMTC device.
type Device struct {
	USBDevice           *usb.Device
	BulkInEndpoint      usb.Endpoint
	BulkOutEndpoint     usb.Endpoint
	InterruptInEndpoint usb.Endpoint
}

// Close closes the Device.
func (d *Device) Close() error {
	return d.USBDevice.Close()
}

// String providers the Stringer interface method for Device.
func (d *Device) String() string {
	return d.USBDevice.Descriptor.SerialNumber
}

// Write writes to the USB device's bulk out endpoint.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.BulkOutEndpoint.Write(p)
}

// WriteString writes the given string to the Device and returns the number
// of bytes written along with an error code.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Read reads from the USB device's bulk in endpoint.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.BulkInEndpoint.Read(p)
}

// Query writes a SCPI command as a string and then returns the queried result
// as a string.
func (d *Device) Query(s string) (string, error) {
	_, err := d.WriteString(s)
	if err != nil {
		return "", err
	}
	var b []byte
	_, err = d.Read(b)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(b)
	return buf.ReadString(0xA)
}