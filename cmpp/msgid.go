package cmpp

import (
	"fmt"
)

/*
MsgID generation for SUBMIT_RESP and DELIVER.
Structure (64-bit integer):
(1) Time (MMDDHHMMSS format): bits 64-39
    - bits 64-61: Month (binary)
    - bits 60-56: Day (binary)
    - bits 55-51: Hour (binary)
    - bits 50-45: Minute (binary)
    - bits 44-39: Second (binary)
(2) SMS Gateway Code: bits 38-17 (integer representation of the gateway code)
(3) Sequence Number: bits 16-1 (sequentially increasing, wraps around)

Left-pad with zeros if necessary, right-aligned.
(SP can get the MsgID of CMPP_Submit message from the consistency of Sequence_Id in request and response messages)
*/

const msgIDFormat = "%02d%02d%02d%02d%02d%07d%05d"

// CombineMsgID generates a 64-bit MsgID based on the provided time components, gateway ID, and sequence ID.
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

// SplitMsgID extracts the time components, gateway ID, and sequence ID from a 64-bit MsgID.
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

// MsgID2String converts a 64-bit MsgID to its string representation (MMDDHHMMSSGGGGGQQQQQ).
func MsgID2String(u uint64) string {
	if u == 0 {
		return ""
	}
	month, day, hour, minute, second, gateID, sequenceID := SplitMsgID(u)
	return fmt.Sprintf(msgIDFormat, month, day, hour, minute, second, gateID, sequenceID)
}

// MsgIDString2Uint64 converts a MsgID string representation (MMDDHHMMSSGGGGGQQQQQ) back to its 64-bit integer form.
func MsgIDString2Uint64(s string) uint64 {
	var month, day, hour, minute, second, gateID, sequenceID uint64
	_, err := fmt.Sscanf(s, msgIDFormat, &month, &day, &hour, &minute, &second, &gateID, &sequenceID)
	if err != nil {
		return 0
	}
	return CombineMsgID(month, day, hour, minute, second, gateID, sequenceID)
}
