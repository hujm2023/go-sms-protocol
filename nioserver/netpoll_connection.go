package nioserver

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/cloudwego/netpoll"
	"github.com/cloudwego/netpoll/mux"
)

// IActiveTest defines the interface for connection active testing.
type IActiveTest interface {
	// OnReceiveActiveTest resets the counter upon receiving an active test response.
	OnReceiveActiveTest()
	// NoActiveTestCount returns the count of consecutive missed active test responses.
	NoActiveTestCount() int
}

// ISMSConn defines the interface for an SMS protocol connection with generic business data.
type ISMSConn[T any] interface {
	IActiveTest

	io.Closer

	// AsyncWrite writes data to the peer asynchronously.
	AsyncWrite(ctx context.Context, data []byte)

	// RemoteAddr returns the remote network address.
	RemoteAddr() string

	// NextSequenceID returns the next available message sequence ID for this connection.
	NextSequenceID() uint32

	// GetBizData gets the business data associated with this connection.
	GetBizData() T
	// SetBizData sets the business data associated with this connection.
	SetBizData(data T)
}

// connkey is a private type for context key to avoid collisions.
type connkey struct{}

// ctxkey is the context key for storing ISMSConn.
var ctxkey connkey

// muxConn implements ISMSConn based on netpoll and mux.
// It manages connection state, write queue, sequence ID, and business data.
type muxConn[T any] struct {
	conn          netpoll.Connection
	wqueue        *mux.ShardQueue // Sharded queue for write operations
	sequenceIDGen *atomic.Uint32  // Sequence ID generator
	noActiveTest  *atomic.Int32   // Counter for missed active test responses
	remoteAddr    string          // Cached remote address string
	bizData       atomic.Value    // Stores business data of type T atomically
}

// newSvrMuxConn creates a new server-side muxConn instance.
func newSvrMuxConn[T any](conn netpoll.Connection) *muxConn[T] {
	mc := &muxConn[T]{}
	mc.conn = conn
	mc.remoteAddr = conn.RemoteAddr().String()
	mc.wqueue = mux.NewShardQueue(mux.ShardSize, conn)
	mc.sequenceIDGen = &atomic.Uint32{}
	mc.noActiveTest = &atomic.Int32{}
	// Initialize bizData with the zero value of T to prevent panic on Load.
	var zero T
	mc.bizData.Store(zero)
	return mc
}

// AsyncWrite adds data to the write queue for asynchronous sending via netpoll mux.
func (m *muxConn[T]) AsyncWrite(ctx context.Context, data []byte) {
	m.wqueue.Add(func() (buf netpoll.Writer, isNil bool) {
		w := netpoll.NewLinkBuffer(0)
		_, _ = w.WriteBinary(data)
		return w, false
	})
}

func (m *muxConn[T]) RemoteAddr() string {
	return m.remoteAddr
}

// NextSequenceID atomically increments and returns the next message sequence ID.
// It wraps around to 1 after reaching the max uint32 value (skipping 0).
func (m *muxConn[T]) NextSequenceID() uint32 {
	n := m.sequenceIDGen.Add(1)
	if n == 0 {
		n = m.sequenceIDGen.Add(1)
	}
	return n
}

// GetBizData atomically loads and returns the business data associated with the connection.
// Returns the zero value of T if the stored value is not of type T (e.g., not set yet).
func (m *muxConn[T]) GetBizData() T {
	v := m.bizData.Load()
	if data, ok := v.(T); ok {
		return data
	}
	// If type assertion fails, return the zero value of T.
	var zero T
	return zero
}

// SetBizData atomically stores the business data associated with the connection.
func (m *muxConn[T]) SetBizData(data T) {
	m.bizData.Store(data)
}

// NoActiveTestCount atomically increments and returns the count of consecutive missed active test responses.
func (m *muxConn[T]) NoActiveTestCount() int {
	return int(m.noActiveTest.Add(1))
}

// OnReceiveActiveTest atomically resets the missed active test response counter to 0.
func (m *muxConn[T]) OnReceiveActiveTest() {
	// Reset to 0
	m.noActiveTest.Store(0)
}

func (m *muxConn[T]) Close() error {
	// _ = m.wqueue.Close() // wqueue will be closed by BaseServer.OnCloseConn
	return m.conn.Close()
}

// fillCtx adds the ISMSConn instance to the context.
func fillCtx[T any](ctx context.Context, conn ISMSConn[T]) context.Context {
	return context.WithValue(ctx, ctxkey, conn)
}

// GetCtxConn extracts the ISMSConn instance from the context.
func GetCtxConn[T any](ctx context.Context) (ISMSConn[T], bool) {
	conn, ok := ctx.Value(ctxkey).(ISMSConn[T])
	return conn, ok
}
