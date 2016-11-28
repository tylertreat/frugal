package frugal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Ensures NewFContext creates an FContext with the given correlation id.
func TestCorrelationID(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, corid, ctx.CorrelationID())
}

// Ensures NewFContext creates an FContext and generates a correlation id if
// one is not supplied.
func TestNewCorrelationID(t *testing.T) {
	cid := "abc"
	oldCID := generateCorrelationID
	defer func() { generateCorrelationID = oldCID }()
	generateCorrelationID = func() string { return cid }

	ctx := NewFContext("")

	assert.Equal(t, cid, ctx.CorrelationID())
}

// Ensures the "_opid" request header for an FContext is returned for calls to
// opID.
func TestOpID(t *testing.T) {
	corid := "fooid"
	opid := "12345"
	ctx := NewFContext(corid)
	ctx.requestHeaders[opIDHeader] = opid
	assert.Equal(t, uint64(12345), ctx.opID())
}

// Ensures the "_timeout" request header is correctly set and calls to Timeout
// return the correct Duration.
func TestTimeout(t *testing.T) {
	// Check default timeout (5 seconds).
	ctx := NewFContext("")
	timeoutStr, _ := ctx.RequestHeader(timeoutHeader)
	assert.Equal(t, "5000", timeoutStr)
	assert.Equal(t, defaultTimeout, ctx.Timeout())

	// Set timeout and check expected values.
	ctx.SetTimeout(10 * time.Second)
	timeoutStr, _ = ctx.RequestHeader(timeoutHeader)
	assert.Equal(t, "10000", timeoutStr)
	assert.Equal(t, 10*time.Second, ctx.Timeout())
}

// Ensures AddRequestHeader properly adds the key-value pair to the context
// RequestHeaders.
func TestRequestHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddRequestHeader("foo", "bar"))
	assert.Equal(t, ctx, ctx.AddRequestHeader("_cid", "123"))
	val, ok := ctx.RequestHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.RequestHeaders()["foo"])
	assert.Equal(t, corid, ctx.RequestHeaders()[cidHeader])
	assert.NotEqual(t, "", ctx.RequestHeaders()[opIDHeader])
}

// Ensures AddResponseHeader properly adds the key-value pair to the context
// ResponseHeaders.
func TestResponseHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddResponseHeader("foo", "bar"))
	assert.Equal(t, ctx, ctx.AddResponseHeader("_opid", "1"))
	val, ok := ctx.ResponseHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.ResponseHeaders()["foo"])
	assert.Equal(t, "", ctx.ResponseHeaders()[cidHeader])
	assert.Equal(t, "", ctx.ResponseHeaders()[opIDHeader])
}
