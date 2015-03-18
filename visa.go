package usbtmc

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type VisaResource struct {
	resourceString string
	interfaceType  string
	boardIndex     uint16
	manufacturerId uint16
	modelCode      uint16
	serialNumber   string
	interfaceIndex uint16
	resourceClass  string
}

func NewVisaResource(resourceString string) (visa *VisaResource, err error) {
	visa = &VisaResource{
		resourceString: resourceString,
		interfaceType:  "",
		boardIndex:     0,
		manufacturerId: 0,
		modelCode:      0,
		serialNumber:   "",
		interfaceIndex: 0,
		resourceClass:  "",
	}
	regString := `^(?P<interfaceType>[A-Za-z]+)(?P<boardIndex>\d*)::` +
		`(?P<manufacturerId>[^\s:]+)::` +
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
	} else {
		visa.interfaceType = "USB"
	}

	if matchMap["boardIndex"] != "" {
		boardIndex, err := strconv.ParseUint(matchMap["boardIndex"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: boardIndex error")
		}
		visa.boardIndex = uint16(boardIndex)
	}

	if matchMap["manufacturerId"] != "" {
		manufacturerId, err := strconv.ParseUint(matchMap["manufacturerId"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: manufacturerId error")
		}
		visa.manufacturerId = uint16(manufacturerId)
	}

	if matchMap["modelCode"] != "" {
		modelCode, err := strconv.ParseUint(matchMap["modelCode"], 0, 16)
		if err != nil {
			return visa, errors.New("visa: modelCode error")
		}
		visa.modelCode = uint16(modelCode)
	}

	visa.serialNumber = matchMap["serialNumber"]

	if strings.ToUpper(matchMap["resourceClass"]) != "INSTR" {
		return visa, errors.New("visa: resource class was not instr")
	} else {
		visa.resourceClass = "INSTR"
	}

	return visa, nil

}
