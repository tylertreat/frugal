package frugal

import (
	"fmt"
	"testing"

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
// getOpID.
func TestOpID(t *testing.T) {
	assert := assert.New(t)
	corid := "fooid"
	opid := "12345"
	ctx := NewFContext(corid)
	ctx.AddRequestHeader(opID, opid)
	actOpID, err := getOpID(ctx)
	assert.Nil(err)
	assert.Equal(uint64(12345), actOpID)

	delete(ctx.(*FContextImpl).requestHeaders, opID)
	_, err = getOpID(ctx)
	assert.Equal(fmt.Errorf("FContext does not have the required %s request header", opID), err)

	opIDStr := "-123"
	ctx.(*FContextImpl).requestHeaders[opID] = opIDStr
	_, err = getOpID(ctx)
	assert.Equal(fmt.Errorf("FContext has an opid that is not a non-negative integer: %s", opIDStr), err)
}

// Ensures AddRequestHeader properly adds the key-value pair to the context
// RequestHeaders.
func TestRequestHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddRequestHeader("foo", "bar"))
	val, ok := ctx.RequestHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.RequestHeaders()["foo"])
	assert.Equal(t, corid, ctx.RequestHeaders()[cid])
	assert.NotEqual(t, "", ctx.RequestHeaders()[opID])

	assert.Equal(t, ctx, ctx.AddRequestHeader(cid, "baz"))
	assert.Equal(t, ctx, ctx.AddRequestHeader(opID, "123"))

	assert.Equal(t, "baz", ctx.CorrelationID())
	actOpID, err := getOpID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), actOpID)
}

// Ensures AddResponseHeader properly adds the key-value pair to the context
// ResponseHeaders.
func TestResponseHeader(t *testing.T) {
	corid := "fooid"
	ctx := NewFContext(corid)
	assert.Equal(t, ctx, ctx.AddResponseHeader("foo", "bar"))
	val, ok := ctx.ResponseHeader("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", val)
	assert.Equal(t, "bar", ctx.ResponseHeaders()["foo"])
	assert.Equal(t, "", ctx.ResponseHeaders()[cid])
	assert.Equal(t, "", ctx.ResponseHeaders()[opID])

	assert.Equal(t, ctx, ctx.AddResponseHeader(opID, "1"))
	assert.Equal(t, "1", ctx.ResponseHeaders()[opID])
}
