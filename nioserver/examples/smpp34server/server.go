package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type connectionSession struct {
	bound bool
	mode  smpp.CMDId
}

type receiptJob struct {
	conn                           nioserver.ISMSConn[connectionSession]
	messageID, source, destination string
	sourceTON, sourceNPI           uint8
	destinationTON, destinationNPI uint8
	submitted                      time.Time
}

type serverHandler struct {
	cfg        config
	logger     smslog.FullLogger
	heartbeats *heartbeatManager
	receipts   *exampleutil.Queue[receiptJob]
	messageIDs atomic.Uint64
}

func nextSequence(conn nioserver.ISMSConn[connectionSession]) uint32 {
	for {
		if sequence := conn.NextSequenceID() & smpp.SEQUENCE_NUM_END; sequence != 0 {
			return sequence
		}
	}
}

func (h *serverHandler) handle(ctx context.Context, pdu protocol.PDU) ([]byte, error) {
	conn, ok := nioserver.GetCtxConn[connectionSession](ctx)
	if !ok {
		return nil, errors.New("missing connection")
	}
	if bind, ok := pdu.(*smpp34.Bind); ok {
		return h.bind(ctx, conn, bind)
	}
	if !conn.GetBizData().bound {
		return nil, errors.New("PDU received before bind")
	}
	switch p := pdu.(type) {
	case *smpp34.SubmitSm:
		return h.submit(ctx, conn, p)
	case *smpp34.DeliverSmResp:
		h.logger.CtxInfof(ctx, "receipt acknowledged sequence=%d status=%d", p.Sequence, p.Status)
		return nil, nil
	case *smpp34.EnquireLink:
		return p.GenEmptyResponse().IEncode()
	case *smpp34.EnquireLinkResp:
		if p.Status == smpp.ESME_ROK {
			conn.OnReceiveActiveTest()
		}
		return nil, nil
	case *smpp34.Unbind:
		data, err := p.GenEmptyResponse().IEncode()
		if err != nil {
			return nil, err
		}
		return nil, conn.WriteAndClose(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported command: %s", pdu.GetCommand())
	}
}

func (h *serverHandler) bind(ctx context.Context, conn nioserver.ISMSConn[connectionSession], p *smpp34.Bind) ([]byte, error) {
	response := p.GenEmptyResponse().(*smpp34.BindResp)
	switch {
	case conn.GetBizData().bound:
		response.Status = smpp.ESME_RALYBND
	case subtle.ConstantTimeCompare([]byte(p.SystemID), []byte(h.cfg.Account)) != 1:
		response.Status = smpp.ESME_RINVSYSID
	case subtle.ConstantTimeCompare([]byte(p.Password), []byte(h.cfg.Password)) != 1:
		response.Status = smpp.ESME_RINVPASWD
	default:
		response.SystemID = "example-smsc"
	}
	data, err := response.IEncode()
	if err != nil {
		return nil, err
	}
	if response.Status != smpp.ESME_ROK {
		return data, errors.New("bind rejected")
	}
	conn.SetBizData(connectionSession{bound: true, mode: p.ID})
	h.heartbeats.register(conn)
	h.logger.CtxInfof(ctx, "smpp bind accepted remote=%s mode=%s", conn.RemoteAddr(), p.ID)
	return data, nil
}

func (h *serverHandler) submit(ctx context.Context, conn nioserver.ISMSConn[connectionSession], p *smpp34.SubmitSm) ([]byte, error) {
	response := p.GenEmptyResponse().(*smpp34.SubmitSmResp)
	mode := conn.GetBizData().mode
	switch {
	case mode == smpp.BIND_RECEIVER || (p.RegisteredDelivery == 1 && mode != smpp.BIND_TRANSCEIVER):
		response.Status = smpp.ESME_RINVBNDSTS
	case p.RegisteredDelivery > 2:
		response.Status = smpp.ESME_RINVREGDLVFLG
	case p.SourceAddr == "":
		response.Status = smpp.ESME_RINVSRCADR
	case p.DestinationAddr == "":
		response.Status = smpp.ESME_RINVDSTADR
	}
	if response.Status != smpp.ESME_ROK {
		return response.IEncode()
	}
	response.MessageID = strconv.FormatUint(h.messageIDs.Add(1), 10)
	if p.RegisteredDelivery == 1 {
		job := receiptJob{
			conn: conn, messageID: response.MessageID, source: p.SourceAddr, destination: p.DestinationAddr,
			sourceTON: p.SourceAddrTon, sourceNPI: p.SourceAddrNpi,
			destinationTON: p.DestAddrTon, destinationNPI: p.DestAddrNpi, submitted: time.Now(),
		}
		if !h.receipts.TryEnqueue(job) {
			response.Status = smpp.ESME_RTHROTTLED
			response.MessageID = ""
			return response.IEncode()
		}
	}
	h.logger.CtxInfof(ctx, "submit accepted msg_id=%s sequence=%d content_bytes=%d receipt_mode=%d",
		response.MessageID, p.Sequence, len(p.ShortMessage), p.RegisteredDelivery)
	return response.IEncode()
}

func buildReceipt(job receiptJob, now time.Time, sequence uint32) (*smpp34.DeliverSm, error) {
	content, err := (&smpp34.DeliveryReceipt{
		ID: job.messageID, Sub: "001", Dlvrd: "001", SubDate: job.submitted.Format("0601021504"),
		DoneDate: now.Format("0601021504"), Stat: smpp.DELIVERED, Err: "000",
	}).IEncode()
	if err != nil {
		return nil, err
	}
	return &smpp34.DeliverSm{
		Header:     smpp.Header{ID: smpp.DELIVER_SM, Sequence: sequence},
		SourceAddr: job.destination, SourceAddrTon: job.destinationTON, SourceAddrNpi: job.destinationNPI,
		DestinationAddr: job.source, DestAddrTon: job.sourceTON, DestAddrNpi: job.sourceNPI,
		ESMClass: 0x04, DataCoding: smpp.ENCODING_IA5, SmLength: uint8(len(content)), ShortMessage: content,
		OptionalTLVs: smpp.TLVList{
			smpp.NewTLV(smpp.RECEIPTED_MESSAGE_ID, append([]byte(job.messageID), 0)),
			smpp.NewTLV(smpp.DR_MESSAGE_STATE, []byte{2}), // DELIVERED
		},
	}, nil
}

func (h *serverHandler) deliver(ctx context.Context, job receiptJob) {
	if !exampleutil.Delay(ctx, h.cfg.MinDelay, h.cfg.MaxDelay) {
		return
	}
	pdu, err := buildReceipt(job, time.Now(), nextSequence(job.conn))
	if err == nil {
		var data []byte
		data, err = pdu.IEncode()
		if err == nil {
			writeCtx, cancel := context.WithTimeout(ctx, h.cfg.WriteTimeout)
			err = job.conn.Write(writeCtx, data)
			cancel()
		}
	}
	if err != nil && ctx.Err() == nil && !errors.Is(err, nioserver.ErrConnectionClosed) {
		h.logger.Errorf("send receipt msg_id=%s: %v", job.messageID, err)
	}
}
