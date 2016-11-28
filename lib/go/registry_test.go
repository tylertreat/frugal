package frugal

import (
	"bytes"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

// Ensures Execute returns an error if a bad frugal frame is passed.
func TestClientRegistryBadFrame(t *testing.T) {
	assert := assert.New(t)
	registry := NewFRegistry()
	assert.Error(registry.Execute([]byte{0}))
}

// Ensures Execute returns an error if the frame fruagl headers are missing an
// opID.
func TestClientRegistryMissingOpID(t *testing.T) {
	assert := assert.New(t)
	registry := NewFRegistry()
	assert.Error(registry.Execute(basicFrame))
}

// Ensures context Register, Execute, and Unregister work as intended with
// a valid frugal frame.
func TestClientRegistry(t *testing.T) {
	assert := assert.New(t)
	called := 0
	cb := func(tr thrift.TTransport) error {
		called++
		return nil
	}
	registry := NewFRegistry()
	ctx := NewFContext("")

	// Register the context for the first time
	assert.Nil(registry.Register(ctx, cb))
	opid, err := getOpID(ctx)
	assert.Nil(err)
	assert.True(opid > 0)
	// Encode a frame with this context
	transport := &thrift.TMemoryBuffer{Buffer: new(bytes.Buffer)}
	proto := &FProtocol{tProtocolFactory.GetProtocol(transport)}
	assert.Nil(proto.writeHeader(ctx.RequestHeaders()))
	// Pass the frame to execute
	frame := transport.Bytes()
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, called)

	// Reregister the same context
	assert.Error(registry.Register(ctx, cb))

	// Unregister the context
	registry.Unregister(ctx)
	opid, err = getOpID(ctx)
	assert.Nil(err)
	_, ok := registry.(*fRegistry).handlers[opid]
	assert.False(ok)
	// But make sure execute sill returns nil when executing a frame with the
	// same opID (it will just drop the frame)
	assert.Nil(registry.Execute(frame))
	assert.Equal(1, called)

	// Now, register the same context again and ensure the opID is increased.
	assert.Nil(registry.Register(ctx, cb))
	newOpID, err := getOpID(ctx)
	assert.Nil(err)
	assert.True(newOpID > opid)
}

type mockProcessor struct {
	iprot *FProtocol
	oprot *FProtocol
}

func (p *mockProcessor) Process(in, out *FProtocol) error {
	p.iprot = in
	p.oprot = out
	return nil
}

func (p *mockProcessor) AddMiddleware(middleware ServiceMiddleware) {}
