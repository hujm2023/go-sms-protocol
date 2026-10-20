package main

import (
	"context"
	"io"
	"testing"

	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

type rejectionConn struct {
	*fakeCallbackConn
	result              <-chan error
	reportedBeforeFlush bool
}

func (c *rejectionConn) WriteAndClose(ctx context.Context, data []byte) error {
	select {
	case <-c.result:
		c.reportedBeforeFlush = true
	default:
	}
	return c.fakeCallbackConn.WriteAndClose(ctx, data)
}

func TestCallbackRejectionFlushesBeforeSignallingShutdown(t *testing.T) {
	cfg, err := parseConfig(nil, testEnv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	h := &callbackHandler{cfg: cfg, result: make(chan error, 1)}
	conn := &rejectionConn{fakeCallbackConn: &fakeCallbackConn{}, result: h.result}
	bind := sgip12.NewBind("gw-demo", "wrong", 2000000001, 1)
	bind.Type = sgip.SMG_SP
	if _, err := h.handleConnection(context.Background(), conn, bind); err == nil {
		t.Fatal("bad callback credentials were accepted")
	}
	if !conn.closed || len(conn.final) == 0 {
		t.Fatal("callback rejection did not flush its response before signalling shutdown")
	}
	if conn.reportedBeforeFlush {
		t.Fatal("shutdown was signalled before the rejection response was flushed")
	}
	var response sgip12.BindResp
	if err := response.IDecode(conn.final); err != nil {
		t.Fatal(err)
	}
	if response.Result != sgip.STAT_ILLLOGIN || response.Sequence != bind.Sequence {
		t.Fatal("rejection response lost its result or request sequence")
	}
	select {
	case err := <-h.result:
		if err == nil {
			t.Fatal("rejection was reported as success")
		}
	default:
		t.Fatal("flushed rejection did not signal completion")
	}
}
