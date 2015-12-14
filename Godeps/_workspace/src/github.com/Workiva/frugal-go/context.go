package frugal

import (
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/mattrobenolt/gocql/uuid"
)

const (
	cid  = "_cid"
	opID = "_opid"
)

var nextOpID uint64 = 0

// Context is the message context for a frugal message.
type Context interface {
	// CorrelationID returns the correlation id for the context
	CorrelationID() string
	// OpID returns the operation id for the context
	OpID() uint64
	// AddRequestHeader adds a request header to the context for the given name
	AddRequestHeader(name, value string)
	// RequestHeader gets the named request header
	RequestHeader(name string) (string, bool)
	// RequestHeaders returns the request headers map
	RequestHeaders() map[string]string
	// AddResponseHeader adds a response header to the context for the given
	// name
	AddResponseHeader(name, value string)
	// ResponseHeader gets the named response header
	ResponseHeader(name string) (string, bool)
	// ResponseHeaders returns the response headers map
	ResponseHeaders() map[string]string
}

type context struct {
	requestHeaders  map[string]string
	responseHeaders map[string]string
}

// Return a new context for the given correlation id. If an empty correlation
// id is given, one will be generated.
func NewContext(correlationID string) Context {
	if correlationID == "" {
		correlationID = generateCorrelationID()
	}
	ctx := &context{
		requestHeaders: map[string]string{
			cid:  correlationID,
			opID: strconv.FormatUint(atomic.LoadUint64(&nextOpID), 10),
		},
		responseHeaders: make(map[string]string),
	}

	atomic.AddUint64(&nextOpID, 1)
	return ctx
}

// CorrelationID returns the correlation id for the context
func (c *context) CorrelationID() string {
	return c.requestHeaders[cid]
}

// OpID returns the operation id for the context
func (c *context) OpID() uint64 {
	opIDStr := c.requestHeaders[opID]
	id, err := strconv.ParseUint(opIDStr, 10, 64)
	if err != nil {
		// Should not happen.
		panic(err)
	}
	return id
}

// AddRequestHeader adds a request header to the context for the given name
func (c *context) AddRequestHeader(name, value string) {
	c.requestHeaders[name] = value
}

// RequestHeader gets the named request header
func (c *context) RequestHeader(name string) (string, bool) {
	val, ok := c.requestHeaders[name]
	return val, ok
}

// RequestHeaders returns the request headers map
func (c *context) RequestHeaders() map[string]string {
	return c.requestHeaders
}

// AddResponseHeader adds a response header to the context for the given name
func (c *context) AddResponseHeader(name, value string) {
	c.responseHeaders[name] = value
}

// ResponseHeader gets the named response header
func (c *context) ResponseHeader(name string) (string, bool) {
	val, ok := c.responseHeaders[name]
	return val, ok
}

// ResponseHeaders returns the response headers map
func (c *context) ResponseHeaders() map[string]string {
	return c.responseHeaders
}

// generateCorrelationID returns a random string id
func generateCorrelationID() string {
	return strings.Replace(uuid.RandomUUID().String(), "-", "", -1)
}
