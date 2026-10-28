package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	protocol "github.com/hujm2023/go-sms-protocol"
	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

const maxFrameSize = 64 * 1024

type connectionSession struct{ bound bool }

type receiptJob struct {
	submitSequence [3]uint32
	recipients     []string
}

type serverHandler struct {
	cfg      config
	logger   smslog.FullLogger
	receipts *exampleutil.Queue[receiptJob]
	sequence atomic.Uint32
}

func (h *serverHandler) header(command sgip.CommandID) sgip.Header {
	sequence := h.sequence.Add(1)
	if sequence == 0 {
		sequence = h.sequence.Add(1)
	}
	return sgip.NewHeader(0, command, h.cfg.nodeID, sequence)
}

func (h *serverHandler) handle(ctx context.Context, pdu protocol.PDU) ([]byte, error) {
	conn, ok := nioserver.GetCtxConn[connectionSession](ctx)
	if !ok {
		return nil, errors.New("missing connection")
	}
	if bind, ok := pdu.(*sgip12.Bind); ok {
		return h.bind(ctx, conn, bind)
	}
	if !conn.GetBizData().bound {
		return nil, errors.New("PDU received before bind")
	}
	switch p := pdu.(type) {
	case *sgip12.Submit:
		return h.submit(ctx, p)
	case *sgip12.Unbind:
		data, err := p.GenEmptyResponse().IEncode()
		if err != nil {
			return nil, err
		}
		return nil, conn.WriteAndClose(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported submission-connection command: %s", pdu.GetCommand())
	}
}

func (h *serverHandler) bind(ctx context.Context, conn nioserver.ISMSConn[connectionSession], p *sgip12.Bind) ([]byte, error) {
	response := p.GenEmptyResponse().(*sgip12.BindResp)
	switch {
	case conn.GetBizData().bound:
		response.Result = sgip.STAT_RPTLOGIN
	case p.Type != sgip.SP_SMG:
		response.Result = sgip.STAT_ERLGNTYPE
	case subtle.ConstantTimeCompare([]byte(p.Name), []byte(h.cfg.Account)) != 1:
		response.Result = sgip.STAT_ILLLOGIN
	case subtle.ConstantTimeCompare([]byte(p.Password), []byte(h.cfg.Password)) != 1:
		response.Result = sgip.STAT_ILLLOGIN
	}
	data, err := response.IEncode()
	if err != nil {
		return nil, err
	}
	if response.Result != sgip.STAT_OK {
		return data, errors.New("bind rejected")
	}
	conn.SetBizData(connectionSession{bound: true})
	h.logger.CtxInfof(ctx, "sgip bind accepted remote=%s type=1", conn.RemoteAddr())
	return data, nil
}

func validateSubmit(p *sgip12.Submit) sgip.RespStatus {
	if p.UserCount == 0 || p.UserCount > 100 || int(p.UserCount) != len(p.UserNumber) {
		return sgip.STAT_ERARGFMT
	}
	if p.ReportFlag > 2 || p.Priority > 9 || p.MessageType != 0 || p.SpNumber == "" {
		return sgip.STAT_ERARGFMT
	}
	if p.MessageLength != uint32(len(p.MessageContent)) || len(p.MessageContent) > 140 {
		return sgip.STAT_ERLEN
	}
	for _, number := range p.UserNumber {
		if !strings.HasPrefix(number, "86") || len(number) <= 2 {
			return sgip.STAT_ILLUSRNUM
		}
	}
	return sgip.STAT_OK
}

func (h *serverHandler) submit(ctx context.Context, p *sgip12.Submit) ([]byte, error) {
	response := p.GenEmptyResponse().(*sgip12.SubmitResp)
	response.Result = validateSubmit(p)
	if response.Result != sgip.STAT_OK {
		return response.IEncode()
	}
	if p.ReportFlag == 1 {
		job := receiptJob{submitSequence: p.Sequence, recipients: append([]string(nil), p.UserNumber...)}
		if !h.receipts.TryEnqueue(job) {
			response.Result = sgip.STAT_NODEBUSY
			return response.IEncode()
		}
	}
	h.logger.CtxInfof(ctx, "submit accepted submit_id=%s recipients=%d content_bytes=%d report_flag=%d",
		sgip.SequenceIDString(p.Sequence), len(p.UserNumber), len(p.MessageContent), p.ReportFlag)
	return response.IEncode()
}

// expectResponse checks all three sequence components, never only GetSequenceID().
func expectResponse(conn io.ReadWriter, request protocol.PDU) error {
	if err := exampleutil.Write(conn, request); err != nil {
		return err
	}
	pdu, err := exampleutil.Read(conn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
	if err != nil {
		return err
	}
	var got, want [3]uint32
	var status sgip.RespStatus
	switch p := request.(type) {
	case *sgip12.Bind:
		response, ok := pdu.(*sgip12.BindResp)
		if !ok {
			return errors.New("expected callback BIND_RESP")
		}
		got, want, status = response.Sequence, p.Sequence, response.Result
	case *sgip12.Report:
		response, ok := pdu.(*sgip12.ReportResp)
		if !ok {
			return errors.New("expected REPORT_RESP")
		}
		got, want, status = response.Sequence, p.Sequence, response.Result
	case *sgip12.Unbind:
		response, ok := pdu.(*sgip12.UnbindResp)
		if !ok {
			return errors.New("expected callback UNBIND_RESP")
		}
		got, want = response.Sequence, p.Sequence
	default:
		return errors.New("unsupported callback request")
	}
	if got != want || status != sgip.STAT_OK {
		return fmt.Errorf("callback response mismatch or rejection: sequence=%v status=%d", got, status)
	}
	return nil
}

func (h *serverHandler) pushReports(ctx context.Context, job receiptJob) error {
	conn, cleanup, err := exampleutil.Dial(ctx, h.cfg.callbackAddress)
	if err != nil {
		return err
	}
	defer cleanup()
	bind := &sgip12.Bind{
		Header: h.header(sgip.SGIP_BIND), Type: sgip.SMG_SP,
		Name: h.cfg.callbackCredentials.Account, Password: h.cfg.callbackCredentials.Password,
	}
	if err := expectResponse(conn, bind); err != nil {
		return fmt.Errorf("callback bind: %w", err)
	}
	for _, number := range job.recipients {
		report := &sgip12.Report{
			Header: h.header(sgip.SGIP_REPORT), SubmitSequence: job.submitSequence,
			ReportType: 0, UserNumber: number, State: sgip.REPORT_OK,
		}
		if err := expectResponse(conn, report); err != nil {
			return fmt.Errorf("report acknowledgement: %w", err)
		}
	}
	return expectResponse(conn, &sgip12.Unbind{Header: h.header(sgip.SGIP_UNBIND)})
}

func (h *serverHandler) deliver(ctx context.Context, job receiptJob) {
	if !exampleutil.Delay(ctx, h.cfg.MinDelay, h.cfg.MaxDelay) {
		return
	}
	callbackCtx, cancel := context.WithTimeout(ctx, h.cfg.callbackTimeout)
	defer cancel()
	if err := h.pushReports(callbackCtx, job); err != nil {
		if ctx.Err() == nil {
			h.logger.Errorf("sgip report callback submit_id=%s: %v", sgip.SequenceIDString(job.submitSequence), err)
		}
		return
	}
	h.logger.Infof("reports acknowledged submit_id=%s recipients=%d", sgip.SequenceIDString(job.submitSequence), len(job.recipients))
}
