package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hujm2023/hlog"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
	"github.com/hujm2023/go-sms-protocol/nioserver"
)

type fakeSMSConn struct {
	bizMu    sync.RWMutex
	bizData  connectionSession
	seq      atomic.Uint32
	missed   atomic.Int32
	isClosed atomic.Bool
	writes   chan []byte
}

func newFakeSMSConn() *fakeSMSConn {
	return &fakeSMSConn{writes: make(chan []byte, 8)}
}

func (c *fakeSMSConn) Write(ctx context.Context, data []byte) error {
	if c.isClosed.Load() {
		return nioserver.ErrConnectionClosed
	}
	dataCopy := append([]byte(nil), data...)
	select {
	case c.writes <- dataCopy:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *fakeSMSConn) WriteAndClose(ctx context.Context, data []byte) error {
	if err := c.Write(ctx, data); err != nil {
		return err
	}
	return c.Close()
}

func (c *fakeSMSConn) AsyncWrite(ctx context.Context, data []byte) {
	_ = c.Write(ctx, data)
}

func (c *fakeSMSConn) Close() error {
	c.isClosed.Store(true)
	return nil
}

func (c *fakeSMSConn) RemoteAddr() string {
	return "127.0.0.1:12345"
}

func (c *fakeSMSConn) NextSequenceID() uint32 {
	return c.seq.Add(1)
}

func (c *fakeSMSConn) GetBizData() connectionSession {
	c.bizMu.RLock()
	defer c.bizMu.RUnlock()
	return c.bizData
}

func (c *fakeSMSConn) SetBizData(data connectionSession) {
	c.bizMu.Lock()
	c.bizData = data
	c.bizMu.Unlock()
}

func (c *fakeSMSConn) OnSendActiveTest() {
	c.missed.Add(1)
}

func (c *fakeSMSConn) OnReceiveActiveTest() {
	c.missed.Store(0)
}

func (c *fakeSMSConn) NoActiveTestCount() int {
	return int(c.missed.Load())
}

func TestBuildDeliveryReceipt(t *testing.T) {
	submittedAt := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	doneAt := submittedAt.Add(300 * time.Millisecond)
	job := receiptJob{
		messageID:     0x0102030405060708,
		deliveryID:    0x1112131415161718,
		destination:   "13800138000",
		destinationID: "10690001",
		serviceID:     "notice",
		submittedAt:   submittedAt,
		smscSequence:  88,
	}

	delivery, err := buildDeliveryReceipt(job, doneAt, 23)
	if err != nil {
		t.Fatalf("buildDeliveryReceipt() error = %v", err)
	}
	if delivery.MsgID != job.deliveryID || delivery.GetSequenceID() != 23 {
		t.Fatalf("delivery identifiers = msg %x sequence %d", delivery.MsgID, delivery.GetSequenceID())
	}
	if delivery.RegisteredDeliver != 1 || int(delivery.MsgLength) != len(delivery.MsgContent) {
		t.Fatalf("delivery report marker/length = %d/%d", delivery.RegisteredDeliver, delivery.MsgLength)
	}

	var content cmpp.SubPduDeliveryContent
	if err := content.IDecode(delivery.MsgContent); err != nil {
		t.Fatalf("SubPduDeliveryContent.IDecode() error = %v", err)
	}
	if content.MsgID != job.messageID || content.Stat != cmpp20.DELIVERED {
		t.Fatalf("delivery content = msg %x status %q", content.MsgID, content.Stat)
	}
	if content.DestTerminalID != job.destination || content.SMSCSequence != job.smscSequence {
		t.Fatalf("delivery destination/sequence = %q/%d", content.DestTerminalID, content.SMSCSequence)
	}
}

func TestReceiptDispatcherPushesDeliveryReceipt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dispatcher := newReceiptDispatcher(
		ctx,
		hlog.DefaultLogger(),
		1,
		99,
		100*time.Millisecond,
		500*time.Millisecond,
		time.Second,
	)
	dispatcher.delay = func() time.Duration { return 0 }
	dispatcher.now = func() time.Time {
		return time.Date(2026, time.August, 25, 12, 0, 1, 0, time.UTC)
	}
	conn := newFakeSMSConn()
	job := receiptJob{
		conn:          conn,
		messageID:     0x0102030405060708,
		deliveryID:    0x1112131415161718,
		destination:   "13800138000",
		destinationID: "10690001",
		serviceID:     "notice",
		submittedAt:   time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC),
		smscSequence:  88,
	}
	if !dispatcher.enqueue([]receiptJob{job}) {
		cancel()
		t.Fatal("receipt job was rejected")
	}

	select {
	case data := <-conn.writes:
		pdu, err := cmpp20.DecodeCMPP20(data)
		if err != nil {
			cancel()
			t.Fatalf("DecodeCMPP20() error = %v", err)
		}
		delivery, ok := pdu.(*cmpp20.PduDeliver)
		if !ok {
			cancel()
			t.Fatalf("pushed PDU type = %T, want *cmpp20.PduDeliver", pdu)
		}
		if delivery.MsgID != job.deliveryID {
			cancel()
			t.Fatalf("pushed delivery MsgID = %x, want %x", delivery.MsgID, job.deliveryID)
		}
	case <-time.After(time.Second):
		cancel()
		t.Fatal("timed out waiting for delivery receipt")
	}

	cancel()
	waitCtx, waitCancel := context.WithTimeout(context.Background(), time.Second)
	defer waitCancel()
	if err := dispatcher.wait(waitCtx); err != nil {
		t.Fatalf("receiptDispatcher.wait() error = %v", err)
	}
}

func TestHeartbeatManagerCheckConnection(t *testing.T) {
	conn := newFakeSMSConn()
	manager := &heartbeatManager{
		ctx:          context.Background(),
		logger:       hlog.DefaultLogger(),
		maxMissed:    3,
		writeTimeout: time.Second,
		connections:  make(map[nioserver.ISMSConn[connectionSession]]struct{}),
	}
	manager.checkConnection(conn)

	select {
	case data := <-conn.writes:
		pdu, err := cmpp20.DecodeCMPP20(data)
		if err != nil {
			t.Fatalf("DecodeCMPP20() error = %v", err)
		}
		if _, ok := pdu.(*cmpp20.PduActiveTest); !ok {
			t.Fatalf("heartbeat PDU type = %T, want *cmpp20.PduActiveTest", pdu)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for heartbeat")
	}
	if conn.NoActiveTestCount() != 1 {
		t.Fatalf("missed heartbeat count = %d, want 1", conn.NoActiveTestCount())
	}

	conn.OnReceiveActiveTest()
	if conn.NoActiveTestCount() != 0 {
		t.Fatalf("missed heartbeat count after response = %d, want 0", conn.NoActiveTestCount())
	}
}

var _ nioserver.ISMSConn[connectionSession] = (*fakeSMSConn)(nil)
var _ protocol.PDU = (*cmpp20.PduDeliver)(nil)
