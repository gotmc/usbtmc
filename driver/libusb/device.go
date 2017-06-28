// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package libusb

import (
	"bytes"

	"github.com/gotmc/libusb"
)

type Device struct {
	Timeout           int
	USBDevice         *libusb.Device
	DeviceDescriptor  *libusb.DeviceDescriptor
	DeviceHandle      *libusb.DeviceHandle
	ConfigDescriptor  *libusb.ConfigDescriptor
	BulkInEndpoint    *libusb.EndpointDescriptor
	BulkOutEndpoint   *libusb.EndpointDescriptor
	InterruptEndpoint *libusb.EndpointDescriptor
}

// Close closes the Device.
func (d *Device) Close() error {
	return d.DeviceHandle.Close()
}

// String providers the Stringer interface method for Device.
func (d *Device) String() string {
	return "usbtmc/driver/libusb needs a better Stringer interface"
}

// Write writes to the USB device's bulk out endpoint.
func (d *Device) Write(p []byte) (n int, err error) {
	return d.DeviceHandle.BulkTransfer(
		d.BulkOutEndpoint.EndpointAddress,
		p,
		len(p),
		d.Timeout,
	)
}

// WriteString writes the given string to the Device and returns the number
// of bytes written along with an error code.
func (d *Device) WriteString(s string) (n int, err error) {
	return d.Write([]byte(s))
}

// Read reads from the USB device's bulk in endpoint.
func (d *Device) Read(p []byte) (n int, err error) {
	return d.DeviceHandle.BulkTransfer(
		d.BulkInEndpoint.EndpointAddress,
		p,
		len(p),
		d.Timeout,
	)
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
