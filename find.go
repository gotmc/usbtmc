// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"log"

	"github.com/truveris/gousb/usb"
)

// FindAllUsbtmcInterfaces seems to do way too much.
func FindAllUsbtmcInterfaces(desc *usb.Descriptor) bool {
	hasUsbtmcInterface := false
	switch {
	case desc.Class == 0x00 && desc.SubClass == 0x00:
		for _, config := range desc.Configs {
			for _, iface := range config.Interfaces {
				for _, setup := range iface.Setups {
					switch {
					case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 00:
						hasUsbtmcInterface = true
						log.Printf(
							"USBTMC interface found on S/N %s, Vendor %d, Product %d",
							desc.SerialNumber,
							desc.Vendor,
							desc.Product,
						)
						log.Printf("--> %s, %s", config, setup)
					case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 01:
						hasUsbtmcInterface = true
						log.Printf(
							"USB488 interface found on S/N %s, Vendor %d, Product %d",
							desc.SerialNumber,
							desc.Vendor,
							desc.Product,
						)
						log.Printf("--> %s, %s", config, setup)
					}
				}
			}
		}
	default:
		return false
	}
	return hasUsbtmcInterface
}

// FindVisaResourceName returns a pointer to a usb.Device given the
// visaResourceName and Context.
func FindVisaResourceName(visaResourceName string, c *usb.Context) (*usb.Device, error) {
	devices, err := c.ListDevices(FindUsbtmcFromResourceString(visaResourceName))
	return devices[0], err
}

// FindUsbtmcFromResource needs a better comment.
func FindUsbtmcFromResource(visaResource *VisaResource) func(desc *usb.Descriptor) bool {
	return func(desc *usb.Descriptor) bool {
		hasUsbtmcInterface := false
		switch {
		case uint16(desc.Vendor) == visaResource.manufacturerID &&
			uint16(desc.Product) == visaResource.modelCode &&
			desc.Class == 0x00 && desc.SubClass == 0x00:
			for _, config := range desc.Configs {
				for _, iface := range config.Interfaces {
					for _, setup := range iface.Setups {
						switch {
						case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 00:
							hasUsbtmcInterface = true
							log.Printf(
								"USBTMC interface found on S/N %s, Vendor %d, Product %d",
								desc.SerialNumber,
								desc.Vendor,
								desc.Product,
							)
							log.Printf("--> %s, %s", config, setup)
						case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 01:
							hasUsbtmcInterface = true
							log.Printf(
								"USB488 interface found on S/N %s, Vendor %d, Product %d",
								desc.SerialNumber,
								desc.Vendor,
								desc.Product,
							)
							log.Printf("--> %s, %s", config, setup)
						}
					}
				}
			}
		default:
			return false
		}
		return hasUsbtmcInterface
	}
}

// FindUsbtmcFromResourceString needs a better comment.
func FindUsbtmcFromResourceString(resourceString string) func(desc *usb.Descriptor) bool {
	visaResource, err := NewVisaResource(resourceString)
	if err != nil {
		return func(desc *usb.Descriptor) bool {
			return false
		}
	}

	return FindUsbtmcFromResource(visaResource)

}
