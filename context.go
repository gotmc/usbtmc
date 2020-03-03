// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"github.com/gotmc/usbtmc/driver"
)

// Context hold the USB context for the registered driver.
type Context struct {
	driver        driver.Driver
	libusbContext driver.Context
}

// NewContext creates a new USB context using the registered driver.
func NewContext() (*Context, error) {
	var context Context
	context.driver = libusbDriver
	ctx, err := libusbDriver.NewContext()
	if err != nil {
		return nil, err
	}
	context.libusbContext = ctx
	return &context, nil
}

// NewDeviceByVIDPID creates new USBTMC compliant device based on the given the
// vendor ID and product ID. If multiple USB devices matching the VID and PID
// are found, only the first is returned.
func (c *Context) NewDeviceByVIDPID(VID, PID int) (*Device, error) {
	d := defaultDevice()
	usbDevice, err := c.libusbContext.NewDeviceByVIDPID(VID, PID)
	if err != nil {
		return nil, err
	}
	d.usbDevice = usbDevice
	return &d, nil
}

// NewDevice creates a new USBTMC compliant device based on the given VISA
// address string.
func (c *Context) NewDevice(address string) (*Device, error) {
	v, err := NewVisaResource(address)
	if err != nil {
		return nil, err
	}
	return c.NewDeviceByVIDPID(v.manufacturerID, v.modelCode)
}

func defaultDevice() Device {
	return Device{
		termChar:        '\n',
		bTag:            1,
		termCharEnabled: true,
	}
}

// Close closes the USB context for the underlying USB driver.
func (c *Context) Close() error {
	return c.libusbContext.Close()
}

// SetDebugLevel sets the debug level for the underlying USB device using the
// given integer.
func (c *Context) SetDebugLevel(level int) {
	c.libusbContext.SetDebugLevel(level)
}
