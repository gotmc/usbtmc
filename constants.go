package usbtmc

type bInterfaceClass byte

const Reserved = 0x00

/*
 * usbtmc.c by Agilent/Stefan Kopp sets this to 2048 with the comment:
 * Size of driver internal IO buffer. Must be multiple of 4 and at least as
 * large as wMaxPacketSize (which is usually 512 bytes).
 */
const UsbtmcSizeIobuffer = 1024 * 1024 // Set to 1MB
const UsbtmcHeaderSize = 12

const (
	AUDIO                bInterfaceClass = 0x01
	COMMUNICATIONS       bInterfaceClass = 0x02
	MISCELLANEOUS        bInterfaceClass = 0xef
	APPLICATION_SPECIFIC bInterfaceClass = 0xfe
	VENDOR_SPECIFIC      bInterfaceClass = 0xff
)

type bInterfaceSubClass byte

const (
	USBTMC_DEVICE bInterfaceSubClass = 0x03
)

type bInterfaceProtocol byte

const (
	USBTMC_PROTOCOL bInterfaceProtocol = 0x00
	USB488_PROTOCOL bInterfaceProtocol = 0x01
)

type msgId uint8

// Table 2 -- MsgID values
// USBTMC Specificiation 1.0, April 14, 2003
const (
	DEV_DEP_MSG_OUT            msgId = 1
	REQUEST_DEV_DEP_MSG_IN     msgId = 2
	DEV_DEP_MSG_IN             msgId = 2
	VENDOR_SPECIFIC_OUT        msgId = 126
	REQUEST_VENDOR_SPECIFIC_IN msgId = 127
	VENDOR_SPECIFIC_IN         msgId = 127
	// Table 1 -- USB488 defined MsgID values
	// USBTMC-USB488 Specification 1.0, April 14, 2003
	TRIGGER msgId = 128 // Table 1 USBTMC-USB488 Spec1.0, 14-Apr-03
)

type bRequest uint8

// Table 15 -- USBTMC bRequest values
// USBTMC Specificiation 1.0, April 14, 2003
const (
	INITIATE_ABORT_BULK_OUT     bRequest = 1
	CHECK_ABORT_BULK_OUT_STATUS bRequest = 2
	INITIATE_ABORT_BULK_IN      bRequest = 3
	CHECK_ABORT_BULK_IN_STATUS  bRequest = 4
	INITIATE_CLEAR              bRequest = 5
	CHECK_CLEAR_STATUS          bRequest = 6
	GET_CAPABILITIES            bRequest = 7
	INDICATOR_PULSE             bRequest = 64
)

// Table 9 -- USB488 defined bRequest values
// USBTMC-USB488 Specification 1.0, April 14, 2003
const (
	READ_STATUS_BYTE bRequest = 128
	REN_CONTROL      bRequest = 160
	GO_TO_LOCAL      bRequest = 161
	LOCAL_LOCKOUT    bRequest = 162
)

var requestDescription = map[bRequest]string{
	INITIATE_ABORT_BULK_OUT:     "Aborts a Bulk-OUT transfer.",
	CHECK_ABORT_BULK_OUT_STATUS: "Returns the status of the previously sent INITIATE_ABORT_BULK_OUT request.",
	INITIATE_ABORT_BULK_IN:      "Aborts a Bulk-IN transfer.",
	CHECK_ABORT_BULK_IN_STATUS:  "Returns the status of the previously sent INITIATE_ABORT_BULK_IN request.",
	INITIATE_CLEAR:              "Clears all previously sent pending and unprocessed Bulk-OUT USBTMC message content and clears all pending Bulk-IN transfers from the USBTMC interface.",
	CHECK_CLEAR_STATUS:          "Returns the status of the previously sent INITIATE_CLEAR request.",
	GET_CAPABILITIES:            "Returns attributes and capabilities of the USBTMC interface.",
	INDICATOR_PULSE:             "A mechanism to turn on an activity indicator for identification purposes. The device indicates whether or not it supports this request in the GET_CAPABILITIES response packet.",
	READ_STATUS_BYTE:            "Returns the IEEE 488 Status Byte.",
	REN_CONTROL:                 "Mechanism to enable or disable local controls on a device.",
	GO_TO_LOCAL:                 "Mechanism to enable local controls on a device.",
	LOCAL_LOCKOUT:               "Mechanism to disable local controls on a device.",
}

func (req bRequest) String() string {
	return requestDescription[req]
}

type UsbtmcStatus byte

// Table 16 -- USBTMC_status values
// USBTMC Specificiation 1.0, April 14, 2003
const (
	STATUS_SUCCESS                  UsbtmcStatus = 0x01
	STATUS_PENDING                  UsbtmcStatus = 0x02
	STATUS_FAILED                   UsbtmcStatus = 0x80
	STATUS_TRANSFER_NOT_IN_PROGRESS UsbtmcStatus = 0x81
	STATUS_SPLIT_NOT_IN_PROGRESS    UsbtmcStatus = 0x82
	STATUS_SPLIT_IN_PROGRESS        UsbtmcStatus = 0x83
)
