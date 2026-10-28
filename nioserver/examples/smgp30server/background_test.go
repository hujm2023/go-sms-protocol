package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
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

func TestHeartbeatManagerCheckConnection(t *testing.T) {
	conn := newFakeSMSConn()
	manager := &heartbeatManager{
		ctx:          context.Background(),
		logger:       smslog.DefaultLogger(),
		maxMissed:    3,
		writeTimeout: time.Second,
		connections:  make(map[nioserver.ISMSConn[connectionSession]]struct{}),
	}
	manager.checkConnection(conn)

	select {
	case data := <-conn.writes:
		pdu, err := smgp30.DecodeSMGP30(data)
		if err != nil {
			t.Fatalf("DecodeSMGP30() error = %v", err)
		}
		if _, ok := pdu.(*smgp30.ActiveTest); !ok {
			t.Fatalf("heartbeat PDU type = %T, want *smgp30.ActiveTest", pdu)
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
	for i := 0; i < 3; i++ {
		conn.OnSendActiveTest()
	}
	manager.checkConnection(conn)
	if !conn.isClosed.Load() {
		t.Fatal("three unanswered heartbeats did not close the connection")
	}
}

func TestReceiptDispatcherPushesBinaryReportAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := newReceiptDispatcher(ctx, smslog.DefaultLogger(), 1, 100, time.Millisecond, time.Millisecond, time.Second)
	t.Cleanup(func() {
		cancel()
		waitCtx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		if err := d.wait(waitCtx); err != nil {
			t.Errorf("receipt worker did not stop: %v", err)
		}
	})
	conn := newFakeSMSConn()
	job := receiptJob{
		conn: conn, messageID: "01006110201234000001", deliveryID: "01006110201234000002",
		destination: "13800138000", destinationID: "118900001", submittedAt: time.Now(),
	}
	if !d.enqueue([]receiptJob{job}) {
		t.Fatal("receipt job rejected")
	}
	select {
	case data := <-conn.writes:
		pdu, err := smgp30.DecodeSMGP30(data)
		if err != nil {
			t.Fatal(err)
		}
		delivery, ok := pdu.(*smgp30.Deliver)
		if !ok {
			t.Fatalf("receipt type = %T", pdu)
		}
		var receipt smgp30.DeliveryReceipt
		if err := receipt.IDecode(delivery.MsgContent); err != nil {
			t.Fatal(err)
		}
		if delivery.MsgID != job.deliveryID || delivery.IsReport != 1 || delivery.MsgFormat != 0 {
			t.Fatalf("wrong delivery identifiers or report markers: %+v", delivery)
		}
		if receipt.ID != job.messageID || receipt.Stat != "DELIVRD" || receipt.Err != "000" {
			t.Fatalf("report did not identify the original submission: %+v", receipt)
		}
		if delivery.SrcTermID != job.destination || delivery.DestTermID != job.destinationID {
			t.Fatal("receipt source/destination did not reverse the submit direction")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for receipt")
	}
	cancel()
	if d.enqueue([]receiptJob{job}) {
		t.Fatal("accepted receipt work after cancellation")
	}
}

func TestReceiptQueueRejectsWholeBatch(t *testing.T) {
	d := &receiptDispatcher{ctx: context.Background(), jobs: make(chan receiptJob, 2)}
	if !d.enqueue([]receiptJob{{messageID: "first"}}) {
		t.Fatal("initial enqueue failed")
	}
	if d.enqueue([]receiptJob{{messageID: "second"}, {messageID: "third"}}) {
		t.Fatal("accepted a batch larger than remaining capacity")
	}
	if len(d.jobs) != 1 || (<-d.jobs).messageID != "first" {
		t.Fatal("a rejected batch partially changed the queue")
	}
}

var _ nioserver.ISMSConn[connectionSession] = (*fakeSMSConn)(nil)
var _ protocol.PDU = (*smgp30.Deliver)(nil)
