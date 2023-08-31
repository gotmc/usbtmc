// Copyright (c) 2015-2020 The usbtmc developers. All rights reserved.
// Project site: https://github.com/gotmc/usbtmc
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package usbtmc

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// VisaResource represents a VISA enabled piece of test equipment.
type VisaResource struct {
	resourceString string
	interfaceType  string
	boardIndex     int
	manufacturerID int
	modelCode      int
	serialNumber   string
	interfaceIndex int
	resourceClass  string
}

// NewVisaResource creates a new VisaResource using the given VISA resourceString.
func NewVisaResource(resourceString string) (*VisaResource, error) {
	visa := &VisaResource{
		resourceString: resourceString,
		interfaceType:  "",
		boardIndex:     0,
		manufacturerID: 0,
		modelCode:      0,
		serialNumber:   "",
		interfaceIndex: 0,
		resourceClass:  "",
	}
	regString := `^(?P<interfaceType>[A-Za-z]+)(?P<boardIndex>\d*)::` +
		`(?P<manufacturerID>[^\s:]+)::` +
		`(?P<modelCode>[^\s:]+)` +
		`(::(?P<serialNumber>[^\s:]+))?` +
		`::(?P<resourceClass>[^\s:]+)$`

	re := regexp.MustCompile(regString)
	res := re.FindStringSubmatch(resourceString)
	subexpNames := re.SubexpNames()
	matchMap := map[string]string{}
	for i, n := range res {
		matchMap[subexpNames[i]] = string(n)
	}

	if strings.ToUpper(matchMap["interfaceType"]) != "USB" {
		return visa, errors.New("visa: interface type was not usb")
	}
	visa.interfaceType = "USB"

	if matchMap["boardIndex"] != "" {
		boardIndex, err := strconv.ParseUint(matchMap["boardIndex"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: boardIndex error")
		}
		visa.boardIndex = int(boardIndex)
	}

	if matchMap["manufacturerID"] != "" {
		manufacturerID, err := strconv.ParseUint(matchMap["manufacturerID"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: manufacturerID error")
		}
		visa.manufacturerID = int(manufacturerID)
	}

	if matchMap["modelCode"] != "" {
		modelCode, err := strconv.ParseUint(matchMap["modelCode"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: modelCode error")
		}
		visa.modelCode = int(modelCode)
	}

	visa.serialNumber = matchMap["serialNumber"]

	if strings.ToUpper(matchMap["resourceClass"]) != "INSTR" {
		return visa, errors.New("visa: resource class was not instr")
	}
	visa.resourceClass = "INSTR"

	return visa, nil

}
