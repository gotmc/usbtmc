package usbtmc

import "testing"

func TestRequestString(t *testing.T) {
	testCases := []struct {
		request     bRequest
		description string
	}{
		{initiateAbortBulkOut, "Aborts a Bulk-OUT transfer."},
		{readStatusByte, "Returns the IEEE 488 Status Byte."},
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
