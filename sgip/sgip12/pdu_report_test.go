package sgip12

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/sgip"
)

func TestReport(t *testing.T) {
	raw := []byte{
		0, 0, 0, 64, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0,
		49, 55, 54, 49, 49, 48, 48, 48, 48, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 9, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	// 实现接口级别的单测
	a := new(Report)
	assert.Nil(t, a.IDecode(raw))

	report := &Report{
		Header: sgip.Header{
			CommandID: sgip.SGIP_REPORT,
			Sequence:  [3]uint32{0, 0, 0},
		},
		SubmitSequence: [3]uint32{0, 0, 2},
		ReportType:     0,
		UserNumber:     "17611000000",
		State:          2,
		ErrorCode:      9,
		Reserved:       "",
	}
	assert.Equal(t, "9", strconv.Itoa(int(report.ErrorCode)))
	value, err := report.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))
	assert.Equal(t, uint32(2), report.GetSubmitId())
	assert.Equal(t, "2", report.GetSubmitIdStr())
	assert.Equal(t, sgip.SGIP_REPORT, report.GetCommand())

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*Report)
	assert.True(t, ok)

	reportResp := report.GenEmptyResponse()
	assert.Equal(t, sgip.SGIP_REPORT_REP, reportResp.GetCommand())
	assert.Nil(t, reportResp.GenEmptyResponse())
}

func TestReportResp(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x1d,
		0x80, 0x0, 0x0, 0x5,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0,
	}

	// 实现接口级别的单测
	a := new(ReportResp)
	assert.Nil(t, a.IDecode(raw))

	encoded, err := a.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, encoded))
	response := &ReportResp{
		Header: sgip.Header{
			CommandID: sgip.SGIP_REPORT_REP,
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
	_, ok := pdu.(*ReportResp)
	assert.True(t, ok)
}
