// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"fmt"
	"testing"
)

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
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("bTag_%d", tc.bTag), func(t *testing.T) {
			got := nextbTag(tc.bTag)
			if got != tc.nextbTag {
				t.Errorf(
					"nextbTag == %x, want %x for given bTag %x",
					got, tc.nextbTag, tc.bTag)
			}
		})
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
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("bTag_%02x", tc.bTag), func(t *testing.T) {
			got := invertbTag(tc.bTag)
			if got != tc.bTagInverse {
				t.Errorf(
					"bTagInverse == %x, want %x for bTag %x",
					got, tc.bTagInverse, tc.bTag)
			}
		})
	}
}

func TestEncodeBulkHeaderPrefix(t *testing.T) {
	tests := []struct {
		name         string
		msgID        msgID
		bTag         byte
		headerPrefix [4]byte
	}{
		{"devDepMsgOut_bTag2", devDepMsgOut, 2, [4]byte{0x01, 0x02, 0xfd, 0x00}},
		{"devDepMsgOut_bTag129", devDepMsgOut, 129, [4]byte{0x01, 0x81, 0x7e, 0x00}},
		{"devDepMsgOut_bTag255", devDepMsgOut, 255, [4]byte{0x01, 0xff, 0x00, 0x00}},
		{"devDepMsgOut_bTag1", devDepMsgOut, 1, [4]byte{0x01, 0x01, 0xfe, 0x00}},
		{"requestDevDepMsgIn_bTag4", requestDevDepMsgIn, 4, [4]byte{0x02, 0x04, 0xfb, 0x00}},
		{"vendorSpecificOut_bTag4", vendorSpecificOut, 4, [4]byte{0x7e, 0x04, 0xfb, 0x00}},
		{"requestVendorSpecificIn_bTag4", requestVendorSpecificIn, 4, [4]byte{0x7f, 0x04, 0xfb, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeBulkHeaderPrefix(tt.bTag, tt.msgID)
			if got != tt.headerPrefix {
				t.Errorf(
					"headerPrefix == %x, want %x",
					got, tt.headerPrefix)
			}
		})
	}
}

func TestEncodeBulkOutHeader(t *testing.T) {
	tests := []struct {
		name         string
		transferSize uint32
		eom          bool
		bTag         byte
		desired      [12]byte
	}{
		{
			"size9_eom_bTag1",
			9, true, 1,
			[12]byte{0x01, 0x01, 0xfe, 0x00, 0x09, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			"size256_noEom_bTag2",
			256, false, 2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			"size256_eom_bTag2",
			256, true, 2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			"size512_eom_bTag2",
			512, true, 2,
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x02, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeBulkOutHeader(tt.bTag, tt.transferSize, tt.eom)
			if got != tt.desired {
				t.Errorf("BulkOutHeader == %x, want %x", got, tt.desired)
			}
		})
	}
}

func TestEncodeMsgInBulkOutHeader(t *testing.T) {
	tests := []struct {
		name            string
		bTag            byte
		transferSize    uint32
		termCharEnabled bool
		termChar        byte
		desired         [12]byte
	}{
		{
			"size9_termChar_bTag1",
			1, 9, true, '\n',
			[12]byte{0x02, 0x01, 0xfe, 0x00, 0x09, 0x00, 0x00, 0x00, 0x02, 0x0a, 0x00, 0x00},
		},
		{
			"size512_termChar_bTag2",
			2, 512, true, '\n',
			[12]byte{0x02, 0x02, 0xfd, 0x00, 0x00, 0x02, 0x00, 0x00, 0x02, 0x0a, 0x00, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeMsgInBulkOutHeader(
				tt.bTag,
				tt.transferSize,
				tt.termCharEnabled,
				tt.termChar,
			)
			if got != tt.desired {
				t.Errorf("BulkOutHeader == %x, want %x", got, tt.desired)
			}
		})
	}
}
