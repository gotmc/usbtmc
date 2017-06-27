// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "github.com/gotmc/usbtmc/driver"

type Context struct {
	libusbContext *driver.Context
}

func NewContext() (*Context, error) {
	var context *Context
	ctx, err := libusbDriver.NewContext()
	if err != nil {
		return context, err
	}
	context.libusbContext = ctx
	return context, nil
}

// NewDeviceByVIDPID creates new USBTMC compliant device based on the given the
// vendor ID and product ID. If multiple USB devices matching the VID and PID
// are found, only the first is returned.
func (c *Context) NewDeviceByVIDPID(VID, PID uint) (*Device, error) {
	var d *Device
	d.termChar = '\n'
	d.termCharEnabled = true
	usbDevice, err := c.libusbContext.NewDeviceByVIDPID(VID, PID)
	if err != nil {
		return d, err
	}
	d.usbDevice = usbDevice
	return d, nil
}

func (c *Context) Close() error {
	return c.libusbContext.Close()
}
