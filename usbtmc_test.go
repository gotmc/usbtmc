package usbtmc

import (
	"testing"
)

func TestInvertingBtag(t *testing.T) {
	testCases := []struct {
		bTag        byte
		bTagInverse byte
	}{
		{0x00, 0xff},
		{0x0f, 0xf0},
		{0xf0, 0x0f},
		{0xaa, 0x55},
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
