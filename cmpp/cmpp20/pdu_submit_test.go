package cmpp20

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func TestPduSubmit(t *testing.T) {
	b := []byte{
		0, 0, 0, 176, 0, 0, 0, 4, 0, 0, 0, 23, 0, 0, 0, 0,
		0, 0, 0, 0, 1, 1, 1, 1, 116, 101, 115, 116, 0, 0, 0, 0,
		0, 0, 2, 49, 51, 53, 48, 48, 48, 48, 50, 54, 57, 54, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 57, 48, 48, 48, 48,
		49, 48, 50, 49, 48, 0, 0, 0, 0, 49, 53, 49, 49, 48, 53, 49,
		51, 49, 53, 53, 53, 49, 48, 49, 43, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 57, 48, 48, 48, 48,
		49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 49, 51, 53, 48, 48, 48, 48, 50, 54, 57, 54, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 17, 103, 111, 32, 115, 117, 98, 109, 105, 116,
		32, 99, 111, 110, 116, 101, 110, 116, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	s := new(PduSubmit)
	assert.Nil(t, s.IDecode(b))
	assert.Equal(t, uint32(176), s.TotalLength)
	assert.Equal(t, cmpp.CommandSubmit, s.CommandID)
	assert.Equal(t, uint32(23), s.SequenceID)

	assert.Equal(t, uint8(1), s.PkTotal)
	assert.Equal(t, uint8(1), s.PkNumber)
	assert.Equal(t, uint8(1), s.RegisteredDelivery)
	assert.Equal(t, uint8(1), s.MsgLevel)
	assert.Equal(t, "test", s.ServiceID)
	assert.Equal(t, "02", s.FeeType)
	assert.Equal(t, "10", s.FeeCode)
	assert.Equal(t, uint8(2), s.FeeUserType)
	assert.Equal(t, uint8(0), s.MsgFmt)
	assert.Equal(t, "900001", s.MsgSrc)
	assert.Equal(t, "13500002696", s.FeeTerminalID)
	assert.Equal(t, "", s.AtTime)
	assert.Equal(t, "900001", s.SrcID)
	assert.Equal(t, uint8(1), s.DestUsrTL)
	assert.Equal(t, int(1), len(s.DestTerminalID))
	assert.Equal(t, "13500002696", s.DestTerminalID[0])
	assert.Equal(t, uint8(17), s.MsgLength)
	assert.Equal(t, "go submit content", s.MsgContent)
	assert.Equal(t, "", s.Reserve)

	encoded, err := s.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(b, encoded))

	t.Log(s.String())
}

func TestA(t *testing.T) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, 1195725856)
	t.Log(b)
}

func TestHeaderString(t *testing.T) {
	h := cmpp.Header{
		TotalLength: 39,
		CommandID:   cmpp.CommandActiveTestResp,
		SequenceID:  123,
	}

	submit := PduSubmit{
		Header:             h,
		PkTotal:            1,
		PkNumber:           1,
		RegisteredDelivery: 1,
		MsgLevel:           1,
		ServiceID:          "test",
		FeeType:            "02",
	}

	ss, _ := json.Marshal(submit)
	var sss PduSubmit
	lo.Must0(json.Unmarshal(ss, &sss))
	t.Log(sss)
}
