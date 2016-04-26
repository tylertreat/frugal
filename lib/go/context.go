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
	cid            = "_cid"
	opID           = "_opid"
	defaultTimeout = time.Minute
)

// FContext is the message context for a frugal message. A FContext should
// belong to a single request for the lifetime of the request. It can be reused
// once its request has completed, though they should generally not be reused.
// This should only be constructed using NewFContext.
type FContext struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
	mu              sync.RWMutex
	timeout         time.Duration
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
			cid:  correlationID,
			opID: "0",
		},
		responseHeaders: make(map[string]string),
		timeout:         defaultTimeout,
	}
	return ctx
}

// CorrelationID returns the correlation id for the context
func (c *FContext) CorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestHeaders[cid]
}

// setOpID returns the operation id for the context
func (c *FContext) setOpID(id uint64) {
	opIDStr := strconv.FormatUint(id, 10)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHeaders[opID] = opIDStr
}

// opID returns the operation id for the context
func (c *FContext) opID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	opIDStr := c.requestHeaders[opID]
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	if err != nil {
		// Should not happen.
		panic(err)
	}
	return id
}

// AddRequestHeader adds a request header to the context for the given name.
// The headers _cid and _opid are reserved.
func (c *FContext) AddRequestHeader(name, value string) {
	if name == cid || name == opID {
		return
	}
	c.mu.Lock()
	c.requestHeaders[name] = value
	c.mu.Unlock()
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
// The _opid header is reserved.
func (c *FContext) AddResponseHeader(name, value string) {
	if name == opID {
		return
	}
	c.addResponseHeader(name, value)
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

// SetTimeout sets the request timeout. Default is 1 minute.
func (c *FContext) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	c.timeout = timeout
	c.mu.Unlock()
}

// Timeout returns the request timeout.
func (c *FContext) Timeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeout
}

func (c *FContext) setResponseOpID(id string) {
	c.mu.Lock()
	c.responseHeaders[opID] = id
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
