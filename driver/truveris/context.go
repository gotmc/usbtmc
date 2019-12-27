// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package truveris

import (
	"fmt"

	"github.com/gotmc/usbtmc"
	"github.com/gotmc/usbtmc/driver"
	"github.com/truveris/gousb/usb"
)

// Driver implements the visa.Driver interface.
type Driver struct {
}

func init() {
	usbtmc.Register(&Driver{})
}

// Context models libusb context and implements the driver.Context interface.
type Context struct {
	ctx *usb.Context
}

// NewContext creates a new libusb session/context.
func (d Driver) NewContext() (driver.Context, error) {
	var c Context
	c.ctx = usb.NewContext()
	return &c, nil
}

// SetDebugLevel sets the debug level for the libusb session/context
func (c *Context) SetDebugLevel(level int) {
	c.ctx.Debug(level)
}

// Close the libusb session/context.
func (c *Context) Close() error {
	return c.ctx.Close()
}

// NewDeviceByVIDPID creates new USB device based on the given the
// vendor ID and product ID. If multiple USB devices matching the VID and PID
// are found, only the first is returned.
func (c *Context) NewDeviceByVIDPID(VID, PID uint) (driver.USBDevice, error) {
	var usbtmcConfig uint8
	var usbtmcInterface uint8
	var usbtmcSetup uint8
	var bulkOutEndpointAddress uint8
	var bulkInEndpointAddress uint8
	var interruptInEndpointAddress uint8
	devices, err := c.ctx.ListDevices(findDeviceByVIDPID(uint16(VID), uint16(PID)))
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Didn't find USB device VID %v PID %v", VID, PID)
	}
	device := devices[0]
	for _, config := range device.Descriptor.Configs {
		for _, iface := range config.Interfaces {
			for _, setup := range iface.Setups {
				if setup.IfClass == 0xfe && setup.IfSubClass == 0x03 {
					usbtmcConfig = uint8(config.Config)
					usbtmcInterface = uint8(iface.Number)
					usbtmcSetup = uint8(setup.Number)
					for _, endpoint := range setup.Endpoints {
						endpointAttributes := endpoint.Attributes
						endpointDirection := endpoint.Address & uint8(usb.ENDPOINT_DIR_MASK)
						endpointType := endpointAttributes & uint8(usb.TRANSFER_TYPE_MASK)
						if endpointType == uint8(usb.TRANSFER_TYPE_BULK) {
							if endpointDirection == uint8(usb.ENDPOINT_DIR_IN) {
								bulkInEndpointAddress = endpoint.Address | uint8(usb.ENDPOINT_DIR_IN)
							} else if endpointDirection == uint8(usb.ENDPOINT_DIR_OUT) {
								bulkOutEndpointAddress = endpoint.Address | uint8(usb.ENDPOINT_DIR_OUT)
							}
						} else if endpointType == uint8(usb.TRANSFER_TYPE_INTERRUPT) {
							if endpointDirection == uint8(usb.ENDPOINT_DIR_IN) {
								interruptInEndpointAddress = endpoint.Address | uint8(usb.ENDPOINT_DIR_IN)
							}
						}
					}
				}
			}
		}
	}

	bulkInEndpoint, err := device.OpenEndpoint(
		usbtmcConfig, usbtmcInterface, usbtmcSetup, bulkInEndpointAddress)
	if err != nil {
		return nil, fmt.Errorf("error opening bulkInEndpoint: %s", err)
	}

	bulkOutEndpoint, err := device.OpenEndpoint(
		usbtmcConfig, usbtmcInterface, usbtmcSetup, bulkOutEndpointAddress)
	if err != nil {
		return nil, fmt.Errorf("Error opening bulkOutEndpoint: %s", err)
	}

	// TODO(mdr): Need to make the interruptInEndpoint optional
	interruptInEndpoint, err := device.OpenEndpoint(
		usbtmcConfig, usbtmcInterface, usbtmcSetup, interruptInEndpointAddress)
	if err != nil {
		return nil, fmt.Errorf("Error opening interruptInEndpoint: %s", err)
	}

	d := Device{
		USBDevice:           device,
		BulkInEndpoint:      bulkInEndpoint,
		BulkOutEndpoint:     bulkOutEndpoint,
		InterruptInEndpoint: interruptInEndpoint,
	}
	return &d, nil
}
