package frugal

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"git.apache.org/thrift.git/lib/go/thrift"
)

var nextOpID uint64

// FAsyncCallback is an internal callback which is constructed by generated
// code and invoked by an FRegistry when a RPC response is received. In other
// words, it's used to complete RPCs. The operation ID on FContext is used to
// look up the appropriate callback. FAsyncCallback is passed an in-memory
// TTransport which wraps the complete message. The callback returns an error
// or throws an exception if an unrecoverable error occurs and the transport
// needs to be shutdown.
type FAsyncCallback func(thrift.TTransport) error

// FRegistry is responsible for multiplexing and handling received messages.
// Typically there is a client implementation and a server implementation. An
// FRegistry is used by an FTransport.
//
// The client implementation is used on the client side, which is making RPCs.
// When a request is made, an FAsyncCallback is registered to an FContext. When
// a response for the FContext is received, the FAsyncCallback is looked up,
// executed, and unregistered.
//
// The server implementation is used on the server side, which is handling
// RPCs. It does not actually register FAsyncCallbacks but rather has an
// FProcessor registered with it. When a message is received, it's buffered and
// passed to the FProcessor to be handled.
type FRegistry interface {
	// Register a callback for the given Context.
	Register(ctx *FContext, callback FAsyncCallback) error
	// Unregister a callback for the given Context.
	Unregister(*FContext)
	// Execute dispatches a single Thrift message frame.
	Execute([]byte) error
}

type fRegistry struct {
	mu       sync.RWMutex
	handlers map[uint64]FAsyncCallback
}

// NewFRegistry creates a Registry intended for use by Frugal clients.
// This is only to be called by generated code.
func NewFRegistry() FRegistry {
	return &fRegistry{handlers: make(map[uint64]FAsyncCallback)}
}

// Register a callback for the given Context.
func (c *fRegistry) Register(ctx *FContext, callback FAsyncCallback) error {
	// An FContext can be reused for multiple requests. Because of this, every
	// time an FContext is registered, it must be assigned a new op id to
	// ensure we can properly correlate responses. We use a monotonically
	// increasing atomic uint64 for this purpose. If the FContext already has
	// an op id, it has been used for a request. We check the handlers map to
	// ensure that request is not still in-flight.
	opID := ctx.opID()
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.handlers[opID]
	if ok {
		return errors.New("frugal: context already registered")
	}
	opID = atomic.AddUint64(&nextOpID, 1)
	ctx.setOpID(opID)
	c.handlers[opID] = callback
	return nil
}

// Unregister a callback for the given Context.
func (c *fRegistry) Unregister(ctx *FContext) {
	opID := ctx.opID()
	c.mu.Lock()
	delete(c.handlers, opID)
	c.mu.Unlock()
}

// Execute dispatches a single Thrift message frame.
func (c *fRegistry) Execute(frame []byte) error {
	headers, err := getHeadersFromFrame(frame)
	if err != nil {
		logger().Warn("frugal: invalid protocol frame headers:", err)
		return err
	}

	opid, err := strconv.ParseUint(headers[opID], 10, 64)
	if err != nil {
		logger().Warn("frugal: invalid protocol frame, op id not a uint64:", err)
		return err
	}

	c.mu.RLock()
	handler, ok := c.handlers[opid]
	if !ok {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	return handler(&thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(frame)})
}
