// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/truveris/gousb/usb"
)

// Device represents a USBTMC enabled device.
type Device struct {
	USBDevice           *usb.Device
	BulkInEndpoint      usb.Endpoint
	BulkOutEndpoint     usb.Endpoint
	InterruptInEndpoint usb.Endpoint
	bTag                byte
	termChar            byte
	termCharEnabled     bool
}

// Close closes the Device.
func (d *Device) Close() error {
	return d.USBDevice.Close()
}

// String providers the Stringer interface method for Device.
func (d *Device) String() string {
	return d.USBDevice.Descriptor.SerialNumber
}

// Write writes the given buffer to the Device using the bulk out endpoint
// and returns the number of bytes written along with an error code.
func (d *Device) Write(buf []byte) (n int, err error) {
	// FIXME(mdr): I need to change this so that I look at the size of the buf
	// being written to see if it can truly fit into one transfer, and if not
	// split it into multiple transfers.
	header := d.createDevDepMsgOutBulkOutHeader(uint32(len(buf)), true)
	// log.Printf("DevDepMsgOutBulkOutHeader = %v", header)
	data := append(header[:], buf...)
	if moduloFour := len(data) % 4; moduloFour > 0 {
		numAlignment := 4 - moduloFour
		alignment := bytes.Repeat([]byte{0x00}, numAlignment)
		data = append(data, alignment...)
	}
	n, err = d.BulkOutEndpoint.Write(data)
	// log.Printf("Wrote %d bytes to BulkOutEndpoint", n)
	// log.Printf("BulkOutEndpoint data: %v", data)
	return n, err
}

// WriteString writes the given string to the Device and returns the number
// of bytes written along with an error code.
func (d *Device) WriteString(s string) (n int, err error) {
	n, err = d.Write([]byte(s))
	return n, err
}

// Read creates and sends the header on the bulk out endpoint and then reads
// from the bulk in endpoint.
func (d *Device) Read(p []byte) (n int, err error) {
	// TODO(mdr): Should I pass in the header instead of creating it from p? That
	// seems like it would be better for SRP, but I think that would break the
	// golang Read() signature.
	header := d.createRequestDevDepMsgInBulkOutHeader(uint32(len(p)))
	log.Printf("RequestDevDepMsg Header to write = %v", header)
	n, err = d.BulkOutEndpoint.Write(header[:])
	n, err = d.BulkInEndpoint.Read(p)
	log.Printf("Read %d bytes on BulkInEndpoint", n)
	return n, err
}

// Query writes a SCPI command as a string and then returns the queried result
// as a string.
func (d *Device) Query(s string) (string, error) {
	_, err := d.WriteString(s)
	if err != nil {
		return "", err
	}
	var b []byte
	_, err = d.Read(b)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(b)
	return buf.ReadString(0xA)
}

func (d *Device) nextbTag() {
	d.bTag = (d.bTag % 255) + 1
}

// Create the first four bytes of the USBTMC meassage Bulk-OUT Header as shown
// in USBTMC Table 1. The msgID value must match USBTMC Table 2.
func (d *Device) encodeBulkHeaderPrefix(msgID msgID) [4]byte {
	d.nextbTag()
	return [4]byte{
		byte(msgID),
		d.bTag,
		invertbTag(d.bTag),
		reservedField,
	}
}

// Create the devDepMsgOut Bulk-OUT Header with command specific content as
// shown in USBTMC Table 3.
func (d *Device) createDevDepMsgOutBulkOutHeader(transferSize uint32, eom bool) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := d.encodeBulkHeaderPrefix(devDepMsgOut)
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
func (d *Device) createRequestDevDepMsgInBulkOutHeader(transferSize uint32) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := d.encodeBulkHeaderPrefix(requestDevDepMsgIn)
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
	if d.termCharEnabled {
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
		d.termChar,
		reservedField,
		reservedField,
	}
}
