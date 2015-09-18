package frugal

import (
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/stretchr/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransportFactory struct {
	mock.Mock
}

func (m *mockTransportFactory) GetTransport() Transport {
	return m.Called().Get(0).(Transport)
}

type mockThriftTransportFactory struct {
	mock.Mock
}

func (m *mockThriftTransportFactory) GetTransport(tr thrift.TTransport) thrift.TTransport {
	return m.Called(tr).Get(0).(thrift.TTransport)
}

type mockProtocolFactory struct {
	mock.Mock
}

func (m *mockProtocolFactory) GetProtocol(tr thrift.TTransport) thrift.TProtocol {
	return m.Called(tr).Get(0).(thrift.TProtocol)
}

// Ensure New returns the Transport with the TransportFactory applied and the
// expected protocol.
func TestNew(t *testing.T) {
	assert := assert.New(t)
	transportFactory := new(mockTransportFactory)
	thriftTransportFactory := new(mockThriftTransportFactory)
	protocolFactory := new(mockProtocolFactory)
	thriftTransport, _ := thrift.NewTSocket("localhost:8000")
	protocol := thrift.NewTSimpleJSONProtocol(thriftTransport)
	transport := new(mockTransport)
	transport.On("ApplyProxy", thriftTransportFactory).Return()
	transport.On("ThriftTransport").Return(thriftTransport)
	transportFactory.On("GetTransport").Return(transport)
	protocolFactory.On("GetProtocol", thriftTransport).Return(protocol)
	p := NewProvider(transportFactory, thriftTransportFactory, protocolFactory)

	tr, proto := p.New()

	assert.Equal(transport, tr)
	assert.Equal(protocol, proto)

	transportFactory.AssertExpectations(t)
	thriftTransportFactory.AssertExpectations(t)
	protocolFactory.AssertExpectations(t)
	transport.AssertExpectations(t)
}
