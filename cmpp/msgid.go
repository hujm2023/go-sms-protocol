package cmpp

import (
	"fmt"
)

/*
SUBMIT_RESP 和 DELIVERY 中 Msg_Id 生成。组成如下:
采用 64 位(8 字节)的整数:
		(1)时间(格式为 MMDDHHMMSS，即 月日时分秒)
			:bit64~bit39，其中
				bit64~bit61:月份的二进制表示;
				bit60~bit56:日的二进制表示;
				bit55~bit51:小时的二进制表示;
				bit50~bit45:分的二进制表示;
				bit44~bit39:秒的二进制表示;
		(2)短信网关代码:bit38~bit17，把短信网关的代码转换为整数填写到该字段中。
		(3)序列号:bit16~bit1，顺序增加，步长为 1，循环使用。

各部分如不能填满，左补零，右对齐。
(SP 根据请求和应答消息的 Sequence_Id 一致性就可得到 CMPP_Submit 消息的 Msg_Id)
*/

const msgIDFormat = "%02d%02d%02d%02d%02d%07d%05d"

// CombineMsgID 生成Msg_Id
func CombineMsgID(month, day, hour, minute, second, gateID, sequenceID uint64) uint64 {
	var msgID uint64
	msgID = month
	msgID = msgID<<5 + day
	msgID = msgID<<5 + hour
	msgID = msgID<<6 + minute
	msgID = msgID<<6 + second
	msgID = msgID<<22 + gateID
	msgID = msgID<<16 + uint64(uint16(sequenceID)) // sequenceID最多2字节(16位,65535)，超出直接夹断
	return msgID
}

// SplitMsgID 从Msg_Id中解析出发送信息
func SplitMsgID(msgID uint64) (month, day, hour, minute, second, gateID, sequenceID uint64) {
	month = msgID >> 60 & 0xf
	day = msgID >> 55 & 0x1f
	hour = msgID >> 50 & 0x1f
	minute = msgID >> 44 & 0x3f
	second = msgID >> 38 & 0x3f
	gateID = msgID >> 16 & 0x3fffff
	sequenceID = msgID & 0xffff
	return
}

// MsgID2String 将Msg_Id转为年月日表示法
func MsgID2String(u uint64) string {
	if u == 0 {
		return ""
	}
	month, day, hour, minute, second, gateID, sequenceID := SplitMsgID(u)
	return fmt.Sprintf(msgIDFormat, month, day, hour, minute, second, gateID, sequenceID)
}

// MsgIDString2Uint64 ...
func MsgIDString2Uint64(s string) uint64 {
	var month, day, hour, minute, second, gateID, sequenceID uint64
	_, err := fmt.Sscanf(s, msgIDFormat, &month, &day, &hour, &minute, &second, &gateID, &sequenceID)
	if err != nil {
		return 0
	}
	return CombineMsgID(month, day, hour, minute, second, gateID, sequenceID)
}
