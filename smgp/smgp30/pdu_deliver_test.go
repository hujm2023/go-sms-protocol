package smgp30

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

type DeliverTestSuite struct {
	suite.Suite

	valueBytes []byte
}

func (s *DeliverTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0x67, 0x0, 0x0, 0x0, 0x3, 0x0, 0x0, 0x0, 0x7b, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x1, 0x1, 0xf, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x30, 0x36, 0x39, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x37, 0x35, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0xe, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6d, 0x73, 0x67, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0,
	}
}

func (s *DeliverTestSuite) TestDeliver_IEncode() {
	d := &Deliver{
		Header: smgp.Header{
			CommandID:  smgp.CommandDeliver,
			SequenceID: 123,
		},
		MsgID:      "01020304050607080901",
		DestTermID: "17500000000",
		MsgFormat:  smgp.GB18030,
		SrcTermID:  "1069000000",
		MsgLength:  uint8(len([]byte("hello test msg"))),
		MsgContent: []byte("hello test msg"),
		RecvTime:   "",
		IsReport:   smgp.IS_REPORT,
	}

	data, err := d.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *DeliverTestSuite) TestDeliver_IDecode() {
	deliver := new(Deliver)
	s.Nil(deliver.IDecode(s.valueBytes))

	s.Equal(uint32(123), deliver.Header.SequenceID)
	s.Equal(smgp.CommandDeliver, deliver.Header.CommandID)
	s.Equal("01020304050607080901", deliver.MsgID)
	s.Equal("17500000000", deliver.DestTermID)
	s.Equal(uint8(smgp.GB18030), deliver.MsgFormat)
	s.Equal("1069000000", deliver.SrcTermID)
	s.Equal(uint8(len([]byte("hello test msg"))), deliver.MsgLength)
	s.Equal([]byte("hello test msg"), deliver.MsgContent)
}

func TestDeliver(t *testing.T) {
	suite.Run(t, new(DeliverTestSuite))
}

type DeliverRespTestSuite struct {
	suite.Suite
	valueBytes []byte
}

func (s *DeliverRespTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0x1a, 0x80, 0x0, 0x0, 0x3, 0x0, 0x0, 0x4, 0xd2, 0x1, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x0, 0x0, 0x0, 0x0,
	}
}

func (s *DeliverRespTestSuite) TestDeliverResp_IEncode() {
	d := &DeliverResp{
		Header: smgp.Header{
			CommandID:  smgp.CommandDeliverResp,
			SequenceID: 1234,
		},
		MsgID:  "01010203040506070809",
		Result: smgp.StatOk,
	}
	data, err := d.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *DeliverRespTestSuite) TestDeliverResp_IDecode() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))
	s.Equal(smgp.CommandDeliverResp, d.Header.CommandID)
	s.Equal(uint32(1234), d.Header.SequenceID)
	s.Equal("01010203040506070809", d.MsgID)
	s.Equal(uint32(0), d.Result.Data())
}

func TestDeliverResp(t *testing.T) {
	suite.Run(t, new(DeliverRespTestSuite))
}

func TestExtractDeliveryReceipt(t *testing.T) {
	// 定义测试用例
	testCases := []struct {
		shortMessage []byte
		expected     DeliveryReceipt
	}{
		{
			shortMessage: []byte{105, 100, 58, 1, 96, 23, 18, 41, 20, 36, 16, 0, 7, 32, 115, 117, 98, 58, 48, 48, 49, 32, 100, 108, 118, 114, 100, 58, 48, 48, 49, 32, 83, 117, 98, 109, 105, 116, 32, 100, 97, 116, 101, 58, 50, 51, 49, 50, 50, 57, 49, 52, 50, 52, 32, 100, 111, 110, 101, 32, 100, 97, 116, 101, 58, 50, 51, 49, 50, 50, 57, 49, 52, 50, 52, 32, 115, 116, 97, 116, 58, 82, 84, 58, 48, 49, 52, 56, 32, 101, 114, 114, 58, 49, 52, 56, 32, 84, 101, 120, 116, 58, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expected: DeliveryReceipt{
				ID:       "01601712291424100007",
				DoneDate: "2312291424",
			},
		},
		{
			shortMessage: []byte{105, 100, 58, 1, 96, 19, 1, 3, 32, 56, 18, 5, 104, 32, 115, 117, 98, 58, 48, 48, 49, 32, 100, 108, 118, 114, 100, 58, 48, 48, 49, 32, 83, 117, 98, 109, 105, 116, 32, 100, 97, 116, 101, 58, 50, 52, 48, 49, 48, 51, 50, 48, 51, 56, 32, 100, 111, 110, 101, 32, 100, 97, 116, 101, 58, 50, 52, 48, 49, 48, 51, 50, 48, 51, 56, 32, 115, 116, 97, 116, 58, 68, 69, 76, 73, 86, 82, 68, 32, 101, 114, 114, 58, 48, 48, 48, 32, 84, 101, 120, 116, 58, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expected: DeliveryReceipt{
				ID:       "01601301032038120568",
				DoneDate: "2401032038",
			},
		},
	}

	// 遍历测试用例
	for _, testCase := range testCases {
		receipt, err := ExtractDeliveryReceipt(string(testCase.shortMessage))
		if err != nil {
			t.Errorf("Error extracting delivery receipt: %v", err)
		}

		assert.Equal(t, testCase.expected.ID, receipt.ID)
		assert.Equal(t, testCase.expected.DoneDate, receipt.DoneDate)
	}
}

// hndx用's'分割且filed首字母大写
func TestExtractDeliveryReceiptHndx(t *testing.T) {
	b := []byte{0x69, 0x64, 0x3a, 0x04, 0x70, 0x15, 0x01, 0x16, 0x12, 0x42, 0x68, 0x61, 0x34, 0x73, 0x53, 0x75, 0x62, 0x3a, 0x30, 0x30, 0x31, 0x73, 0x44, 0x6c, 0x76, 0x72, 0x64, 0x3a, 0x30, 0x30, 0x30, 0x73, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x5f, 0x44, 0x61, 0x74, 0x65, 0x3a, 0x32, 0x34, 0x30, 0x31, 0x31, 0x36, 0x31, 0x32, 0x34, 0x32, 0x73, 0x44, 0x6f, 0x6e, 0x65, 0x5f, 0x44, 0x61, 0x74, 0x65, 0x3a, 0x32, 0x34, 0x30, 0x31, 0x31, 0x36, 0x31, 0x32, 0x34, 0x32, 0x73, 0x53, 0x74, 0x61, 0x74, 0x3a, 0x42, 0x57, 0x4c, 0x49, 0x53, 0x54, 0x53, 0x73, 0x45, 0x72, 0x72, 0x3a, 0x31, 0x36, 0x33, 0x73, 0x54, 0x65, 0x78, 0x74, 0x3a, 0x30, 0x30, 0x37, 0x42, 0x57, 0x4c, 0x49, 0x53, 0x54, 0x53, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	receipt, _ := ExtractDeliveryReceipt(string(b))
	assert.Equal(t, receipt.ID, "04701501161242686134")
	assert.Equal(t, receipt.Sub, "001")
	assert.Equal(t, receipt.Dlvrd, "000")
	assert.Equal(t, receipt.SubDate, "2401161242")
	assert.Equal(t, receipt.DoneDate, "2401161242")
	assert.Equal(t, receipt.Stat, "BWLISTS")
	assert.Equal(t, receipt.Err, "163")
	// assert.Equal(t, receipt.Text, "007BWLISTS")
	t.Logf("%+v", receipt)
}

func TestExtractDeliveryReceipt2(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "id:1235 sub:001 Dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:aha"
		d, err := ExtractDeliveryReceipt(s)
		assert.Nil(t, err)
		d.ID = ""
		expectDeliveryReceipt := DeliveryReceipt{
			Sub:      "001",
			Dlvrd:    "1",
			SubDate:  "2107081716",
			DoneDate: "2210131801",
			Stat:     "DELIVRD",
			Err:      "0",
			Text:     "aha",
		}
		assert.Equal(t, expectDeliveryReceipt, d)
		t.Logf("%+v", d)
	})
	t.Run("参数错位", func(t *testing.T) {
		s := "id:1235 sub:001 dlvrd:1 Done_Date:2210131801 submit date:2107081716 stat:DELIVRD err:0 text:aha"
		d, err := ExtractDeliveryReceipt(s)
		assert.Nil(t, err)
		d.ID = ""
		expectDeliveryReceipt := DeliveryReceipt{
			Sub:      "001",
			Dlvrd:    "1",
			SubDate:  "2107081716",
			DoneDate: "2210131801",
			Stat:     "DELIVRD",
			Err:      "0",
			Text:     "aha",
		}
		assert.Equal(t, expectDeliveryReceipt, d)
		t.Logf("%+v", d)
	})
	t.Run("empty source", func(t *testing.T) {
		d, err := ExtractDeliveryReceipt("")
		assert.Nil(t, err)
		assert.Equal(t, DeliveryReceipt{}, d)
	})
	t.Run("demo", func(t *testing.T) {
		d := ""
		dd, err := ExtractDeliveryReceipt(d)
		t.Log(err)
		t.Log(dd)
	})
}

func TestFindSubValue(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		s := "id:1235 sub:001 dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			expectValue string
			maxSize     int
		}{
			{key: "sub", expectValue: "001", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, "", item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("顺序不对", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 sub:001 dlvrd:1 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "sub", expectValue: "001", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, "", item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("少一个key", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 dlvrd:1 done date:2210131801 stat:DELIVRD err:0 text:"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "sub", expectValue: "", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, "", item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
	t.Run("中间的 key 为空", func(t *testing.T) {
		s := "id:1235 submit date:2107081716 sub: dlvrd:1 done date:2210131801 stat:DELIVRD err:0"
		for _, item := range []struct {
			key         string
			maxSize     int
			expectValue string
		}{
			{key: "sub", expectValue: "", maxSize: 3},
			{key: "dlvrd", expectValue: "1", maxSize: 3},
			{key: "submit date", expectValue: "2107081716", maxSize: 10},
			{key: "done date", expectValue: "2210131801", maxSize: 10},
			{key: "stat", expectValue: "DELIVRD", maxSize: 7},
			{key: "err", expectValue: "0", maxSize: 3},
			{key: "text", expectValue: "", maxSize: 20},
		} {
			v := findSubValue(s, item.key, "", item.maxSize)
			assert.Equal(t, item.expectValue, v)
		}
	})
}

type DeliverContentTestSuite struct {
	suite.Suite
	msgContentString string
	msgContent       []byte
}

func (s *DeliverContentTestSuite) SetupTest() {
	s.msgContentString = `id:"!2Bys sub:001 dlvrd:001 submit date:2508222132 done date:2508222132 stat:DELIVRD err:000 text:`
	s.msgContent = []byte{105, 100, 58, 0, 0, 99, 8, 34, 33, 50, 66, 121, 115, 32, 115, 117, 98, 58, 48, 48, 49, 32, 100, 108, 118, 114, 100, 58, 48, 48, 49, 32, 115, 117, 98, 109, 105, 116, 32, 100, 97, 116, 101, 58, 50, 53, 48, 56, 50, 50, 50, 49, 51, 50, 32, 100, 111, 110, 101, 32, 100, 97, 116, 101, 58, 50, 53, 48, 56, 50, 50, 50, 49, 51, 50, 32, 115, 116, 97, 116, 58, 68, 69, 76, 73, 86, 82, 68, 32, 101, 114, 114, 58, 48, 48, 48, 32, 116, 101, 120, 116, 58, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (s *DeliverContentTestSuite) TestDeliverContent_IEncode() {
	d := &DeliveryReceipt{
		ID:       "00006308222132427973",
		Sub:      "001",
		Dlvrd:    "001",
		SubDate:  "2508222132",
		DoneDate: "2508222132",
		Stat:     "DELIVRD",
		Err:      "000",
		Text:     "",
	}
	b, err := d.IEncode()
	assert.NoError(s.T(), err)
	// s.msgContent 末尾的0不应该计算在内
	shouldEqual := make([]byte, len(s.msgContent))
	copy(shouldEqual, s.msgContent)
	for i := len(shouldEqual) - 1; i >= 0; i-- {
		if shouldEqual[i] != 0 {
			shouldEqual = shouldEqual[:i+1]
			break
		}
	}
	assert.Equal(s.T(), shouldEqual, b)
}

func (s *DeliverContentTestSuite) TestDeliverContent_IDecode() {
	d := new(DeliveryReceipt)
	assert.Nil(s.T(), d.IDecode(s.msgContent))
	assert.Equal(s.T(), "00006308222132427973", d.ID)
	assert.Equal(s.T(), "001", d.Sub)
	assert.Equal(s.T(), "001", d.Dlvrd)
	assert.Equal(s.T(), "2508222132", d.SubDate)
	assert.Equal(s.T(), "2508222132", d.DoneDate)
	assert.Equal(s.T(), "DELIVRD", d.Stat)
	assert.Equal(s.T(), "000", d.Err)
	assert.Equal(s.T(), "", d.Text)
}

func TestDeliverContentTestSuite(t *testing.T) {
	suite.Run(t, new(DeliverContentTestSuite))
}
