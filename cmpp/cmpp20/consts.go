package cmpp20

import "errors"

const (
	HeaderLength = 4 + 4 + 4 // cmpp2 PDU Header的长度

	MaxConnectLength        = HeaderLength + 6 + 16 + 1 + 4
	MaxConnectRespLength    = HeaderLength + 1 + 16 + 1
	MaxTerminateLength      = HeaderLength
	MaxTerminateRespLength  = HeaderLength
	MaxSubmitLength         = HeaderLength + 8 + 1 + 1 + 1 + 1 + 10 + 1 + 21 + 1 + 1 + 1 + 6 + 2 + 6 + 17 + 17 + 21 + 1 + 21*99 + 1 + 160 + 8
	MaxSubmitRespLength     = HeaderLength + 8 + 1
	MaxActiveTestLength     = HeaderLength
	MaxActiveTestRespLength = HeaderLength + 1
	MaxDeliverLength        = HeaderLength + 8 + 21 + 10 + 1 + 1 + 1 + 21 + 1 + 1 + 255 + 8
	MaxDeliverRespLength    = HeaderLength + 8 + 1
	MaxQueryLength          = HeaderLength + 8 + 1 + 10 + 8
	MaxQueryRespLength      = HeaderLength + 8 + 1 + 10 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 4
)

const (
	DELIVERED     = "DELIVRD" // 成功送达
	UNDELIVERABLE = "UNDELIV" // 无法送达
	EXPIRED       = "EXPIRED"
	DELETED       = "DELETED"
	ACCEPTED      = "ACCEPTD"
	UNKNOWN       = "UNKNOWN"
	REJECTED      = "REJECTD"
)

var (
	ErrInvalidPudLength = errors.New("invalid pdu length")
)
