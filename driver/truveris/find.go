// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package truveris

import (
	"log"

	"github.com/truveris/gousb/usb"
)

// FindAllUsbtmcInterfaces seems to do way too much.
func findAllUsbtmcInterfaces(desc *usb.Descriptor) bool {
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
					case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 01:
						hasUsbtmcInterface = true
						log.Printf(
							"USB488 interface found on S/N %s, Vendor %d, Product %d",
							desc.SerialNumber,
							desc.Vendor,
							desc.Product,
						)
					}
				}
			}
		}
	default:
		return false
	}
	return hasUsbtmcInterface
}

// FindUsbtmcFromResource needs a better comment.
func findDeviceByVIDPID(VID, PID uint16) func(desc *usb.Descriptor) bool {
	return func(desc *usb.Descriptor) bool {
		hasUsbtmcInterface := false
		switch {
		case uint16(desc.Vendor) == VID &&
			uint16(desc.Product) == PID &&
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
						case setup.IfClass == 0xfe && setup.IfSubClass == 0x03 && setup.IfProtocol == 01:
							hasUsbtmcInterface = true
							log.Printf(
								"USB488 interface found on S/N %s, Vendor %d, Product %d",
								desc.SerialNumber,
								desc.Vendor,
								desc.Product,
							)
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
