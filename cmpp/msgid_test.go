package cmpp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsgID(t *testing.T) {
	month, day, hour, minute, second, gateID, sequenceID := uint64(8), uint64(30), uint64(10), uint64(56), uint64(00), uint64(300), uint64(32768)
	msgID := CombineMsgID(month, day, hour, minute, second, gateID, sequenceID)
	t.Log(msgID)
	a, b, c, d, e, f, g := SplitMsgID(msgID)
	assert.Equal(t, month, a)
	assert.Equal(t, day, b)
	assert.Equal(t, hour, c)
	assert.Equal(t, minute, d)
	assert.Equal(t, second, e)
	assert.Equal(t, gateID, f)
	assert.Equal(t, sequenceID, g)

	msgIDString := MsgID2String(msgID)
	t.Log(msgIDString)
	assert.Equal(t, "0830105600000030032768", msgIDString)
}

func TestXX(t *testing.T) {
	t.Log(MsgID2String(3688792466254149936))
}

func TestSplitMsgID(t *testing.T) {
	t.Log(MsgIDString2Uint64("0901000115001693265292"))
}
