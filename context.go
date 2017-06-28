// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "github.com/gotmc/usbtmc/driver"

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
func (c *Context) NewDeviceByVIDPID(VID, PID uint) (*Device, error) {
	var d Device
	d.termChar = '\n'
	d.bTag = 1
	d.termCharEnabled = true
	usbDevice, err := c.libusbContext.NewDeviceByVIDPID(VID, PID)
	if err != nil {
		return &d, err
	}
	d.usbDevice = usbDevice
	return &d, nil
}

// Close closes the USB context for the underlying USB driver.
func (c *Context) Close() error {
	return c.libusbContext.Close()
}
