package protocol

import (
	"errors"
)

var ErrUnsupportedPacket = errors.New("unsupported packed")

// ICommander 表示 PDU 的command.
type ICommander interface {
	String() string

	ToUint32() uint32
}

type Encoder interface {
	IEncode() ([]byte, error)
}

type Decoder interface {
	IDecode(data []byte) error
}

type EncoderDecoder interface {
	Encoder
	Decoder
}

// PDU stands for Protocol Data Unit, which is the package for standard SMS protocols.
// PDU 表示标准协议的包(所有标准协议pdu都需要实现)
type PDU interface {
	EncoderDecoder

	// SetSequenceID 为 PDU 设置提交时的 seqID
	SetSequenceID(id uint32)

	// GetSequenceID 获取当前PDU的序列号
	GetSequenceID() uint32

	// GetCommand 获取当前 PDU 对应的 ICommand.
	GetCommand() ICommander

	// GenEmptyResponse 生成对应的response，只设置sequenceID字段，其余均为空值.
	// 比如 cmpp20.PduActiveTest -> cmpp20.PduActiveTestResp,
	// 如果一个 PDU 本身就是response，则返回nil.
	GenEmptyResponse() PDU

	// 用于日志打印
	String() string
}
