package frugal

import (
	"errors"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFTransport struct {
	mock.Mock
}

func (m *mockFTransport) Subscribe(topic string) error {
	return m.Called(topic).Error(0)
}

func (m *mockFTransport) Unsubscribe() error {
	return m.Called().Error(0)
}

func (m *mockFTransport) PreparePublish(subject string) {
	m.Called(subject)
}

func (m *mockFTransport) ThriftTransport() thrift.TTransport {
	return m.Called().Get(0).(thrift.TTransport)
}

func (m *mockFTransport) ApplyProxy(tr thrift.TTransportFactory) {
	m.Called(tr)
}

// Ensure Unsubscribe calls Unsubscribe on the Transport and returns the
// result.
func TestUnsubscribe(t *testing.T) {
	tr := new(mockFTransport)
	tr.On("Unsubscribe").Return(nil)
	sub := NewSubscription("foo", tr)

	assert.Nil(t, sub.Unsubscribe())

	tr.AssertExpectations(t)

	tr = new(mockFTransport)
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
