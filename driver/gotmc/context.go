// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package gotmc

import (
	"fmt"
	"log"

	libusb "github.com/gotmc/libusb/v2"
	"github.com/gotmc/usbtmc"
	"github.com/gotmc/usbtmc/driver"
)

// Driver implements the visa.Driver interface required by usbtmc using the
// github.com/gotmc/libusb libusb driver.
type Driver struct {
}

func init() {
	usbtmc.Register(&Driver{})
}

// Context models libusb context and implements the driver.Context interface.
type Context struct {
	ctx *libusb.Context
}

// NewContext creates a new libusb session/context.
func (d Driver) NewContext() (driver.Context, error) {
	var c Context
	ctx, err := libusb.NewContext()
	c.ctx = ctx
	return &c, err
}

// SetDebugLevel sets the debug level for the libusb session/context
func (c *Context) SetDebugLevel(level int) {
	c.ctx.SetDebug(libusb.LogLevel(level))
}

// Close the libusb session/context.
func (c *Context) Close() error {
	return c.ctx.Close()
}

// NewDeviceByVIDPID creates new USB device based on the given the
// vendor ID and product ID. If multiple USB devices matching the VID and PID
// are found, only the first is returned.
func (c *Context) NewDeviceByVIDPID(VID, PID int) (driver.USBDevice, error) {
	dev, dh, err := c.ctx.OpenDeviceWithVendorProduct(uint16(VID), uint16(PID)) //nolint:gosec
	if err != nil {
		return nil, err
	}
	usbDeviceDescriptor, err := dev.DeviceDescriptor()
	if err != nil {
		_ = dh.Close()
		return nil, err
	}
	configDescriptor, err := dev.ActiveConfigDescriptor()
	if err != nil {
		_ = dh.Close()
		return nil, fmt.Errorf("failed getting active config: %w", err)
	}
	log.Printf("Grabbed active config: %v", configDescriptor)
	firstDescriptor := configDescriptor.SupportedInterfaces[0].InterfaceDescriptors[0]
	err = dh.ClaimInterface(0)
	if err != nil {
		_ = dh.Close()
		return nil, fmt.Errorf("error claiming USB interface: %w", err)
	}
	log.Println("Claimed interface 0")
	log.Printf("Found %d endpoint descriptors", len(firstDescriptor.EndpointDescriptors))
	var bulkIn, bulkOut, interruptIn *libusb.EndpointDescriptor
	for _, ep := range firstDescriptor.EndpointDescriptors {
		switch {
		case ep.Direction() == 0 && ep.TransferType() == libusb.BulkTransfer:
			bulkOut = ep
		case ep.Direction() == 1 && ep.TransferType() == libusb.BulkTransfer:
			bulkIn = ep
		case ep.Direction() == 1 && ep.TransferType() == libusb.InterruptTransfer:
			interruptIn = ep
		}
	}
	if bulkIn == nil || bulkOut == nil {
		_ = dh.Close()
		return nil, fmt.Errorf("missing required bulk endpoints on device")
	}

	d := Device{
		Timeout:           2000,
		USBDevice:         dev,
		DeviceDescriptor:  usbDeviceDescriptor,
		DeviceHandle:      dh,
		ConfigDescriptor:  configDescriptor,
		BulkInEndpoint:    bulkIn,
		BulkOutEndpoint:   bulkOut,
		InterruptEndpoint: interruptIn,
	}
	return &d, nil
}
