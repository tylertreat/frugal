package frugal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattrobenolt/gocql/uuid"
)

// ErrTimeout is returned when a request timed out.
var ErrTimeout = errors.New("frugal: request timed out")

const (
	cid            = "_cid"
	opID           = "_opid"
	defaultTimeout = 5 * time.Second
)

// FContext is the context for a Frugal message. Every RPC has an FContext,
// which can be used to set request headers, response headers, and the request
// timeout. The default timeout is five seconds. An FContext is also sent with
// every publish message which is then received by subscribers.
//
// In addition to headers, the FContext also contains a correlation ID which
// can be used for distributed tracing purposes. A random correlation ID is
// generated for each FContext if one is not provided.
//
// FContext also plays a key role in Frugal's multiplexing support. A unique,
// per-request operation ID is set on every FContext before a request is made.
// This operation ID is sent in the request and included in the response, which
// is then used to correlate a response to a request. The operation ID is an
// internal implementation detail and is not exposed to the user.
//
// An FContext should belong to a single request for the lifetime of that
// request. It can be reused once the request has completed, though they should
// generally not be reused.
//
// Implementations of FContext must adhere to the following:
//		1)	The CorrelationID should be stored as a request header with the
//			header name "_cid"
//		2)	Threadsafe
type FContext interface {
	// CorrelationID returns the correlation id for the context.
	CorrelationID() string

	// AddRequestHeader adds a request header to the context for the given
	// name. The headers _cid and _opid are reserved. Returns the same FContext
	// to allow for chaining calls.
	AddRequestHeader(name, value string) FContext

	// RequestHeader gets the named request header.
	RequestHeader(name string) (string, bool)

	// RequestHeaders returns the request headers map.
	RequestHeaders() map[string]string

	// AddResponseHeader adds a response header to the context for the given
	// name. The _opid header is reserved. Returns the same FContext to allow
	// for chaining calls.
	AddResponseHeader(name, value string) FContext

	// ResponseHeader gets the named response header.
	ResponseHeader(name string) (string, bool)

	// ResponseHeaders returns the response headers map.
	ResponseHeaders() map[string]string

	// SetTimeout sets the request timeout. Default is 5 seconds. Returns the
	// same FContext to allow for chaining calls.
	SetTimeout(timeout time.Duration) FContext

	// Timeout returns the request timeout.
	Timeout() time.Duration
}

// FContextImpl is an implementation of FContext.
type FContextImpl struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
	mu              sync.RWMutex
	timeout         time.Duration
}

// NewFContext returns a Context for the given correlation id. If an empty
// correlation id is given, one will be generated. A Context should belong to a
// single request for the lifetime of the request. It can be reused once its
// request has completed, though they should generally not be reused.
func NewFContext(correlationID string) FContext {
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}
	ctx := &FContextImpl{
		requestHeaders: map[string]string{
			cid:  correlationID,
			opID: "0",
		},
		responseHeaders: make(map[string]string),
		timeout:         defaultTimeout,
	}
	return ctx
}

// CorrelationID returns the correlation id for the context.
func (c *FContextImpl) CorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestHeaders[cid]
}

// AddRequestHeader adds a request header to the context for the given name.
// The headers _cid and _opid are reserved. Returns the same FContext to allow
// for chaining calls.
func (c *FContextImpl) AddRequestHeader(name, value string) FContext {
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
	return c
}

// RequestHeader gets the named request header.
func (c *FContextImpl) RequestHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.requestHeaders[name]
	return val, ok
}

// RequestHeaders returns the request headers map.
func (c *FContextImpl) RequestHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.requestHeaders))
	for name, value := range c.requestHeaders {
		headers[name] = value
	}
	return headers
}

// AddResponseHeader adds a response header to the context for the given name.
// The _opid header is reserved. Returns the same FContext to allow for
// chaining calls.
func (c *FContextImpl) AddResponseHeader(name, value string) FContext {
	c.mu.Lock()
	c.responseHeaders[name] = value
	c.mu.Unlock()
	return c
}

// ResponseHeader gets the named response header.
func (c *FContextImpl) ResponseHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.responseHeaders[name]
	return val, ok
}

// ResponseHeaders returns the response headers map.
func (c *FContextImpl) ResponseHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	headers := make(map[string]string, len(c.responseHeaders))
	for name, value := range c.responseHeaders {
		headers[name] = value
	}
	return headers
}

// SetTimeout sets the request timeout. Default is 5 seconds. Returns the same
// FContext to allow for chaining calls.
func (c *FContextImpl) SetTimeout(timeout time.Duration) FContext {
	c.mu.Lock()
	c.timeout = timeout
	c.mu.Unlock()
	return c
}

// Timeout returns the request timeout.
func (c *FContextImpl) Timeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeout
}

// setRequestOpID sets the request operation id for context
func setRequestOpID(ctx FContext, id uint64) {
	opIDStr := strconv.FormatUint(id, 10)
	ctx.AddRequestHeader(opID, opIDStr)
}

// opID returns the request operation id for the given context
func getOpID(ctx FContext) uint64 {
	opIDStr, ok := ctx.RequestHeader(opID)
	// Should not happen unless this is a malformed implemtation of FContext.
	if !ok {
		panic(fmt.Errorf("The FContext should always contain the %s request header", opID))
	}
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	// Should not happen unless this is a malformed implemtation of FContext.
	if err != nil {
		panic(err)
	}
	return id
}

// setResponseOpID sets the response operation id for context
func setResponseOpID(ctx FContext, id string) {
	ctx.AddResponseHeader(opID, id)
}

// generateCorrelationID returns a random string id. It's assigned to a var for
// testability purposes.
var generateCorrelationID = func() string {
	return strings.Replace(uuid.RandomUUID().String(), "-", "", -1)
}
