// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"encoding/binary"
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

// Instrument represents a USBTMC enabled device.
type Instrument struct {
	/*
		TODO(mdr) Not sure I like the name Instrument. Below are the names used by
		different projects:
		- python-ivi: class Instrument(object)
		- Linux usbtmc.c: struct usbtmc_device_data
		- pyvisa-py: class USBTMC(USBRaw)
	*/

	Device              *usb.Device
	BulkInEndpoint      usb.Endpoint
	BulkOutEndpoint     usb.Endpoint
	InterruptInEndpoint usb.Endpoint
	bTag                byte
	termChar            byte
	termCharEnabled     bool
}

// NewInstrument creates new USBTMC compliant device based on the given VISA
// resource name.
func (c *Context) NewInstrument(visaResourceName string) *Instrument {
	// FIXME(mdr) Need to speed this up. Currently, it takes about 5s to create a
	// new instrument.
	var usbtmcConfig uint8
	var usbtmcInterface uint8
	var usbtmcSetup uint8
	var bulkOutEndpointAddress uint8
	var bulkInEndpointAddress uint8
	var interruptInEndpointAddress uint8
	// TODO(mdr) Need to handle the error potentially return by ListDevices
	// FIXME(mdr) Need to handle error in case given a bad visaResource
	start := time.Now()
	devices, _ := c.ctx.ListDevices(FindUsbtmcFromResourceString(visaResourceName))
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
	return &Instrument{
		Device:              device,
		BulkInEndpoint:      bulkInEndpoint,
		BulkOutEndpoint:     bulkOutEndpoint,
		InterruptInEndpoint: interruptInEndpoint,
		termChar:            '\n',
		termCharEnabled:     true,
	}
}

// Close closes the Instrument.
func (instr *Instrument) Close() error {
	return instr.Device.Close()
}

// String providers the Stringer interface method for Instrument.
func (instr *Instrument) String() string {
	return instr.Device.Descriptor.SerialNumber
}

// Write writes the given buffer to the Instrument using the bulk out endpoint
// and returns the number of bytes written along with an error code.
func (instr *Instrument) Write(buf []byte) (n int, err error) {
	// FIXME(mdr): I need to change this so that I look at the size of the buf
	// being written to see if it can truly fit into one transfer.
	header := instr.createDevDepMsgOutBulkOutHeader(uint32(len(buf)), true)
	log.Printf("DevDepMsgOutBulkOutHeader = %v", header)
	data := append(header[:], buf...)
	if moduloFour := len(data) % 4; moduloFour > 0 {
		numAlignment := 4 - moduloFour
		alignment := bytes.Repeat([]byte{0x00}, numAlignment)
		data = append(data, alignment...)
	}
	n, err = instr.BulkOutEndpoint.Write(data)
	log.Printf("Wrote %d bytes to BulkOutEndpoint", n)
	log.Printf("BulkOutEndpoint data: %v", data)
	return n, err
}

// WriteString writes the given string to the Instrument and returns the number
// of bytes written along with an error code.
func (instr *Instrument) WriteString(s string) (n int, err error) {
	n, err = instr.Write([]byte(s))
	return n, err
}

// Read creates and sends the header on the bulk out endpoint and then reads
// from the bulk in endpoint.
func (instr *Instrument) Read(p []byte) (n int, err error) {
	// TODO(mdr): Should I pass in the header instead of creating it from p? That
	// seems like it would be better for SRP, but I think that would break the
	// golang Read() signature.
	header := instr.createRequestDevDepMsgInBulkOutHeader(uint32(len(p)))
	log.Printf("RequestDevDepMsg Header to write = %v", header)
	n, err = instr.BulkOutEndpoint.Write(header[:])
	n, err = instr.BulkInEndpoint.Read(p)
	log.Printf("Read %d bytes on BulkInEndpoint", n)
	return n, err
}

func (instr *Instrument) nextbTag() {
	instr.bTag = (instr.bTag % 255) + 1
}

// Create the first four bytes of the USBTMC meassage Bulk-OUT Header as shown
// in USBTMC Table 1. The msgID value must match USBTMC Table 2.
func (instr *Instrument) encodeBulkHeaderPrefix(msgID msgID) [4]byte {
	instr.nextbTag()
	return [4]byte{
		byte(msgID),
		instr.bTag,
		invertbTag(instr.bTag),
		reservedField,
	}
}

// Create the devDepMsgOut Bulk-OUT Header with command specific content as
// shown in USBTMC Table 3.
func (instr *Instrument) createDevDepMsgOutBulkOutHeader(transferSize uint32, eom bool) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := instr.encodeBulkHeaderPrefix(devDepMsgOut)
	// Offset 4-7: TransferSize
	// Per USBTMC Table 3, the TransferSize is the "total number of USBTMC
	// message data bytes to be sent in this USB transfer. This does not include
	// the number of bytes in this Bulk-OUT Header or alignment bytes. Sent least
	// significant byte first, most significant byte last. TransferSize must be >
	// 0x00000000."
	packedTransferSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(packedTransferSize, transferSize)
	// Offset 8: bmTransferAttributes
	// Per USBTMC Table 3, D0 of bmTransferAttributes:
	//   1 - The last USBTMC message data byte in the transfer is the last byte
	//       of the USBTMC message.
	//   0 - The last USBTMC message data byte in the transfer is not the last
	//       byte of the USBTMC message.
	// All other bits of bmTransferAttributes must be 0.
	bmTransferAttributes := byte(0x00)
	if eom {
		bmTransferAttributes = byte(0x01)
	}
	// Offset 9-11: reservedField. Must be 0x000000.
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
		reservedField,
		reservedField,
		reservedField,
	}
}

// Create the requestDevDepMsgIn Bulk-OUT Header with command specific
// content as shown in USBTMC Table 4.
func (instr *Instrument) createRequestDevDepMsgInBulkOutHeader(transferSize uint32) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := instr.encodeBulkHeaderPrefix(requestDevDepMsgIn)
	// Offset 4-7: TransferSize
	// Per USBTMC Table 4, the TransferSize is the "maximum number of USBTMC
	// message data bytes to be sent in response to the command. This does not
	// include the number of bytes in this Bulk-IN Header or alignment bytes.
	// Sent least significant byte first, most significant byte last.
	// TransferSize must be > 0x00000000."
	packedTransferSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(packedTransferSize, transferSize)
	// Offset 8: bmTransferAttributes
	// Per USBTMC Table 4, D1 of bmTransferAttributes:
	//   1 - "The Bulk-IN transfer must terminate on the specified TermChar. The
	//       Host may only set this bit if the USBTMC interface indicates it
	//       supports TermChar in the getCapabilities response packet."
	//   0 - "The device must ignore TermChar."
	// All other bits of bmTransferAttributes must be 0.
	bmTransferAttributes := byte(0x00)
	if instr.termCharEnabled {
		bmTransferAttributes = byte(0x02)
	}
	// Offset 9: TermChar
	// Offset 10-11: reservedField. Must be 0x000000.
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
		instr.termChar,
		reservedField,
		reservedField,
	}
}
