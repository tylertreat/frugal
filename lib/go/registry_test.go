package frugal

import (
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
)

func TestClientRegistryRegisterUnregister(t *testing.T) {
	assert := assert.New(t)
	called := false
	cb := func(tr thrift.TTransport) error {
		called = true
		return nil
	}
	registry := NewFClientRegistry()
	ctx := NewFContext("")

	// Register the context for the first time
	assert.Nil(registry.Register(ctx, cb))
	handler, ok := registry.(*clientRegistry).handlers[ctx.opID()]
	assert.True(ok)
	assert.Nil(handler(&mockFTransport{}))
	assert.True(called)

	// Reregister the same context
	assert.Error(registry.Register(ctx, cb))

	// Unregister the context
	registry.Unregister(ctx)
	_, ok = registry.(*clientRegistry).handlers[ctx.opID()]
	assert.False(ok)
}

func TestClientRegistryBadFrame(t *testing.T) {
	assert := assert.New(t)
	registry := NewFClientRegistry()
	assert.Error(registry.Execute([]byte{0}))
}

func TestClientRegistryMissingOpID(t *testing.T) {
	assert := assert.New(t)
	registry := NewFClientRegistry()
	assert.Error(registry.Execute(basicFrame))
}

func TestClientRegistryExecute(t *testing.T) {
	assert := assert.New(t)
	called := false
	cb := func(tr thrift.TTransport) error {
		called = true
		mem := tr.(*thrift.TMemoryBuffer).Bytes()
		assert.Equal(frugalFrame, mem)
		return nil
	}
	registry := NewFClientRegistry()

	// Try executing without registering first
	assert.Nil(registry.Execute(frugalFrame))
	assert.False(called)

	// Register and execute for real
	registry.(*clientRegistry).handlers[0] = cb
	assert.Nil(registry.Execute(frugalFrame))
	assert.True(called)
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

func TestServerRegistry(t *testing.T) {
	assert := assert.New(t)
	processor := &mockProcessor{}
	protocolFactory := NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
	oprot := protocolFactory.GetProtocol(&mockFTransport{})

	registry := NewServerRegistry(processor, protocolFactory, oprot)
	assert.Nil(registry.Execute(frugalFrame))

	ctx, err := processor.iprot.ReadRequestHeader()
	assert.Nil(err)
	assert.Equal(ctx.CorrelationID(), frugalHeaders[cid])
}
