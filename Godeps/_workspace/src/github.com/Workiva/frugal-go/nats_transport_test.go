package frugal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockThriftTransport struct {
	mock.Mock
}

func (m *mockThriftTransport) Open() error {
	return m.Called().Error(0)
}

func (m *mockThriftTransport) Close() error {
	return m.Called().Error(0)
}

func (m *mockThriftTransport) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockThriftTransport) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockThriftTransport) Flush() error {
	return m.Called().Error(0)
}

func (m *mockThriftTransport) RemainingBytes() uint64 {
	return m.Called().Get(0).(uint64)
}

func (m *mockThriftTransport) IsOpen() bool {
	return m.Called().Bool(0)
}

// Ensure Subscribe sets the subject on the underlying transport and calls
// open.
func TestTransportSubscribe(t *testing.T) {
	factory := NewNATSTransportFactory(nil)
	tr := factory.GetTransport()
	thriftTransport := new(mockThriftTransport)
	tr.(*natsTransport).thriftTransport = thriftTransport
	thriftTransport.On("Open").Return(nil)

	assert.Nil(t, tr.Subscribe("foo"))
	assert.Equal(t, "foo", tr.(*natsTransport).nats.subject)

	thriftTransport.AssertExpectations(t)
}

// Ensure Unsubscribe calls close on the underlying transport.
func TestTransportUnsubscribe(t *testing.T) {
	factory := NewNATSTransportFactory(nil)
	tr := factory.GetTransport()
	thriftTransport := new(mockThriftTransport)
	tr.(*natsTransport).thriftTransport = thriftTransport
	thriftTransport.On("Close").Return(nil)

	assert.Nil(t, tr.Unsubscribe())

	thriftTransport.AssertExpectations(t)
}

// Ensure PreparePublish sets the subject on the transport.
func TestPreparePublish(t *testing.T) {
	factory := NewNATSTransportFactory(nil)
	tr := factory.GetTransport()

	tr.PreparePublish("foo")
	assert.Equal(t, "foo", tr.(*natsTransport).nats.subject)
}

// Ensure ThriftTransport returns the underlying transport.
func TestThriftTransport(t *testing.T) {
	factory := NewNATSTransportFactory(nil)
	tr := factory.GetTransport()
	thriftTransport := new(mockThriftTransport)
	tr.(*natsTransport).thriftTransport = thriftTransport

	assert.Equal(t, thriftTransport, tr.ThriftTransport())
}

// Ensure ApplyProxy sets the underlying transport to that returned by the
// factory.
func TestApplyProxy(t *testing.T) {
	factory := NewNATSTransportFactory(nil)
	tr := factory.GetTransport()
	thriftTransport := new(mockThriftTransport)
	tr.(*natsTransport).thriftTransport = thriftTransport
	transportFactory := new(mockThriftTransportFactory)
	proxy := new(mockThriftTransport)
	transportFactory.On("GetTransport", thriftTransport).Return(proxy)

	tr.ApplyProxy(transportFactory)

	assert.Equal(t, proxy, tr.(*natsTransport).thriftTransport)
	transportFactory.AssertExpectations(t)
}
