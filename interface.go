package protocol

import "errors"

var (
	ErrPacketNotComplete = errors.New("packet not completed")
	ErrUnsupportedPacket = errors.New("unsupported packed")
)

// PDU 表示标准协议的包(所有标准协议pdu都需要实现)
type PDU interface {
	// IEncode 序列化
	IEncode() ([]byte, error)

	// IDecode 反序列
	IDecode(data []byte) error

	// SetSequenceID 为 PDU 设置提交时的 seqID
	SetSequenceID(id uint32)
}
