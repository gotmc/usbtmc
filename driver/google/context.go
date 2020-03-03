// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package google

import (
	"fmt"
	"log"
	"time"

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

// SetDebugLevel sets the debug level for the libusb session/context
func (c *Context) SetDebugLevel(level int) {
	c.ctx.Debug(level)
}

// Close the libusb session/context.
func (c *Context) Close() error {
	return c.ctx.Close()
}

// NewDeviceByVIDPID creates new USB device based on the given the vendor ID
// and product ID. If multiple USB devices matching the VID and PID are found,
// only the first is returned.
func (c *Context) NewDeviceByVIDPID(VID, PID int) (driver.USBDevice, error) {
	// Iterate through available devices. Find all devices that match the given
	// Vendor ID and Product ID.
	vid, usbtmcPID := gousb.ID(VID), gousb.ID(PID)
	devs, err := c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// This anonymous function is called for every device present. Returning
		// true means the device should be opened.
		return desc.Vendor == vid && desc.Product == usbtmcPID
	})
	if err != nil {
		// Close all devices and return error.
		for _, d := range devs {
			// I'm ignoring any errors on close at the moment.
			d.Close()
		}
		return nil, err
	}
	if len(devs) == 0 && VID == 0x0957 {
		// Didn't find a device matching the given vendor ID and product ID. Need
		// to make sure we're not trying to find one of the Agilent/Keysight (VID
		// 2391 = 0x0957) USB modular test equipment that powers up in a firmware
		// update mode. In order to get out of boot mode and into the normal USBTMC
		// mode, some control commands need to be sent. Once in normal USBTMC mode,
		// the Product ID will change.
		bootPIDs := map[gousb.ID]gousb.ID{
			// usbtmcPID: bootPID
			0x2818: 0x2918, // U2702A 200 MHz Oscilloscope
			0x3D18: 0x3E18, // U2751A 4x8 2-wire Switch Matrix
			0x4118: 0x4218, // U2722A Source Measure Unit
			0x4318: 0x4418, // U2723A Source Measure Unit
		}
		if bootPID, ok := bootPIDs[usbtmcPID]; ok {
			// Iterate through available USB devices. Find all devices that match the
			// Keysight USB modular boot PID.
			devs, err = c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
				return desc.Vendor == vid && desc.Product == bootPID
			})
			if err != nil {
				// Close all devices and return error.
				for _, d := range devs {
					// I'm ignoring any errors on close at the moment.
					d.Close()
				}
				return nil, err
			}
			if len(devs) == 0 {
				return nil, fmt.Errorf("no devices found matching VID %s and PID %s", vid, usbtmcPID)
			}
			// Found a Keysight USB modular device, so exit boot mode.
			err = exitBootMode(devs[0], bootPID)

			// Now find the normal USBTMC mode.
			log.Printf("Look for vendor ID %s and product ID %s", vid, usbtmcPID)
			devs, err = c.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
				// This anonymous function is called for every device present. Returning
				// true means the device should be opened.
				return desc.Vendor == vid && desc.Product == usbtmcPID
			})
			log.Printf("Found %d devices with vendor ID %s and product ID %s", len(devs), vid, usbtmcPID)
			if err != nil {
				// Close all devices and return error.
				for _, d := range devs {
					// I'm ignoring any errors on close at the moment.
					d.Close()
				}
				return nil, err
			}
			if len(devs) == 0 {
				return nil, fmt.Errorf("no devices found after reboot matching VID %s and PID %s", vid, usbtmcPID)
			}
		}
	} else if len(devs) == 0 {
		return nil, fmt.Errorf("no devices found matching VID %s and PID %s", vid, usbtmcPID)
	}

	// Close all except the first returned device.
	for i, d := range devs {
		if i != 0 {
			d.Close()
		}
	}

	// Pick the first device found.
	dev := devs[0]

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
	var intx *gousb.Interface
	// Loop through the interfaces
	for _, interfaceDesc := range cfg.Desc.Interfaces {
		// TODO(mdr): I should probably check this interface or config to confirm
		// it meets the USBTMC requirements.
		intf, err := cfg.Interface(interfaceDesc.Number, 0)
		if err != nil {
			return nil, err
		}
		intx = intf
		// Loop through all the endpoints on this interface
		for _, ep := range intf.Setting.Endpoints {
			if ep.Direction == gousb.EndpointDirectionOut && ep.TransferType == gousb.TransferTypeBulk {
				bulkOut, err = intf.OutEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
			if ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeBulk {
				bulkIn, err = intf.InEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
			if ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeInterrupt {
				intIn, err = intf.InEndpoint(ep.Number)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	d := Device{
		dev:                 dev,
		intf:                intx,
		cfg:                 cfg,
		BulkInEndpoint:      bulkIn,
		BulkOutEndpoint:     bulkOut,
		InterruptInEndpoint: intIn,
	}
	return &d, nil
}

func exitBootMode(dev *gousb.Device, bootPID gousb.ID) error {
	thirdIndex := uint16(0x0487)
	if bootPID == 0x2818 || bootPID == 0x3E18 {
		thirdIndex = 0x0484
	}
	bRequest := uint8(0x0C)
	value := uint16(0x0000)
	packets := []struct {
		bmRequestType uint8
		index         uint16
		data          []byte
	}{
		{0xC0, 0x047E, make([]byte, 0x01)},
		{0xC0, 0x047D, make([]byte, 0x06)},
		{0xC0, thirdIndex, make([]byte, 0x05)},
		{0xC0, 0x0472, make([]byte, 0x0C)},
		{0xC0, 0x047A, make([]byte, 0x01)},
		{0x40, 0x0475, []byte{0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x08, 0x01}},
	}
	for i, packet := range packets {
		_, err := dev.Control(
			packet.bmRequestType,
			bRequest,
			value,
			packet.index,
			packet.data,
		)
		if err != nil {
			return fmt.Errorf("Error sending control transfer #%d: %s", i+1, err)
		}
	}

	// We need to wait for the USB device to exit boot mode and reboot in normal
	// mode.
	rebootDelay := time.Second * 7
	time.Sleep(rebootDelay)
	return nil
}
