// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package google

import (
	"log"

	"github.com/google/gousb"
	"github.com/gotmc/usbtmc"
	"github.com/gotmc/usbtmc/driver"
)

// Driver implements the visa.Driver interface.
type Driver struct {
}

func init() {
	usbtmc.Register(&Driver{})
}

// Context models libusb context and implements the driver.Context interface.
type Context struct {
	ctx *gousb.Context
}

// NewContext creates a new libusb session/context.
func (d Driver) NewContext() (driver.Context, error) {
	var c Context
	c.ctx = gousb.NewContext()
	return &c, nil
}

// Debug sets the debug level for the libusb session/context
func (c *Context) Debug(level int) {
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

	// Iterate through available Devices, finding all that match a known VID/PID.
	vid, pid := gousb.ID(VID), gousb.ID(PID)
	devs, err := c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// this function is called for every device present.
		// Returning true means the device should be opened.
		return desc.Vendor == vid && desc.Product == pid
	})
	log.Printf("Found %d devices", len(devs))
	// All returned devices are now open and will need to be closed.
	for i, d := range devs {
		if i != 0 {
			defer d.Close()
		}
	}
	if err != nil {
		log.Fatalf("OpenDevices(): %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("no devices found matching VID %s and PID %s", vid, pid)
	}

	log.Printf("Found %d USB devices.", len(devs))
	// Pick the first device found.
	dev := devs[0]

	log.Printf("Device Vendor ID = %s, Product ID = %s",
		dev.Desc.Vendor,
		dev.Desc.Product,
	)
	log.Printf("Found %s device", dev)

	log.Printf("Device class %x, subclass %x", dev.Desc.Class, dev.Desc.SubClass)
	// Switch to configuration #0
	activeConfig, err := dev.ActiveConfigNum()
	if err != nil {
		return nil, err
	}
	cfg, err := dev.Config(activeConfig)
	if err != nil {
		return nil, err
	}
	var bulkIn *gousb.InEndpoint
	var bulkOut *gousb.OutEndpoint
	var intIn *gousb.InEndpoint
	var intf *gousb.Interface
	log.Printf("Found %d interfaces", len(cfg.Desc.Interfaces))
	// Loop through the interfaces
	for _, interfaceDesc := range cfg.Desc.Interfaces {
		log.Printf("Found %d interfaces", len(cfg.Desc.Interfaces))
		intf, err := cfg.Interface(interfaceDesc.Number, 0)
		if err != nil {
			log.Printf("err: %s")
		}
		// Loop through all the endpoints on this interface
		for j, ep := range intf.Setting.Endpoints {
			log.Printf("Endpoint idx %d = %s", j, ep)
			if ep.Direction == gousb.EndpointDirectionOut && ep.TransferType == gousb.TransferTypeBulk {
				log.Printf("Found BulkOut endpoint #%d on interface %d", ep.Number, interfaceDesc.Number)
				bulkOut, err = intf.OutEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
			if ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeBulk {
				log.Printf("Found BulkOut endpoint #%d on interface %d", ep.Number, interfaceDesc.Number)
				bulkIn, err = intf.InEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
			if ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeInterrupt {
				log.Printf("Found BulkOut endpoint #%d on interface %d", ep.Number, interfaceDesc.Number)
				intIn, err = intf.InEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	d := Device{
		dev:                 dev,
		intf:                intf,
		cfg:                 cfg,
		BulkInEndpoint:      bulkIn,
		BulkOutEndpoint:     bulkOut,
		InterruptInEndpoint: intIn,
	}
	return &d, nil
}
