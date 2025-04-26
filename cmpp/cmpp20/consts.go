package cmpp20

import "errors"

const (
	// HeaderLength is the length of the CMPP 2.0 PDU header (12 bytes).
	HeaderLength = 4 + 4 + 4 // TotalLength(4) + CommandID(4) + SequenceID(4)

	// MaxConnectLength is the maximum length of a Connect PDU (39 bytes).
	MaxConnectLength = HeaderLength + 6 + 16 + 1 + 4
	// MaxConnectRespLength is the maximum length of a ConnectResp PDU (30 bytes).
	MaxConnectRespLength = HeaderLength + 1 + 16 + 1
	// MaxTerminateLength is the maximum length of a Terminate PDU (12 bytes).
	MaxTerminateLength = HeaderLength
	// MaxTerminateRespLength is the maximum length of a TerminateResp PDU (12 bytes).
	MaxTerminateRespLength = HeaderLength
	// MaxSubmitLength is the maximum length of a Submit PDU (variable, depends on DestUsrTL and MsgLength).
	// Calculation: Header + fixed fields + DestUsrTL(1) + DestTerminalID(21*99) + MsgLength(1) + MsgContent(160) + Reserve(8)
	MaxSubmitLength = HeaderLength + 8 + 1 + 1 + 1 + 1 + 10 + 1 + 21 + 1 + 1 + 1 + 6 + 2 + 6 + 17 + 17 + 21 + 1 + 21*99 + 1 + 160 + 8
	// MaxSubmitRespLength is the maximum length of a SubmitResp PDU (21 bytes).
	MaxSubmitRespLength = HeaderLength + 8 + 1
	// MaxActiveTestLength is the maximum length of an ActiveTest PDU (12 bytes).
	MaxActiveTestLength = HeaderLength
	// MaxActiveTestRespLength is the maximum length of an ActiveTestResp PDU (13 bytes).
	MaxActiveTestRespLength = HeaderLength + 1
	// MaxDeliverLength is the maximum length of a Deliver PDU (variable, depends on MsgLength).
	// Calculation: Header + fixed fields + MsgLength(1) + MsgContent(255) + Reserve(8)
	MaxDeliverLength = HeaderLength + 8 + 21 + 10 + 1 + 1 + 1 + 21 + 1 + 1 + 255 + 8
	// MaxDeliverRespLength is the maximum length of a DeliverResp PDU (21 bytes).
	MaxDeliverRespLength = HeaderLength + 8 + 1
	// MaxQueryLength is the maximum length of a Query PDU (39 bytes).
	MaxQueryLength = HeaderLength + 8 + 1 + 10 + 8
	// MaxQueryRespLength is the maximum length of a QueryResp PDU (51 bytes).
	MaxQueryRespLength = HeaderLength + 8 + 1 + 10 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 4
)

const (
	// DELIVERED indicates successful delivery status report code.
	DELIVERED = "DELIVRD"
	// UNDELIVERABLE indicates the message could not be delivered status report code.
	UNDELIVERABLE = "UNDELIV"
	// EXPIRED indicates the message validity period has expired status report code.
	EXPIRED = "EXPIRED"
	// DELETED indicates the message has been deleted status report code.
	DELETED = "DELETED"
	// ACCEPTED indicates the message is accepted by the SME status report code.
	ACCEPTED = "ACCEPTD"
	// UNKNOWN indicates the message state is unknown status report code.
	UNKNOWN = "UNKNOWN"
	// REJECTED indicates the message is rejected status report code.
	REJECTED = "REJECTD"
)

var (
	// ErrInvalidPudLength indicates an invalid PDU length error.
	ErrInvalidPudLength = errors.New("invalid pdu length")
)
