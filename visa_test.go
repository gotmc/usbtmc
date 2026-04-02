// Copyright (c) 2015-2026 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"testing"
)

func TestParsingVisaResourceString(t *testing.T) {
	testCases := []struct {
		name           string
		resourceString string
		interfaceType  string
		boardIndex     int
		manufacturerID int
		modelCode      int
		serialNumber   string
		interfaceIndex int
		resourceClass  string
		isError        bool
		errorString    string
	}{
		{
			"lowercase_with_serial",
			"usb0::2391::1031::MY44123456::INSTR",
			"USB", 0, 2391, 1031, "MY44123456", 0, "INSTR",
			false, "",
		},
		{
			"no_board_index_no_serial",
			"USB::1234::5678::INSTR",
			"USB", 0, 1234, 5678, "", 0, "INSTR",
			false, "",
		},
		{
			"with_serial",
			"USB::1234::5678::SERIAL::INSTR",
			"USB", 0, 1234, 5678, "SERIAL", 0, "INSTR",
			false, "",
		},
		{
			"hex_vid_pid",
			"USB0::0x1234::0x5678::INSTR",
			"USB", 0, 4660, 22136, "", 0, "INSTR",
			false, "",
		},
		{
			"hex_with_serial_and_interface",
			"USB0::0x0957::0x2007::MY57004760::0::INSTR",
			"USB", 0, 2391, 8199, "MY57004760", 0, "INSTR",
			false, "",
		},
		{
			"wrong_interface_type",
			"UBS::1234::5678::INSTR",
			"", 0, 0, 0, "", 0, "",
			true, "visa: interface type was not usb",
		},
		{
			"wrong_resource_class",
			"USB::1234::5678::INTSR",
			"USB", 0, 1234, 5678, "", 0, "",
			true, "visa: resource class was not instr",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource, err := NewVisaResource(tc.resourceString)
			if resource.interfaceType != tc.interfaceType {
				t.Errorf(
					"interfaceType == %s, want %s",
					resource.interfaceType, tc.interfaceType,
				)
			}
			if resource.boardIndex != tc.boardIndex {
				t.Errorf(
					"boardIndex == %d, want %d",
					resource.boardIndex, tc.boardIndex,
				)
			}
			if resource.manufacturerID != tc.manufacturerID {
				t.Errorf(
					"manufacturerID == %d, want %d",
					resource.manufacturerID, tc.manufacturerID,
				)
			}
			if resource.modelCode != tc.modelCode {
				t.Errorf(
					"modelCode == %d, want %d",
					resource.modelCode, tc.modelCode,
				)
			}
			if resource.serialNumber != tc.serialNumber {
				t.Errorf(
					"serialNumber == %s, want %s",
					resource.serialNumber, tc.serialNumber,
				)
			}
			if resource.interfaceIndex != tc.interfaceIndex {
				t.Errorf(
					"interfaceIndex == %d, want %d",
					resource.interfaceIndex, tc.interfaceIndex,
				)
			}
			if resource.resourceClass != tc.resourceClass {
				t.Errorf(
					"resourceClass == %s, want %s",
					resource.resourceClass, tc.resourceClass,
				)
			}
			if err != nil && tc.isError {
				if err.Error() != tc.errorString {
					t.Errorf("err == %s, want %s", err, tc.errorString)
				}
			}
			if err != nil && !tc.isError {
				t.Errorf("unexpected error: %q", err)
			}
		})
	}
}
