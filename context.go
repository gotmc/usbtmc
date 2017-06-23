// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"fmt"
	"log"
	"time"

	"github.com/truveris/gousb/usb"
)

// Context represents a libusb session/context.
type Context struct {
	ctx *usb.Context
}

// NewContext creates a new libusb session/context.
func NewContext() *Context {
	c := &Context{
		ctx: usb.NewContext(),
	}
	return c
}

// Debug sets the debug level for the libusb session/context
func (c *Context) Debug(level int) {
	c.ctx.Debug(level)
}

// Close the libusb session/context.
func (c *Context) Close() error {
	return c.ctx.Close()
}

// NewDevice creates new USBTMC compliant device based on the given VISA
// resource name.
func (c *Context) NewDevice(visaResourceName string) (*Device, error) {
	var usbtmcConfig uint8
	var usbtmcInterface uint8
	var usbtmcSetup uint8
	var bulkOutEndpointAddress uint8
	var bulkInEndpointAddress uint8
	var interruptInEndpointAddress uint8
	start := time.Now()
	v, err := NewVisaResource(visaResourceName)
	if err != nil {
		return nil, err
	}
	devices, err := c.ctx.ListDevices(FindUsbtmcFromResource(v))
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Didn't find usbtmc device for %s", visaResourceName)
	}
	device := devices[0]
	log.Printf("%.2fs to get first USB device matching VisaResource\n", time.Since(start).Seconds())
	start = time.Now()
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
		log.Fatal("Error opening bulkInEndpoint")
	}

	bulkOutEndpoint, err := device.OpenEndpoint(
		usbtmcConfig, usbtmcInterface, usbtmcSetup, bulkOutEndpointAddress)
	if err != nil {
		log.Fatal("Error opening bulkOutEndpoint")
	}

	// TODO(mdr): Need to make the interruptInEndpoint optional
	interruptInEndpoint, err := device.OpenEndpoint(
		usbtmcConfig, usbtmcInterface, usbtmcSetup, interruptInEndpointAddress)
	if err != nil {
		log.Fatal("Error opening interruptInEndpoint")
	}

	// TODO(mdr): Should I set the bTag to 1? Instead of storing bTag, should I
	// store nextbTag, or maybe renamed this to lastbTag?
	d := Device{
		USBDevice:           device,
		BulkInEndpoint:      bulkInEndpoint,
		BulkOutEndpoint:     bulkOutEndpoint,
		InterruptInEndpoint: interruptInEndpoint,
		termChar:            '\n',
		termCharEnabled:     true,
	}
	return &d, nil
}
