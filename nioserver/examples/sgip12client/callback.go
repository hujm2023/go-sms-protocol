package main

import (
	"context"
	"crypto/subtle"
	"errors"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

type callbackSession struct{ bound, reportSeen bool }

type callbackHandler struct {
	cfg      config
	expected [3]uint32
	result   chan error
}

func (h *callbackHandler) finish(err error) {
	select {
	case h.result <- err:
	default:
	}
}

func (h *callbackHandler) handle(ctx context.Context, pdu protocol.PDU) ([]byte, error) {
	conn, ok := nioserver.GetCtxConn[callbackSession](ctx)
	if !ok {
		return nil, errors.New("missing callback connection")
	}
	return h.handleConnection(ctx, conn, pdu)
}

func (h *callbackHandler) handleConnection(ctx context.Context, conn nioserver.ISMSConn[callbackSession], pdu protocol.PDU) ([]byte, error) {
	if bind, ok := pdu.(*sgip12.Bind); ok {
		return h.bind(ctx, conn, bind)
	}
	session := conn.GetBizData()
	if !session.bound {
		return nil, errors.New("callback command received before bind")
	}
	switch p := pdu.(type) {
	case *sgip12.Report:
		response := p.GenEmptyResponse().(*sgip12.ReportResp)
		if p.SubmitSequence != h.expected || p.UserNumber != h.cfg.destination {
			response.Result = sgip.STAT_ERSEQ
		}
		if p.ReportType != 0 || p.State != sgip.REPORT_OK || p.ErrorCode != 0 {
			response.Result = sgip.STAT_ERARGFMT
		}
		data, err := response.IEncode()
		if err != nil {
			return nil, err
		}
		if response.Result != sgip.STAT_OK {
			err := errors.New("callback report does not match the submitted message or success state")
			return h.reject(ctx, conn, data, err)
		}
		session.reportSeen = true
		conn.SetBizData(session)
		return data, nil
	case *sgip12.Unbind:
		data, err := p.GenEmptyResponse().IEncode()
		if err != nil {
			return nil, err
		}
		err = conn.WriteAndClose(ctx, data)
		if err == nil && !session.reportSeen {
			err = errors.New("callback ended without a report")
		}
		h.finish(err)
		return nil, err
	default:
		return nil, errors.New("unexpected callback command")
	}
}

func (h *callbackHandler) bind(ctx context.Context, conn nioserver.ISMSConn[callbackSession], p *sgip12.Bind) ([]byte, error) {
	response := p.GenEmptyResponse().(*sgip12.BindResp)
	switch {
	case conn.GetBizData().bound:
		response.Result = sgip.STAT_RPTLOGIN
	case p.Type != sgip.SMG_SP:
		response.Result = sgip.STAT_ERLGNTYPE
	case subtle.ConstantTimeCompare([]byte(p.Name), []byte(h.cfg.callbackCredentials.Account)) != 1:
		response.Result = sgip.STAT_ILLLOGIN
	case subtle.ConstantTimeCompare([]byte(p.Password), []byte(h.cfg.callbackCredentials.Password)) != 1:
		response.Result = sgip.STAT_ILLLOGIN
	}
	data, err := response.IEncode()
	if err != nil {
		return nil, err
	}
	if response.Result != sgip.STAT_OK {
		err := errors.New("gateway callback bind rejected")
		return h.reject(ctx, conn, data, err)
	}
	conn.SetBizData(callbackSession{bound: true})
	return data, nil
}

func (h *callbackHandler) reject(
	ctx context.Context,
	conn nioserver.ISMSConn[callbackSession],
	data []byte,
	reason error,
) ([]byte, error) {
	// finish lets runClient shut down the listener. Flush the rejection first;
	// returning response bytes after signalling would race with that shutdown.
	err := errors.Join(reason, conn.WriteAndClose(ctx, data))
	h.finish(err)
	return nil, err
}
