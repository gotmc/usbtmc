package usbtmc

import "testing"

func TestRequestString(t *testing.T) {
	testCases := []struct {
		request     bRequest
		description string
	}{
		{INITIATE_ABORT_BULK_OUT, "Aborts a Bulk-OUT transfer."},
		{READ_STATUS_BYTE, "Returns the IEEE 488 Status Byte."},
	}
	for _, testCase := range testCases {
		if testCase.request.String() != testCase.description {
			t.Errorf(
				"request.String() == %x, want %x for request %x",
				testCase.request.String(),
				testCase.description,
				testCase.request,
			)
		}
	}
}
