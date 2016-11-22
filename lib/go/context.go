package frugal

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattrobenolt/gocql/uuid"
)

// ErrTimeout is returned when a request timed out.
var ErrTimeout = errors.New("frugal: request timed out")

const (
	// Header containing correlation id
	cidHeader = "_cid"

	// Header containing op id (uint64 as string)
	opIDHeader = "_opid"

	// Header containing request timeout (milliseconds as string)
	timeoutHeader = "_timeout"

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
type FContext struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
	mu              sync.RWMutex
}

// NewFContext returns a Context for the given correlation id. If an empty
// correlation id is given, one will be generated. A Context should belong to a
// single request for the lifetime of the request. It can be reused once its
// request has completed, though they should generally not be reused.
func NewFContext(correlationID string) *FContext {
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}
	ctx := &FContext{
		requestHeaders: map[string]string{
			cidHeader:     correlationID,
			opIDHeader:    "0",
			timeoutHeader: strconv.FormatInt(int64(defaultTimeout/time.Millisecond), 10),
		},
		responseHeaders: make(map[string]string),
	}
	return ctx
}

// CorrelationID returns the correlation id for the context
func (c *FContext) CorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestHeaders[cidHeader]
}

// setOpID returns the operation id for the context
func (c *FContext) setOpID(id uint64) {
	opIDStr := strconv.FormatUint(id, 10)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHeaders[opIDHeader] = opIDStr
}

// opID returns the operation id for the context
func (c *FContext) opID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	opIDStr := c.requestHeaders[opIDHeader]
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	if err != nil {
		// Should not happen.
		panic(err)
	}
	return id
}

// AddRequestHeader adds a request header to the context for the given name.
// The headers _cid and _opid are reserved. Returns the same FContext to allow
// for chaining calls.
func (c *FContext) AddRequestHeader(name, value string) *FContext {
	if name == cidHeader || name == opIDHeader {
		return c
	}
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
	return c
}

// RequestHeader gets the named request header
func (c *FContext) RequestHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.requestHeaders[name]
	return val, ok
}

// RequestHeaders returns the request headers map
func (c *FContext) RequestHeaders() map[string]string {
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
func (c *FContext) AddResponseHeader(name, value string) *FContext {
	if name == opIDHeader {
		return c
	}
	c.addResponseHeader(name, value)
	return c
}

// ResponseHeader gets the named response header
func (c *FContext) ResponseHeader(name string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.responseHeaders[name]
	return val, ok
}

// ResponseHeaders returns the response headers map
func (c *FContext) ResponseHeaders() map[string]string {
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
func (c *FContext) SetTimeout(timeout time.Duration) *FContext {
	c.mu.Lock()
	c.requestHeaders[timeoutHeader] = strconv.FormatInt(int64(timeout/time.Millisecond), 10)
	c.mu.Unlock()
	return c
}

// Timeout returns the request timeout.
func (c *FContext) Timeout() time.Duration {
	c.mu.RLock()
	timeoutMillisStr := c.requestHeaders[timeoutHeader]
	c.mu.RUnlock()
	timeoutMillis, err := strconv.ParseInt(timeoutMillisStr, 10, 64)
	if err != nil {
		return 0
	}
	return time.Millisecond * time.Duration(timeoutMillis)
}

func (c *FContext) setResponseOpID(id string) {
	c.mu.Lock()
	c.responseHeaders[opIDHeader] = id
	c.mu.Unlock()
}

// addRequestHeader bypasses the check for reserved headers.
func (c *FContext) addRequestHeader(name, value string) {
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
}

// addResponseHeader bypasses the check for reserved headers.
func (c *FContext) addResponseHeader(name, value string) {
	c.mu.Lock()
	c.responseHeaders[name] = value
	c.mu.Unlock()
}

// generateCorrelationID returns a random string id. It's assigned to a var for
// testability purposes.
var generateCorrelationID = func() string {
	return strings.Replace(uuid.RandomUUID().String(), "-", "", -1)
}
