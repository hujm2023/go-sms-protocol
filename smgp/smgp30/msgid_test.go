package smgp30

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGenMsgID tests the GenMsgID function
func TestGenMsgID(t *testing.T) {

	// Mock the time function to return a fixed time
	nowFunc := func() time.Time {
		return time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC)
	}

	// Test case 1: Normal case
	smgw := uint32(123456)
	seqID := uint32(654321)
	msgID := GenMsgID(nowFunc, smgw, seqID)
	assert.Equal(t, "12345605151430654321", msgID)

	// Test case 2: SMGW code with less than 6 digits
	smgw = uint32(123)
	msgID = GenMsgID(nowFunc, smgw, seqID)
	assert.Equal(t, "00012305151430654321", msgID)

	// Test case 3: SeqID with less than 6 digits
	seqID = uint32(456)
	msgID = GenMsgID(nowFunc, smgw, seqID)
	assert.Equal(t, "00012305151430000456", msgID)

	// Test case 4: SMGW code with more than 6 digits
	smgw = uint32(123456789)
	msgID = GenMsgID(nowFunc, smgw, seqID)
	assert.Equal(t, "45678905151430000456", msgID)

	// Test case 5: SeqID with more than 6 digits
	seqID = uint32(123456789)
	msgID = GenMsgID(nowFunc, smgw, seqID)
	assert.Equal(t, "45678905151430456789", msgID)

}

// TestParseMsgID tests the ParseMsgID function
func TestParseMsgID(t *testing.T) {
	// Test case 1: Normal case
	msgID := "12345605151430654321"
	smgw, timeStr, seqID, err := ParseMsgID(msgID)
	assert.NoError(t, err)
	assert.Equal(t, uint32(123456), smgw)
	assert.Equal(t, "05151430", timeStr)
	assert.Equal(t, uint32(654321), seqID)

	// Test case 2: Invalid length
	msgID = "123456789"
	_, _, _, err = ParseMsgID(msgID)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid msgID length")

	// Test case 3: Invalid SMGW part
	msgID = "abcdef05151430654321"
	_, _, _, err = ParseMsgID(msgID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid smgw")

	// Test case 4: Invalid SeqID part
	msgID = "12345605151430abcdef"
	_, _, _, err = ParseMsgID(msgID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid seqID")

	// Test case 5: SMGW part with leading zeros
	msgID = "00012305151430654321"
	smgw, timeStr, seqID, err = ParseMsgID(msgID)
	assert.NoError(t, err)
	assert.Equal(t, "05151430", timeStr)
	assert.Equal(t, uint32(123), smgw)
	assert.Equal(t, uint32(654321), seqID)
}

func TestXxx(t *testing.T) {
	t.Logf("%06d", 123456789)
}
