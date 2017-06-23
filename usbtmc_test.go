// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"testing"
)

func TestEncodeBulkHeaderPrefix(t *testing.T) {
	tests := []struct {
		msgID        msgID
		d            Device
		headerPrefix [4]byte
	}{
		{devDepMsgOut, Device{bTag: 1}, [4]byte{0x01, 0x02, 0xfd, 0x00}},
		{devDepMsgOut, Device{bTag: 128}, [4]byte{0x01, 0x81, 0x7e, 0x00}},
		{devDepMsgOut, Device{bTag: 254}, [4]byte{0x01, 0xff, 0x00, 0x00}},
		{devDepMsgOut, Device{bTag: 255}, [4]byte{0x01, 0x01, 0xfe, 0x00}},
		{requestDevDepMsgIn, Device{bTag: 3}, [4]byte{0x02, 0x04, 0xfb, 0x00}},
		{vendorSpecificOut, Device{bTag: 3}, [4]byte{0x7e, 0x04, 0xfb, 0x00}},
		{requestVendorSpecificIn, Device{bTag: 3}, [4]byte{0x7f, 0x04, 0xfb, 0x00}},
	}
	for _, test := range tests {
		headerPrefix := test.d.encodeBulkHeaderPrefix(test.msgID)
		if headerPrefix != test.headerPrefix {
			t.Errorf(
				"headerPrefix == %x, want %x",
				headerPrefix, test.headerPrefix)
		}
	}
}

func TestNextbTag(t *testing.T) {
	tests := []struct {
		nextbTag byte
		d        Device
	}{
		{1, Device{bTag: 255}},
		{2, Device{bTag: 1}},
		{11, Device{bTag: 10}},
		{130, Device{bTag: 129}},
		{200, Device{bTag: 199}},
		{254, Device{bTag: 253}},
		{255, Device{bTag: 254}},
	}
	for _, test := range tests {
		test.d.nextbTag()
		if test.nextbTag != test.d.bTag {
			t.Errorf(
				"bTag == %x, want %x",
				test.d.bTag, test.nextbTag)
		}
	}
}

func TestCreateDevDepMsgOutBulkOutHeader(t *testing.T) {
	tests := []struct {
		transferSize uint32
		eom          bool
		d            Device
		desired      [12]byte
	}{
		{
			9,
			true,
			Device{bTag: 255},
			[12]byte{0x01, 0x01, 0xfe, 0x00, 0x09, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			256,
			false,
			Device{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			256,
			true,
			Device{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			512,
			true,
			Device{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x02, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
	}
	for _, test := range tests {
		result := test.d.createDevDepMsgOutBulkOutHeader(test.transferSize, test.eom)
		if result != test.desired {
			t.Errorf("BulkOutHeader == %x, want %x", result, test.desired)
		}
	}
}
