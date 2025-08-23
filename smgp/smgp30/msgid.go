package smgp30

import (
	"fmt"
	"strconv"
	"time"
)

/*
GenMsgID 生成Deliver或SubmitResp中的MsgID字段.
返回的是长度为20的十六进制值字符串，在Encode时，会转为10字节的二进制.

MsgID 由三部分组成：
- SMGW 代码：3字节（BCD码。编码规则如下：3 位区号（不足前添 0）+2 位设备类别+1 位序号区号：所在省长途区号设备类别：SMGW 取 06 序号：所在省的设备编码，例如第一个网关编号为 1
- 时间：4 字节（BCD 码），格式为 MMDDHHMM（月日时分）
- 序列号：3 字节（BCD 码），取值范围为 000000～999999，从 0 开始，顺序累加，步长为1,循环使用。
例如某 SMGW 的代码为 010061，在 2003年1月16日下午5时0分收到一条短消息，
这条短消息的 MsgID 为：0x01006101161700012345，其中 010061 为 SMGW 代码，01161700 为时间，012345 为序列号。
*/
func GenMsgID(nowFunc func() time.Time, smscgw uint32, seqID uint32) string {
	now := nowFunc()
	part1 := fmt.Sprintf("%06d", smscgw)
	if len(part1) > 6 {
		part1 = part1[len(part1)-6:]
	}
	part2 := fmt.Sprintf("%02d%02d%02d%02d", now.Month(), now.Day(), now.Hour(), now.Minute())
	part3 := fmt.Sprintf("%06d", seqID)
	if len(part3) > 6 {
		part3 = part3[len(part3)-6:]
	}
	return part1 + part2 + part3
}

func ParseMsgID(msgID string) (smgw uint32, timeStr string, seqID uint32, err error) {
	if len(msgID) != 20 {
		return 0, "", 0, fmt.Errorf("invalid msgID length")
	}
	tmp, err := strconv.ParseUint(msgID[:6], 10, 32)
	smgw = uint32(tmp)
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid smgw: %w", err)
	}
	timeStr = msgID[6:14]
	tmp, err = strconv.ParseUint(msgID[14:], 10, 32)
	seqID = uint32(tmp)
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid seqID: %w", err)
	}
	return uint32(smgw), timeStr, uint32(seqID), nil
}
