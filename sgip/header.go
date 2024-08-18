package sgip

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/hujm2023/go-sms-protocol/packet"
)

const (
	HeaderLength = 20

	PacketTotalLengthBytes = 4 // Message Length 消息的总长度(字节)

	MaxHeaderRespLength = 29 // 通用的响应长度  20 + 1 + 8
)

type Header struct {
	// sgip 协议比较特殊，sequenceId是一个数组组成，理论上只有Sequence[2]是seqId，同时响应与回执也是靠用户侧提交的seqId作为串联
	// 整体长度 4 + 4 + 12 = 20
	// 4 字节 Message Length 消息的总长度(字节)
	TotalLength uint32
	// 4 字节 Command ID 命令ID
	CommandID CommandID
	//  sgip 专属，4字节 * 3 = 12 Sequence Number 序列号 {cid, sid, seq}
	Sequence [3]uint32
}

func (h Header) String() string {
	return fmt.Sprintf("{TotalLength:%d, CommandID:%s, Sequence:%v}", h.TotalLength, h.CommandID, h.Sequence)
}

func NewHeader(totalLength uint32, commandID CommandID, nodeID, sequenceID uint32) Header {
	return Header{TotalLength: totalLength, CommandID: commandID, Sequence: [3]uint32{nodeID, Timestamp(time.Now()), sequenceID}}
}

func ReadHeader(r *packet.Reader) Header {
	h := Header{}
	h.TotalLength = r.ReadUint32()
	h.CommandID = CommandID(r.ReadUint32())
	h.Sequence = [3]uint32{r.ReadUint32(), r.ReadUint32(), r.ReadUint32()}
	return h
}

func WriteHeaderNoLength(h Header, buf *packet.Writer) {
	buf.WriteUint32(uint32(h.CommandID))
	buf.WriteUint32(h.Sequence[0])
	buf.WriteUint32(h.Sequence[1])
	buf.WriteUint32(h.Sequence[2])
}

// PeekHeader 尝试读取前 HeaderLength 长度的字节并解析成 Header， 不影响原有 reader 的游标
func PeekHeader(buf []byte) (h Header, err error) {
	if len(buf) < MinSGIPPduLength {
		return h, ErrInvalidPudLength
	}
	h.TotalLength = binary.BigEndian.Uint32(buf[:4])
	h.CommandID = CommandID(binary.BigEndian.Uint32(buf[4:8]))
	h.Sequence = [3]uint32{
		binary.BigEndian.Uint32(buf[8:12]),
		binary.BigEndian.Uint32(buf[12:16]),
		binary.BigEndian.Uint32(buf[16:20]),
	}
	return h, nil
}

// GetSequenceID 获取SGIP的SequenceID
func (p *Header) GetSequenceID() uint32 {
	return p.Sequence[2]
}

// GetMsgId 获取SGIP的MsgId
func (p *Header) GetMsgId() string {
	return strconv.FormatUint(uint64(p.Sequence[2]), 10)
}
