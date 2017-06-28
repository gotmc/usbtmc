// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "encoding/binary"

func invertbTag(bTag byte) byte {
	return bTag ^ 0xff
}

// nextbTag returns the next bTag given the current bTag. Per the USBTMC
// standard, "the Host must set bTag such that 1<=bTag<=255."
func nextbTag(bTag byte) byte {
	return (bTag % 255) + 1
}

// Create the devDepMsgOut Bulk-OUT Header with command specific content as
// shown in USBTMC Table 3.
func createDevDepMsgOutBulkOutHeader(bTag byte, transferSize uint32, eom bool) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := encodeBulkHeaderPrefix(bTag, devDepMsgOut)
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

// Create the first four bytes of the USBTMC meassage Bulk-OUT Header as shown
// in USBTMC Table 1. The msgID value must match USBTMC Table 2.
func encodeBulkHeaderPrefix(bTag byte, msgID msgID) [4]byte {
	return [4]byte{
		byte(msgID),
		bTag,
		invertbTag(bTag),
		reservedField,
	}
}

// Create the requestDevDepMsgIn Bulk-OUT Header with command specific
// content as shown in USBTMC Table 4.
func createRequestDevDepMsgInBulkOutHeader(bTag byte, transferSize uint32, termCharEnabled bool, termChar byte) [12]byte {
	// Offset 0-3: See Table 1.
	prefix := encodeBulkHeaderPrefix(bTag, requestDevDepMsgIn)
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
	if termCharEnabled {
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
		termChar,
		reservedField,
		reservedField,
	}
}
