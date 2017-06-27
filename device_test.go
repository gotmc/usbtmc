// Copyright (c) 2015-2017 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import "testing"

func TestNextbTag(t *testing.T) {
	tests := []struct {
		nextbTag byte

		d Device
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
