package frugal

import (
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFTransportFactory struct {
	mock.Mock
}

func (m *mockFTransportFactory) GetTransport() FTransport {
	return m.Called().Get(0).(FTransport)
}

type mockTTransportFactory struct {
	mock.Mock
}

func (m *mockTTransportFactory) GetTransport(tr thrift.TTransport) thrift.TTransport {
	return m.Called(tr).Get(0).(thrift.TTransport)
}

type mockProtocolFactory struct {
	mock.Mock
}

func (m *mockProtocolFactory) GetProtocol(tr thrift.TTransport) thrift.TProtocol {
	return m.Called(tr).Get(0).(thrift.TProtocol)
}

// Ensure New returns the FTransport with the FTransportFactory applied and the
// expected protocol.
func TestNew(t *testing.T) {
	assert := assert.New(t)
	fTransportFactory := new(mockFTransportFactory)
	tTransportFactory := new(mockTTransportFactory)
	protocolFactory := new(mockProtocolFactory)
	thriftTransport, _ := thrift.NewTSocket("localhost:8000")
	protocol := thrift.NewTSimpleJSONProtocol(thriftTransport)
	transport := new(mockFTransport)
	transport.On("ApplyProxy", tTransportFactory).Return()
	transport.On("ThriftTransport").Return(thriftTransport)
	fTransportFactory.On("GetTransport").Return(transport)
	protocolFactory.On("GetProtocol", thriftTransport).Return(protocol)
	p := NewProvider(fTransportFactory, tTransportFactory, protocolFactory)

	tr, proto := p.New()

	assert.Equal(transport, tr)
	assert.Equal(protocol, proto)

	fTransportFactory.AssertExpectations(t)
	tTransportFactory.AssertExpectations(t)
	protocolFactory.AssertExpectations(t)
	transport.AssertExpectations(t)
}
