// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package gotmc

import (
	"context"
	"time"

	libusb "github.com/gotmc/libusb/v2"
)

// Device models the libusb device that will form the basis of the USBTMC
// compliant device.
type Device struct {
	Timeout           int
	USBDevice         *libusb.Device
	DeviceDescriptor  *libusb.Descriptor
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

// ReadContext reads from the USB device's bulk in endpoint in a context aware
// manner. If the context has a deadline, it is converted to a libusb timeout
// in milliseconds; otherwise the device's default timeout is used.
func (d *Device) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return d.DeviceHandle.BulkTransfer(
		d.BulkInEndpoint.EndpointAddress,
		p,
		len(p),
		d.contextTimeout(ctx),
	)
}

// WriteContext writes to the USB device's bulk out endpoint in a context aware
// manner. If the context has a deadline, it is converted to a libusb timeout
// in milliseconds; otherwise the device's default timeout is used.
func (d *Device) WriteContext(ctx context.Context, p []byte) (n int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return d.DeviceHandle.BulkTransfer(
		d.BulkOutEndpoint.EndpointAddress,
		p,
		len(p),
		d.contextTimeout(ctx),
	)
}

// contextTimeout returns a libusb timeout in milliseconds derived from the
// context's deadline. If no deadline is set, the device's default Timeout is
// returned.
func (d *Device) contextTimeout(ctx context.Context) int {
	if deadline, ok := ctx.Deadline(); ok {
		ms := time.Until(deadline).Milliseconds()
		if ms <= 0 {
			return 1 // minimum timeout to avoid blocking indefinitely
		}
		return int(ms)
	}
	return d.Timeout
}
