package nioserver

import (
	"context"
	"sync/atomic"

	"github.com/cloudwego/netpoll"
	"github.com/cloudwego/netpoll/mux"
)

type IActiveTest interface {
	OnReceiveActiveTest()

	NoActiveTestCount() int
}

type ISMSConn[T any] interface {
	IActiveTest

	// 异步写入对端
	AsyncWrite(ctx context.Context, data []byte)

	// RemoteAddr 对端地址
	RemoteAddr() string

	// NextSequenceID 单条连接，序列号递增
	NextSequenceID() uint32

	GetBizData() T
	SetBizData(data T)
}

type connkey struct{}

var ctxkey connkey

type muxConn[T any] struct {
	conn          netpoll.Connection
	wqueue        *mux.ShardQueue // use for write
	sequenceIDGen *atomic.Uint32  // sequenceID 生成
	noActiveTest  *atomic.Int32   // 周期内未回复探活的次数
	remoteAddr    string          // 远端地址
	bizData       atomic.Value    // 业务数据, 存储 T 类型的值
}

// 修改 newSvrMuxConn 以支持泛型 T
func newSvrMuxConn[T any](conn netpoll.Connection) *muxConn[T] {
	mc := &muxConn[T]{}
	mc.conn = conn
	mc.remoteAddr = conn.RemoteAddr().String()
	mc.wqueue = mux.NewShardQueue(mux.ShardSize, conn)
	mc.sequenceIDGen = &atomic.Uint32{}
	mc.noActiveTest = &atomic.Int32{}
	// 初始化 bizData 为 T 类型的零值，防止 Load 时 panic
	var zero T
	mc.bizData.Store(zero)
	return mc
}

// 修改 AsyncWrite 的接收者
func (m *muxConn[T]) AsyncWrite(ctx context.Context, data []byte) {
	m.wqueue.Add(func() (buf netpoll.Writer, isNil bool) {
		w := netpoll.NewLinkBuffer(0)
		_, _ = w.WriteBinary(data)
		return w, false
	})
}

// 修改 RemoteAddr 的接收者
func (m *muxConn[T]) RemoteAddr() string {
	return m.remoteAddr
}

// 修改 NextSequenceID 的接收者
func (m *muxConn[T]) NextSequenceID() uint32 {
	n := m.sequenceIDGen.Add(1)
	if n == 0 {
		n = m.sequenceIDGen.Add(1)
	}
	return n
}

// 修改 NoActiveTestCount 的接收者
func (m *muxConn[T]) NoActiveTestCount() int {
	return int(m.noActiveTest.Add(1))
}

// 修改 GetBizData 以支持泛型 T
func (m *muxConn[T]) GetBizData() T {
	v := m.bizData.Load()
	if data, ok := v.(T); ok {
		return data
	}
	// 如果类型断言失败，返回 T 类型的零值
	var zero T
	return zero
}

// 修改 SetBizData 以支持泛型 T
func (m *muxConn[T]) SetBizData(data T) {
	m.bizData.Store(data)
}

// 修改 OnReceiveActiveTest 的接收者
func (m *muxConn[T]) OnReceiveActiveTest() {
	// 重置为0
	m.noActiveTest.Store(0)
}

// 修改 fillCtx 以支持泛型 T
func fillCtx[T any](ctx context.Context, conn ISMSConn[T]) context.Context {
	return context.WithValue(ctx, ctxkey, conn)
}

// 修改 GetCtxConn 以支持泛型 T
func GetCtxConn[T any](ctx context.Context) (ISMSConn[T], bool) {
	conn, ok := ctx.Value(ctxkey).(ISMSConn[T])
	return conn, ok
}
