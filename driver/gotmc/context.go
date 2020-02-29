// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package gotmc

import (
	"fmt"
	"log"

	"github.com/gotmc/libusb"
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
func (c *Context) NewDeviceByVIDPID(VID, PID uint) (driver.USBDevice, error) {
	dev, dh, err := c.ctx.OpenDeviceWithVendorProduct(uint16(VID), uint16(PID))
	if err != nil {
		return nil, err
	}
	usbDeviceDescriptor, _ := dev.GetDeviceDescriptor()
	if err != nil {
		return nil, err
	}
	defer dh.Close()
	configDescriptor, err := dev.GetActiveConfigDescriptor()
	if err != nil {
		return nil, fmt.Errorf("failed getting active config: %s", err)
	}
	log.Printf("Grabbed active config: %v", configDescriptor)
	firstDescriptor := configDescriptor.SupportedInterfaces[0].InterfaceDescriptors[0]
	err = dh.ClaimInterface(0)
	if err != nil {
		return nil, fmt.Errorf("error claiming USB interface: %s", err)
	}
	log.Println("Claimed interface 0")
	log.Printf("Found %d endpoint descriptors", len(firstDescriptor.EndpointDescriptors))
	// FIXME(mdr): Probably not a good idea to blindly assume these endpoints are
	// always in this order.
	bulkOutput := firstDescriptor.EndpointDescriptors[0]
	bulkInput := firstDescriptor.EndpointDescriptors[1]
	interruptEndpoint := firstDescriptor.EndpointDescriptors[2]

	d := Device{
		Timeout:           2000,
		USBDevice:         dev,
		DeviceDescriptor:  usbDeviceDescriptor,
		DeviceHandle:      dh,
		ConfigDescriptor:  configDescriptor,
		BulkInEndpoint:    bulkInput,
		BulkOutEndpoint:   bulkOutput,
		InterruptEndpoint: interruptEndpoint,
	}
	return &d, nil
}
