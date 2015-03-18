package usbtmc

import (
	"github.com/truveris/gousb/usb"
	"log"
)

func FindUsbtmcInterfaces(desc *usb.Descriptor) bool {
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
