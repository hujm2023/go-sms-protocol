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
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
	"github.com/hujm2023/go-sms-protocol/nioserver"
)

const maxCMPP20FrameSize uint32 = cmpp20.MaxSubmitLength

var (
	errConnectRejected   = errors.New("cmpp20server: connect rejected")
	errMissingConnection = errors.New("cmpp20server: connection missing from context")
	errNotAuthenticated  = errors.New("cmpp20server: request received before authentication")
)

type connectionSession struct {
	account         string
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
	gatewayID uint64
	sequence  atomic.Uint32
}

func (g *messageIDGenerator) next(now time.Time) uint64 {
	// CMPP MsgID 由时间、网关编号和低 16 位序号组成。跳过 0 便于日志和排障识别。
	// 这是单进程示例生成器；多实例生产部署应集中分配或持久化序号，避免实例间碰撞。
	sequence := g.sequence.Add(1)
	for uint16(sequence) == 0 {
		sequence = g.sequence.Add(1)
	}
	return cmpp.CombineMsgID(
		uint64(now.Month()),
		uint64(now.Day()),
		uint64(now.Hour()),
		uint64(now.Minute()),
		uint64(now.Second()),
		g.gatewayID,
		uint64(sequence),
	)
}

func newUnpackFunc(maxFrameSize uint32) nioserver.UnpackFunc {
	return func(_ context.Context, reader netpoll.Reader) (pdu protocol.PDU, err error) {
		// netpoll.Reader 持有可复用缓冲区，任何返回路径都必须 Release，否则会长期占用内存。
		defer func() {
			err = errors.Join(err, reader.Release())
		}()

		lengthBytes, err := reader.Peek(cmpp.PacketTotalLengthBytes)
		if err != nil {
			return nil, fmt.Errorf("read cmpp pdu length: %w", err)
		}
		// 在读取和分配完整帧之前限制长度，防止伪造长度导致过大内存占用。
		totalLength := binary.BigEndian.Uint32(lengthBytes)
		if totalLength < cmpp.MinCMPPPduLength || totalLength > maxFrameSize {
			return nil, fmt.Errorf("invalid cmpp pdu length: %d", totalLength)
		}

		frame, err := reader.ReadBinary(int(totalLength))
		if err != nil {
			return nil, fmt.Errorf("read cmpp pdu: %w", err)
		}
		pdu, err = cmpp20.DecodeCMPP20(frame)
		if err != nil {
			return nil, fmt.Errorf("decode cmpp 2.0 pdu: %w", err)
		}
		// 当前协议解码器按字段读取，重新编码后检查长度可以拒绝尾部夹带数据和结构长度不一致的帧。
		canonical, err := pdu.IEncode()
		if err != nil {
			return nil, fmt.Errorf("validate cmpp 2.0 pdu: %w", err)
		}
		if len(canonical) != len(frame) {
			return nil, fmt.Errorf("invalid cmpp 2.0 pdu size: decoded %d bytes from %d-byte frame", len(canonical), len(frame))
		}
		return pdu, nil
	}
}

func (h *serverHandler) handle(ctx context.Context, pdu protocol.PDU) ([]byte, error) {
	conn, ok := nioserver.GetCtxConn[connectionSession](ctx)
	if !ok {
		return nil, errMissingConnection
	}
	if pdu == nil {
		return nil, errors.New("cmpp20server: decoded pdu is nil")
	}

	if request, ok := pdu.(*cmpp20.PduConnect); ok {
		return h.handleConnect(ctx, conn, request)
	}
	// CONNECT 是连接上的第一条业务报文；鉴权前不允许处理其他 PDU。
	if !conn.GetBizData().isAuthenticated {
		return nil, errNotAuthenticated
	}

	switch request := pdu.(type) {
	case *cmpp20.PduSubmit:
		return h.handleSubmit(ctx, conn, request)
	case *cmpp20.PduActiveTest:
		return request.GenEmptyResponse().IEncode()
	case *cmpp20.PduActiveTestResp:
		conn.OnReceiveActiveTest()
		return nil, nil
	case *cmpp20.PduDeliverResp:
		h.logger.CtxInfof(
			ctx,
			"delivery receipt acknowledged remote=%s delivery_msg_id=%s sequence=%d result=%d",
			conn.RemoteAddr(),
			cmpp.MsgID2String(request.MsgID),
			request.GetSequenceID(),
			request.Result,
		)
		return nil, nil
	case *cmpp20.PduTerminate:
		return h.handleTerminate(ctx, conn, request)
	default:
		return nil, fmt.Errorf("cmpp20server: unsupported command %s", pdu.GetCommand())
	}
}

func (h *serverHandler) handleConnect(
	ctx context.Context,
	conn nioserver.ISMSConn[connectionSession],
	request *cmpp20.PduConnect,
) ([]byte, error) {
	status := cmpp20.ConnectRespStatusSuccess
	if conn.GetBizData().isAuthenticated {
		status = cmpp20.ConnectRespStatusInvalidStructure
	} else {
		status = authenticateConnect(request, h.account, h.password, h.now(), h.authSkew)
	}

	responseData, err := newConnectResponse(request, status, h.password).IEncode()
	if err != nil {
		return nil, fmt.Errorf("encode cmpp connect response: %w", err)
	}
	if status != cmpp20.ConnectRespStatusSuccess {
		// nioserver 对“响应数据 + error”执行写回后关闭连接，客户端仍能收到明确的鉴权失败状态。
		return responseData, errConnectRejected
	}

	conn.SetBizData(connectionSession{
		account:         request.SourceAddr,
		isAuthenticated: true,
	})
	h.heartbeats.register(conn)
	h.logger.CtxInfof(
		ctx,
		"cmpp connection authenticated remote=%s account=%q",
		conn.RemoteAddr(),
		request.SourceAddr,
	)
	return responseData, nil
}

func (h *serverHandler) handleSubmit(
	ctx context.Context,
	conn nioserver.ISMSConn[connectionSession],
	request *cmpp20.PduSubmit,
) ([]byte, error) {
	response, ok := request.GenEmptyResponse().(*cmpp20.PduSubmitResp)
	if !ok {
		return nil, errors.New("cmpp20server: unexpected submit response type")
	}
	response.Result = validateSubmit(request)
	if response.Result != 0 {
		return response.IEncode()
	}

	submittedAt := h.now()
	response.MsgID = h.messageIDs.next(submittedAt)
	if request.RegisteredDelivery == 1 {
		jobs := make([]receiptJob, 0, len(request.DestTerminalID))
		for _, destination := range request.DestTerminalID {
			jobs = append(jobs, receiptJob{
				conn:          conn,
				messageID:     response.MsgID,
				deliveryID:    h.messageIDs.next(submittedAt),
				destination:   destination,
				destinationID: request.SrcID,
				serviceID:     request.ServiceID,
				submittedAt:   submittedAt,
				smscSequence:  request.GetSequenceID(),
			})
		}
		// 先确保一个 SUBMIT 的回执任务全部进入有界队列，再确认受理。
		// 队列满时返回协议流控错误 8，不能先回复成功再静默丢回执。
		if !h.receipts.enqueue(jobs) {
			response.Result = 8
			h.logger.CtxWarnf(
				ctx,
				"submit rejected by receipt backpressure remote=%s sequence=%d recipients=%d",
				conn.RemoteAddr(),
				request.GetSequenceID(),
				len(request.DestTerminalID),
			)
			return response.IEncode()
		}
	}

	// 示例只打印处理元数据，不记录密码、短信正文和完整接收号码，避免日志成为敏感数据副本。
	h.logger.CtxInfof(
		ctx,
		"submit accepted remote=%s account=%q msg_id=%s sequence=%d recipients=%d service_id=%q content_bytes=%d receipt_requested=%t",
		conn.RemoteAddr(),
		conn.GetBizData().account,
		cmpp.MsgID2String(response.MsgID),
		request.GetSequenceID(),
		len(request.DestTerminalID),
		request.ServiceID,
		len(request.MsgContent),
		request.RegisteredDelivery == 1,
	)
	return response.IEncode()
}

func (h *serverHandler) handleTerminate(
	ctx context.Context,
	conn nioserver.ISMSConn[connectionSession],
	request *cmpp20.PduTerminate,
) ([]byte, error) {
	responseData, err := request.GenEmptyResponse().IEncode()
	if err != nil {
		return nil, fmt.Errorf("encode cmpp terminate response: %w", err)
	}
	// TERMINATE_RESP 必须先写完再关闭；普通 Close 可能让响应仍在缓冲区时就断开连接。
	if err := conn.WriteAndClose(ctx, responseData); err != nil {
		return nil, fmt.Errorf("write cmpp terminate response: %w", err)
	}
	return nil, nil
}

func validateSubmit(request *cmpp20.PduSubmit) uint8 {
	// 返回 CMPP_SUBMIT_RESP.Result，而不是 Go error：这些都是客户端可修正的协议拒绝，
	// 不应被 nioserver 当成连接级故障直接断开。
	if request.PkTotal == 0 || request.PkNumber == 0 || request.PkNumber > request.PkTotal {
		return 1
	}
	if request.DestUsrTL == 0 || int(request.DestUsrTL) != len(request.DestTerminalID) {
		return 1
	}
	if request.RegisteredDelivery > 2 {
		return 1
	}
	if int(request.MsgLength) != len(request.MsgContent) {
		return 4
	}
	return 0
}
