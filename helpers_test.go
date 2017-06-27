// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "testing"

func TestNextBtag(t *testing.T) {
	testCases := []struct {
		bTag     byte
		nextbTag byte
	}{
		{0x01, 0x02},
		{0xff, 0x01},
		{255, 1},
		{1, 2},
		{10, 11},
		{254, 255},
	}
	for _, testCase := range testCases {
		nextbTag := nextbTag(testCase.bTag)
		if nextbTag != testCase.nextbTag {
			t.Errorf(
				"nextbTag == %x, want %x for given bTag %x",
				nextbTag, testCase.nextbTag, testCase.bTag)
		}
	}
}
func TestInvertingBtag(t *testing.T) {
	testCases := []struct {
		bTag        byte
		bTagInverse byte
	}{
		{0x00, 0xff},
		{0x0f, 0xf0},
		{0x55, 0xaa},
		{0xaa, 0x55},
		{0xf0, 0x0f},
		{0xff, 0x00},
	}
	for _, testCase := range testCases {
		bTagInverse := invertbTag(testCase.bTag)
		if bTagInverse != testCase.bTagInverse {
			t.Errorf(
				"bTagInverse == %x, want %x for bTag %x",
				bTagInverse, testCase.bTagInverse, testCase.bTag)
		}
	}
}

func TestEncodeBulkHeaderPrefix(t *testing.T) {
	tests := []struct {
		msgID        msgID
		bTag         byte
		headerPrefix [4]byte
	}{
		{devDepMsgOut, 2, [4]byte{0x01, 0x02, 0xfd, 0x00}},
		{devDepMsgOut, 129, [4]byte{0x01, 0x81, 0x7e, 0x00}},
		{devDepMsgOut, 255, [4]byte{0x01, 0xff, 0x00, 0x00}},
		{devDepMsgOut, 1, [4]byte{0x01, 0x01, 0xfe, 0x00}},
		{requestDevDepMsgIn, 4, [4]byte{0x02, 0x04, 0xfb, 0x00}},
		{vendorSpecificOut, 4, [4]byte{0x7e, 0x04, 0xfb, 0x00}},
		{requestVendorSpecificIn, 4, [4]byte{0x7f, 0x04, 0xfb, 0x00}},
	}
	for _, test := range tests {
		headerPrefix := encodeBulkHeaderPrefix(test.bTag, test.msgID)
		if headerPrefix != test.headerPrefix {
			t.Errorf(
				"headerPrefix == %x, want %x",
				headerPrefix, test.headerPrefix)
		}
	}
}

func TestCreateDevDepMsgOutBulkOutHeader(t *testing.T) {
	tests := []struct {
		transferSize uint32
		eom          bool
		bTag         byte
		desired      [12]byte
	}{
		{
			9,
			true,
			1,
			[12]byte{0x01, 0x01, 0xfe, 0x00, 0x09, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			256,
			false,
			2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			256,
			true,
			2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			512,
			true,
			2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x02, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
	}
	for _, test := range tests {
		result := createDevDepMsgOutBulkOutHeader(test.bTag, test.transferSize, test.eom)
		if result != test.desired {
			t.Errorf("BulkOutHeader == %x, want %x", result, test.desired)
		}
	}
}
