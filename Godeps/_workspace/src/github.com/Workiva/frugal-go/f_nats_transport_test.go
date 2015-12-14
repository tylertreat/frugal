package frugal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTTransport struct {
	mock.Mock
}

func (m *mockTTransport) Open() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) Close() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) Read(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockTTransport) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockTTransport) Flush() error {
	return m.Called().Error(0)
}

func (m *mockTTransport) RemainingBytes() uint64 {
	return m.Called().Get(0).(uint64)
}

func (m *mockTTransport) IsOpen() bool {
	return m.Called().Bool(0)
}

// Ensure Subscribe sets the subject on the underlying transport and calls
// open.
func TestTransportSubscribe(t *testing.T) {
	factory := NewFNatsTransportFactory(nil)
	tr := factory.GetTransport()
	tTransport := new(mockTTransport)
	tr.(*fNatsTransport).tTransport = tTransport
	tTransport.On("Open").Return(nil)

	assert.Nil(t, tr.Subscribe("foo"))
	assert.Equal(t, "foo", tr.(*fNatsTransport).nats.subject)

	tTransport.AssertExpectations(t)
}

// Ensure Unsubscribe calls close on the underlying transport.
func TestTransportUnsubscribe(t *testing.T) {
	factory := NewFNatsTransportFactory(nil)
	tr := factory.GetTransport()
	tTransport := new(mockTTransport)
	tr.(*fNatsTransport).tTransport = tTransport
	tTransport.On("Close").Return(nil)

	assert.Nil(t, tr.Unsubscribe())

	tTransport.AssertExpectations(t)
}

// Ensure PreparePublish sets the subject on the transport.
func TestPreparePublish(t *testing.T) {
	factory := NewFNatsTransportFactory(nil)
	tr := factory.GetTransport()

	tr.PreparePublish("foo")
	assert.Equal(t, "foo", tr.(*fNatsTransport).nats.subject)
}

// Ensure ThriftTransport returns the underlying transport.
func TestThriftTransport(t *testing.T) {
	factory := NewFNatsTransportFactory(nil)
	tr := factory.GetTransport()
	tTransport := new(mockTTransport)
	tr.(*fNatsTransport).tTransport = tTransport

	assert.Equal(t, tTransport, tr.ThriftTransport())
}

// Ensure ApplyProxy sets the underlying transport to that returned by the
// factory.
func TestApplyProxy(t *testing.T) {
	factory := NewFNatsTransportFactory(nil)
	tr := factory.GetTransport()
	tTransport := new(mockTTransport)
	tr.(*fNatsTransport).tTransport = tTransport
	transportFactory := new(mockTTransportFactory)
	proxy := new(mockTTransport)
	transportFactory.On("GetTransport", tTransport).Return(proxy)

	tr.ApplyProxy(transportFactory)

	assert.Equal(t, proxy, tr.(*fNatsTransport).tTransport)
	transportFactory.AssertExpectations(t)
}
