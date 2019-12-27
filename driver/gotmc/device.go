// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package libusb

import "github.com/gotmc/libusb"

// Device models the libusb device that will form the basis of the USBTMC
// compliant device.
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
