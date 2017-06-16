// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.
package usbtmc

import "testing"

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
