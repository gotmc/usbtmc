package usbtmc

import (
	"testing"
)

func TestEncodeBulkHeaderPrefix(t *testing.T) {
	tests := []struct {
		msgId        msgId
		instrument   Instrument
		headerPrefix [4]byte
	}{
		{DEV_DEP_MSG_OUT, Instrument{bTag: 1}, [4]byte{0x01, 0x02, 0xfd, 0x00}},
		{DEV_DEP_MSG_OUT, Instrument{bTag: 128}, [4]byte{0x01, 0x81, 0x7e, 0x00}},
		{DEV_DEP_MSG_OUT, Instrument{bTag: 254}, [4]byte{0x01, 0xff, 0x00, 0x00}},
		{DEV_DEP_MSG_OUT, Instrument{bTag: 255}, [4]byte{0x01, 0x01, 0xfe, 0x00}},
		{REQUEST_DEV_DEP_MSG_IN, Instrument{bTag: 3}, [4]byte{0x02, 0x04, 0xfb, 0x00}},
		{VENDOR_SPECIFIC_OUT, Instrument{bTag: 3}, [4]byte{0x7e, 0x04, 0xfb, 0x00}},
		{REQUEST_VENDOR_SPECIFIC_IN, Instrument{bTag: 3}, [4]byte{0x7f, 0x04, 0xfb, 0x00}},
	}
	for _, test := range tests {
		headerPrefix := test.instrument.encodeBulkHeaderPrefix(test.msgId)
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
		inst     Instrument
	}{
		{1, Instrument{bTag: 255}},
		{2, Instrument{bTag: 1}},
		{11, Instrument{bTag: 10}},
		{130, Instrument{bTag: 129}},
		{200, Instrument{bTag: 199}},
		{254, Instrument{bTag: 253}},
		{255, Instrument{bTag: 254}},
	}
	for _, test := range tests {
		test.inst.nextbTag()
		if test.nextbTag != test.inst.bTag {
			t.Errorf(
				"bTag == %x, want %x",
				test.inst.bTag, test.nextbTag)
		}
	}
}

func TestCreateDevDepMsgOutBulkOutHeader(t *testing.T) {
	tests := []struct {
		transferSize uint32
		eom          bool
		inst         Instrument
		desired      [12]byte
	}{
		{
			9,
			true,
			Instrument{bTag: 255},
			[12]byte{0x01, 0x01, 0xfe, 0x00, 0x09, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			256,
			false,
			Instrument{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			256,
			true,
			Instrument{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			512,
			true,
			Instrument{bTag: 1},
			[12]byte{0x01, 0x02, 0xfd, 0x00, 0x00, 0x02, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
	}
	for _, test := range tests {
		result := test.inst.createDevDepMsgOutBulkOutHeader(test.transferSize, test.eom)
		if result != test.desired {
			t.Errorf("BulkOutHeader == %x, want %x", result, test.desired)
		}
	}
}
