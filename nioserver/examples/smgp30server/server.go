package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/hujm2023/hlog"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

const maxSMGP30FrameSize uint32 = 64 * 1024

var errLoginRejected = errors.New("smgp30server: login rejected")

type connectionSession struct {
	account         string
	loginMode       uint8
	isAuthenticated bool
}

type serverHandler struct {
	account    string
	password   string
	authSkew   time.Duration
	logger     hlog.FullLogger
	receipts   *receiptDispatcher
	heartbeats *heartbeatManager
	messageIDs messageIDGenerator
	now        func() time.Time
}

type messageIDGenerator struct {
	gatewayID uint32
	sequence  atomic.Uint32
}

func (g *messageIDGenerator) next(now time.Time) string {
	// Six BCD digits allow one million IDs per minute in this single process.
	sequence := g.sequence.Add(1) % 1000000
	return smgp30.GenMsgID(func() time.Time { return now }, g.gatewayID, sequence)
}

func newUnpackFunc(maxFrameSize uint32) nioserver.UnpackFunc {
	return func(_ context.Context, reader netpoll.Reader) (pdu protocol.PDU, err error) {
		defer func() { err = errors.Join(err, reader.Release()) }()
		lengthBytes, err := reader.Peek(smgp.PacketTotalLengthBytes)
		if err != nil {
			return nil, fmt.Errorf("read smgp pdu length: %w", err)
		}
		totalLength := binary.BigEndian.Uint32(lengthBytes)
		if totalLength < smgp.HeaderLength || totalLength > maxFrameSize {
			return nil, fmt.Errorf("invalid smgp pdu length: %d", totalLength)
		}
		// ReadBinary waits for a complete frame, including fragmented TCP input.
		frame, err := reader.ReadBinary(int(totalLength))
		if err != nil {
			return nil, fmt.Errorf("read smgp pdu: %w", err)
		}
		pdu, err = smgp30.DecodeSMGP30(frame)
		if err != nil {
			return nil, fmt.Errorf("decode smgp 3.0 pdu: %w", err)
		}
		canonical, err := pdu.IEncode()
		if err != nil {
			return nil, fmt.Errorf("validate smgp 3.0 pdu: %w", err)
		}
		if len(canonical) != len(frame) {
			return nil, fmt.Errorf("invalid smgp pdu size: decoded %d bytes from %d-byte frame", len(canonical), len(frame))
		}
		return pdu, nil
	}
}

func (h *serverHandler) handle(ctx context.Context, pdu protocol.PDU) ([]byte, error) {
	conn, ok := nioserver.GetCtxConn[connectionSession](ctx)
	if !ok {
		return nil, errors.New("smgp30server: connection missing from context")
	}
	if pdu == nil {
		return nil, errors.New("smgp30server: decoded pdu is nil")
	}
	if request, ok := pdu.(*smgp30.Login); ok {
		return h.handleLogin(ctx, conn, request)
	}
	if !conn.GetBizData().isAuthenticated {
		return nil, errors.New("smgp30server: request received before authentication")
	}

	switch request := pdu.(type) {
	case *smgp30.Submit:
		return h.handleSubmit(ctx, conn, request)
	case *smgp30.ActiveTest:
		return request.GenEmptyResponse().IEncode()
	case *smgp30.ActiveTestResp:
		conn.OnReceiveActiveTest()
		return nil, nil
	case *smgp30.DeliverResp:
		h.logger.CtxInfof(
			ctx,
			"delivery receipt acknowledged remote=%s delivery_msg_id=%s sequence=%d result=%d",
			conn.RemoteAddr(), request.MsgID, request.GetSequenceID(), request.Result,
		)
		return nil, nil
	case *smgp30.Exit:
		data, err := request.GenEmptyResponse().IEncode()
		if err != nil {
			return nil, fmt.Errorf("encode smgp exit response: %w", err)
		}
		return nil, conn.WriteAndClose(ctx, data)
	default:
		return nil, fmt.Errorf("smgp30server: unsupported command %s", pdu.GetCommand())
	}
}

func (h *serverHandler) handleLogin(
	ctx context.Context,
	conn nioserver.ISMSConn[connectionSession],
	request *smgp30.Login,
) ([]byte, error) {
	status := statusStructure
	if !conn.GetBizData().isAuthenticated {
		cfg := config{account: h.account, password: h.password, authSkew: h.authSkew}
		status = authenticateLogin(request, cfg, h.now())
	}
	data, err := newLoginResponse(request, status, h.password).IEncode()
	if err != nil {
		return nil, fmt.Errorf("encode smgp login response: %w", err)
	}
	if status != statusOK {
		// BaseServer flushes response bytes before closing on a handler error.
		return data, errLoginRejected
	}
	conn.SetBizData(connectionSession{
		account: request.ClientID, loginMode: request.LoginMode, isAuthenticated: true,
	})
	h.heartbeats.register(conn)
	h.logger.CtxInfof(
		ctx,
		"smgp connection authenticated remote=%s account=%q login_mode=%d",
		conn.RemoteAddr(), request.ClientID, request.LoginMode,
	)
	return data, nil
}

func (h *serverHandler) handleSubmit(
	ctx context.Context,
	conn nioserver.ISMSConn[connectionSession],
	request *smgp30.Submit,
) ([]byte, error) {
	response := request.GenEmptyResponse().(*smgp30.SubmitResp)
	// Even failed responses carry the fixed-width ten-byte MsgID field.
	response.MsgID = "00000000000000000000"
	response.Status = validateSubmit(request)
	mode := conn.GetBizData().loginMode
	if mode == smgp.RECEIVE_MODE || (mode == smgp.SEND_MODE && request.NeedReport == smgp.NEED_REPORT) {
		// This example routes receipts to the submitting connection only.
		response.Status = statusStructure
	}
	if response.Status != statusOK {
		return response.IEncode()
	}

	submittedAt := h.now()
	response.MsgID = h.messageIDs.next(submittedAt)
	if request.NeedReport == smgp.NEED_REPORT {
		jobs := make([]receiptJob, 0, len(request.DestTermID))
		for _, destination := range request.DestTermID {
			jobs = append(jobs, receiptJob{
				conn: conn, messageID: response.MsgID, deliveryID: h.messageIDs.next(submittedAt),
				destination: destination, destinationID: request.SrcTermID, submittedAt: submittedAt,
			})
		}
		if !h.receipts.enqueue(jobs) {
			response.Status = statusBusy
			return response.IEncode()
		}
	}
	h.logger.CtxInfof(
		ctx,
		"submit accepted remote=%s account=%q msg_id=%s sequence=%d recipients=%d content_bytes=%d receipt_requested=%t",
		conn.RemoteAddr(), conn.GetBizData().account, response.MsgID, request.GetSequenceID(),
		len(request.DestTermID), len(request.MsgContent), request.NeedReport == smgp.NEED_REPORT,
	)
	return response.IEncode()
}

func validateSubmit(request *smgp30.Submit) uint32 {
	if request.MsgType != smgp.MT {
		return 30 // Illegal MsgType; this example accepts MT submissions only.
	}
	if request.NeedReport > 1 {
		return statusStructure
	}
	if request.Priority > 9 {
		return 31
	}
	if request.DestTermIDCount == 0 || request.DestTermIDCount > 100 {
		return 47
	}
	if int(request.DestTermIDCount) != len(request.DestTermID) {
		return statusStructure
	}
	for _, destination := range request.DestTermID {
		if destination == "" {
			return 47
		}
	}
	if request.SrcTermID == "" {
		return 46
	}
	if len(request.MsgContent) > 140 || int(request.MsgLength) != len(request.MsgContent) {
		return 36
	}
	switch request.MsgFormat {
	case smgp.ASCII, smgp.BINARY, smgp.UCS2, smgp.GB18030:
		return statusOK
	default:
		return 34
	}
}
