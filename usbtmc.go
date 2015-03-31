package usbtmc

import (
	"encoding/binary"
	"log"

	"github.com/truveris/gousb/usb"
)

type UsbtmcContext struct {
	ctx *usb.Context
}

func NewContext() *UsbtmcContext {
	c := &UsbtmcContext{
		ctx: usb.NewContext(),
	}
	return c
}

func (c *UsbtmcContext) Debug(level int) {
	c.ctx.Debug(level)
}

func (c *UsbtmcContext) Close() error {
	return c.ctx.Close()
}

type Instrument struct {
	Device              *usb.Device
	BulkInEndpoint      usb.Endpoint
	BulkOutEndpoint     usb.Endpoint
	InterruptInEndpoint usb.Endpoint
	bTag                byte
}

func (c *UsbtmcContext) NewInstrument(visaResourceName string) *Instrument {
	var usbtmcConfig uint8
	var usbtmcInterface uint8
	var usbtmcSetup uint8
	var bulkOutEndpointAddress uint8
	var bulkInEndpointAddress uint8
	var interruptInEndpointAddress uint8
	// TODO(mdr) Need to handle the error potentially return by ListDevices
	// FIXME(mdr) Need to handle error in case given a bad visaResource
	devices, _ := c.ctx.ListDevices(FindUsbtmcFromResourceString(visaResourceName))
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

	return &Instrument{
		Device:              device,
		BulkInEndpoint:      bulkInEndpoint,
		BulkOutEndpoint:     bulkOutEndpoint,
		InterruptInEndpoint: interruptInEndpoint,
	}
}

func (i *Instrument) Close() error {
	return i.Device.Close()
}

func (inst *Instrument) nextbTag() {
	inst.bTag = (inst.bTag % 255) + 1
}

func (inst *Instrument) createBulkOutHeaderPrefix(msgId MsgId) [4]byte {
	inst.nextbTag()
	return [4]byte{byte(msgId), inst.bTag, invertbTag(inst.bTag), Reserved}
}

func (inst *Instrument) createDevDepMsgOutBulkOutHeader(transferSize uint32, eom bool) [12]byte {
	prefix := inst.createBulkOutHeaderPrefix(DEV_DEP_MSG_OUT)
	packedTransferSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(packedTransferSize, transferSize)
	bmTransferAttributes := byte(0x00)
	if eom {
		bmTransferAttributes = byte(0x01)
	}
	return [12]byte{
		prefix[0],
		prefix[1],
		prefix[2],
		prefix[3],
		packedTransferSize[0],
		packedTransferSize[1],
		packedTransferSize[2],
		packedTransferSize[3],
		bmTransferAttributes,
		Reserved,
		Reserved,
		Reserved,
	}
}

func (inst *Instrument) SendScpi(scpi string) (int, error) {

}
