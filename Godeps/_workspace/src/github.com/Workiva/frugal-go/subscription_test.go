package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) Subscribe(topic string) error {
	return m.Called(topic).Error(0)
}

func (m *mockTransport) Unsubscribe() error {
	return m.Called().Error(0)
}

func (m *mockTransport) PreparePublish(subject string) {
	m.Called(subject)
}

func (m *mockTransport) ThriftTransport() thrift.TTransport {
	return m.Called().Get(0).(thrift.TTransport)
}

func (m *mockTransport) ApplyProxy(tr thrift.TTransportFactory) {
	m.Called(tr)
}

// Ensure Unsubscribe calls Unsubscribe on the Transport and returns the
// result.
func TestUnsubscribe(t *testing.T) {
	tr := new(mockTransport)
	tr.On("Unsubscribe").Return(nil)
	sub := NewSubscription("foo", tr)

	assert.Nil(t, sub.Unsubscribe())

	tr.AssertExpectations(t)

	tr = new(mockTransport)
	err := errors.New("error")
	tr.On("Unsubscribe").Return(err)
	sub = NewSubscription("foo", tr)

	assert.Equal(t, err, sub.Unsubscribe())

	tr.AssertExpectations(t)
}

// Ensure Signal sends the error on the errors channel and closes the channel.
func TestSignal(t *testing.T) {
	sub := NewSubscription("foo", nil)
	c := sub.Error()

	err := errors.New("error")
	sub.Signal(err)

	assert.Equal(t, err, <-c)
}
