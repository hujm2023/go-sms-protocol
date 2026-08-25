package sgip12

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/sgip"
)

func TestSubmit(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0xc1, 0x0, 0x0, 0x0, 0x3,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x31, 0x30, 0x36, 0x39,
		0x30, 0x30, 0x39, 0x30, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x31, 0x37, 0x36, 0x30, 0x30, 0x35, 0x33,
		0x37, 0x33, 0x30, 0x30, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x31,
		0x37, 0x36, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
		0x30, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x31, 0x37, 0x36, 0x31,
		0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x31, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x31, 0x32, 0x33, 0x34, 0x35, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x79, 0x79, 0x6d, 0x6d, 0x64, 0x64, 0x68, 0x68,
		0x6d, 0x6d, 0x73, 0x73, 0x74, 0x6e, 0x6e, 0x70,
		0x79, 0x79, 0x6d, 0x6d, 0x64, 0x64, 0x68, 0x68,
		0x6d, 0x6d, 0x73, 0x73, 0x74, 0x6e, 0x6e, 0x70,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x8, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6d, 0x73,
		0x67, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0,
	}
	// 实现接口级别的单测
	a := new(Submit)
	assert.Nil(t, a.IDecode(raw))
	submit := Submit{
		Header: sgip.Header{
			TotalLength: 193,
			CommandID:   sgip.SGIP_SUBMIT,
			Sequence:    [3]uint32{},
		},
		SpNumber:         "10690090",
		ChargeNumber:     "17600537300",
		UserCount:        2,
		UserNumber:       []string{"17600000000", "17611111111"},
		CorpID:           "12345",
		ServiceType:      "",
		FeeType:          0,
		FeeValue:         "",
		GivenValue:       "",
		AgentFlag:        0,
		MorelatetoMTFlag: 0,
		Priority:         0,
		ExpireTime:       "yymmddhhmmsstnnp",
		ScheduleTime:     "yymmddhhmmsstnnp",
		ReportFlag:       0,
		TpPid:            0,
		TpUdhi:           0,
		MessageCoding:    0,
		MessageType:      0,
		MessageLength:    8,
		MessageContent:   []byte("test msg"),
		Reserved:         "",
	}

	value, err := submit.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))
	assert.Equal(t, sgip.SGIP_SUBMIT, submit.GetCommand())

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*Submit)
	assert.True(t, ok)

	submitResp := submit.GenEmptyResponse()
	assert.Equal(t, sgip.SGIP_SUBMIT_REP, submitResp.GetCommand())
	assert.Nil(t, submitResp.GenEmptyResponse())
}

func TestSubmitResp(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x1d,
		0x80, 0x0, 0x0, 0x3,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0,
	}

	// 实现接口级别的单测
	a := new(SubmitResp)
	assert.Nil(t, a.IDecode(raw))

	encoded, err := a.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, encoded))
	response := &SubmitResp{
		Header: sgip.Header{
			CommandID: sgip.SGIP_SUBMIT_REP,
			Sequence:  [3]uint32{0, 0, 0},
		},
		Result:   sgip.STAT_OK,
		Reserved: "",
	}

	value, err := response.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*SubmitResp)
	assert.True(t, ok)
}

func TestSubmit_1(t *testing.T) {
	raw := []byte{
		0, 0, 0, 192, 0, 0, 0, 3, 0, 0, 0, 0, 31, 1,
		241, 15, 0, 0, 0, 98, 49, 48, 54, 57, 48, 50,
		54, 57, 51, 49, 50, 48, 51, 0, 0, 0, 0, 0, 0,
		0, 0, 57, 51, 49, 50, 48, 51, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 49, 56, 48, 50, 54,
		57, 48, 49, 48, 50, 52, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 49, 50, 51, 52, 53, 57, 51, 49, 50, 48, 51, 0, 0, 0, 0, 1, 48, 48, 48, 0, 0, 0, 48, 48, 48, 48, 48, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 8, 0, 0, 0, 0, 28, 0, 116, 0, 101, 0, 115, 0, 116, 0, 115, 0, 109, 0, 115, 0, 115, 0, 103, 0, 105, 0, 112, 0, 49, 0, 50, 0, 51, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	// 实现接口级别的单测
	a := new(Submit)
	assert.Nil(t, a.IDecode(raw))
	assert.Equal(t, []byte{
		0, 't', 0, 'e', 0, 's', 0, 't', 0, 's', 0, 'm', 0, 's',
		0, 's', 0, 'g', 0, 'i', 0, 'p', 0, '1', 0, '2', 0, '3',
	}, a.MessageContent)
}

func submitWireFixture() []byte {
	var body bytes.Buffer
	writeFixed := func(value string, width int) {
		field := make([]byte, width)
		copy(field, value)
		body.Write(field)
	}
	writeFixed("10690090", 21)
	writeFixed("", 21)
	body.WriteByte(1)
	writeFixed("8613800138000", 21)
	writeFixed("12345", 5)
	writeFixed("svc", 10)
	body.WriteByte(0)
	writeFixed("", 6)
	writeFixed("", 6)
	body.WriteByte(0)
	body.WriteByte(0)
	body.WriteByte(1)
	writeFixed("", 16)
	writeFixed("", 16)
	body.WriteByte(1)
	body.WriteByte(0)
	body.WriteByte(0)
	body.WriteByte(0)
	body.WriteByte(0)
	var messageLength [4]byte
	binary.BigEndian.PutUint32(messageLength[:], 3)
	body.Write(messageLength[:])
	body.WriteString("abc")
	writeFixed("", 8)

	data := make([]byte, sgip.HeaderLength, sgip.HeaderLength+body.Len())
	binary.BigEndian.PutUint32(data[0:4], uint32(len(data)+body.Len()))
	binary.BigEndian.PutUint32(data[4:8], uint32(sgip.SGIP_SUBMIT))
	binary.BigEndian.PutUint32(data[8:12], 11)
	binary.BigEndian.PutUint32(data[12:16], 22)
	binary.BigEndian.PutUint32(data[16:20], 33)
	return append(data, body.Bytes()...)
}

func TestSubmit_DecodeIsReusableAndRejectsLengthAnomalies(t *testing.T) {
	raw := submitWireFixture()
	var submit Submit
	for i := 0; i < 2; i++ {
		if err := submit.IDecode(raw); err != nil {
			t.Fatalf("IDecode() pass %d error = %v", i+1, err)
		}
		if len(submit.UserNumber) != 1 || submit.UserNumber[0] != "8613800138000" {
			t.Fatalf("decode pass %d UserNumber = %#v", i+1, submit.UserNumber)
		}
	}

	tests := []struct {
		name   string
		mutate func([]byte)
	}{
		{
			name: "user count exceeds encoded numbers",
			mutate: func(data []byte) {
				data[sgip.HeaderLength+21+21] = 2
			},
		},
		{
			name: "message length is too small",
			mutate: func(data []byte) {
				binary.BigEndian.PutUint32(data[152:156], 2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := append([]byte(nil), raw...)
			tt.mutate(data)
			if err := new(Submit).IDecode(data); err == nil {
				t.Fatal("IDecode() accepted malformed Submit")
			}
		})
	}
	if _, err := (&Submit{MessageLength: 2, MessageContent: []byte("x")}).IEncode(); err == nil {
		t.Fatal("IEncode() accepted mismatched MessageLength")
	}
}

func TestSubmit_GenEmptyResponsePreservesSequence(t *testing.T) {
	submit := &Submit{
		Header: sgip.Header{
			CommandID: sgip.SGIP_SUBMIT,
			Sequence:  [3]uint32{11, 22, 33},
		},
	}
	response, ok := submit.GenEmptyResponse().(*SubmitResp)
	if !ok {
		t.Fatalf("GenEmptyResponse() type = %T, want *SubmitResp", submit.GenEmptyResponse())
	}
	if response.CommandID != sgip.SGIP_SUBMIT_REP || response.Sequence != submit.Sequence {
		t.Fatalf("response header = %#v, want command %v and sequence %#v", response.Header, sgip.SGIP_SUBMIT_REP, submit.Sequence)
	}
	encoded, err := response.IEncode()
	if err != nil {
		t.Fatalf("response IEncode() error = %v", err)
	}
	pdu, err := DecodeSGIP12(encoded)
	if err != nil {
		t.Fatalf("DecodeSGIP12() error = %v", err)
	}
	decoded, ok := pdu.(*SubmitResp)
	if !ok || decoded.Sequence != submit.Sequence {
		t.Fatalf("decoded response = %T %#v", pdu, pdu)
	}
}
