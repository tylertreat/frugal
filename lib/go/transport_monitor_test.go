package frugal

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testTimeout = 10 * time.Millisecond

// Ensure that NewFTransportMonitor returns a non-nil monitor with failure callbacks.
func TestNewFTransportMonitor(t *testing.T) {
	m := NewFTransportMonitor(100, time.Millisecond, time.Minute)
	require.NotNil(t, m)
	assert.NotNil(t, m.ClosedUncleanly)
	assert.NotNil(t, m.ReopenFailed)
}

// Ensure that monitor attempts to handle a clean close.
func TestMonitorCleanClose(t *testing.T) {
	m := &FTransportMonitor{}

	closedChannel := make(chan bool, 1)
	closedChannel <- true

	monitorExited := make(chan struct{})
	go func() {
		m.monitor(nil, closedChannel)
		monitorExited <- struct{}{}
	}()

	select {
	case <-monitorExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that monitor attempts to handle an unclean close.
func TestMonitorUncleanClose(t *testing.T) {
	m := &FTransportMonitor{}

	closedChannel := make(chan bool, 1)
	closedChannel <- false

	monitorExited := make(chan struct{})
	go func() {
		m.monitor(nil, closedChannel)
		monitorExited <- struct{}{}
	}()

	select {
	case <-monitorExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that handleCleanClose deals with a nil callback.
func TestHandleCleanCloseNilCallback(t *testing.T) {
	m := &FTransportMonitor{}

	handleCleanCloseExited := make(chan struct{})
	go func() {
		m.handleCleanClose()
		handleCleanCloseExited <- struct{}{}
	}()

	select {
	case <-handleCleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that handleCleanClose invokes a non-nil callback.
func TestHandleCleanClose(t *testing.T) {
	closedCleanlyInvoked := make(chan struct{}, 1)
	m := &FTransportMonitor{
		ClosedCleanly: func() { closedCleanlyInvoked <- struct{}{} },
	}

	handleCleanCloseExited := make(chan struct{})
	go func() {
		m.handleCleanClose()
		handleCleanCloseExited <- struct{}{}
	}()

	select {
	case <-closedCleanlyInvoked:
	case <-time.After(testTimeout):
		t.Fail()
	}

	select {
	case <-handleCleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that handleUncleanClose returns false with a nil callback.
func TestHandleUncleanCloseNilCallback(t *testing.T) {
	m := &FTransportMonitor{}

	handleUncleanCloseExited := make(chan struct{})
	go func() {
		assert.False(t, m.handleUncleanClose(nil))
		handleUncleanCloseExited <- struct{}{}
	}()

	select {
	case <-handleUncleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that handleUncleanClose returns false with a non-nil callback that instructs not to retry.
func TestHandleUncleanCloseCallbackNoRetry(t *testing.T) {
	closedUncleanlyInvoked := make(chan struct{})
	m := &FTransportMonitor{
		ClosedUncleanly: func() (reopen bool, wait time.Duration) {
			closedUncleanlyInvoked <- struct{}{}
			return false, 0
		},
	}

	handleUncleanCloseExited := make(chan struct{})
	go func() {
		assert.False(t, m.handleUncleanClose(nil))
		handleUncleanCloseExited <- struct{}{}
	}()

	select {
	case <-closedUncleanlyInvoked:
	case <-time.After(testTimeout):
		t.Fail()
	}

	select {
	case <-handleUncleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
}

// Ensure that handleUncleanClose attempts re-open with a non-nil callback that instructs to retry.
func TestHandleUncleanCloseCallback(t *testing.T) {
	closedUncleanlyInvoked := make(chan struct{})
	m := &FTransportMonitor{
		ClosedUncleanly: func() (reopen bool, wait time.Duration) {
			closedUncleanlyInvoked <- struct{}{}
			return true, 0
		},
	}
	mft := &mockFTransport{}
	mft.On("Open").Return(nil).Once()

	handleUncleanCloseExited := make(chan struct{})
	go func() {
		assert.True(t, m.handleUncleanClose(mft))
		handleUncleanCloseExited <- struct{}{}
	}()

	select {
	case <-closedUncleanlyInvoked:
	case <-time.After(testTimeout):
		t.Fail()
	}

	select {
	case <-handleUncleanCloseExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen returns false with an open error and a nil callback
func TestAttemptReopenNilCallback(t *testing.T) {
	m := &FTransportMonitor{}

	mft := &mockFTransport{}
	mft.On("Open").Return(errors.New("Tears of a thousand children")).Once()

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.False(t, m.attemptReopen(0, mft))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen returns true with an open success.
func TestAttemptReopenSuccess(t *testing.T) {
	m := &FTransportMonitor{}

	mft := &mockFTransport{}
	mft.On("Open").Return(nil).Once()

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.True(t, m.attemptReopen(0, mft))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen retries if the callback instructs to do so.
func TestAttemptReopenFailRetrySucceed(t *testing.T) {
	reopenFailedInvoked := make(chan struct{})
	m := &FTransportMonitor{
		ReopenFailed: func(uint, time.Duration) (bool, time.Duration) {
			reopenFailedInvoked <- struct{}{}
			return true, 0
		},
	}

	mft := &mockFTransport{}
	mft.On("Open").Return(errors.New("Your potted plant withered and died")).Once()
	mft.On("Open").Return(nil).Once()

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.True(t, m.attemptReopen(0, mft))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-reopenFailedInvoked:
	case <-time.After(testTimeout):
		t.Fail()
	}

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	mft.AssertExpectations(t)
}

// Ensure that attemptReopen returns false and does not retry if the callback instructs it not to do so.
func TestAttemptReopenFailNoRetry(t *testing.T) {
	reopenFailedInvoked := make(chan struct{})
	m := &FTransportMonitor{
		ReopenFailed: func(uint, time.Duration) (bool, time.Duration) {
			reopenFailedInvoked <- struct{}{}
			return false, 0
		},
	}

	mft := &mockFTransport{}
	mft.On("Open").Return(errors.New("Shattered dreams of a PhD student")).Once()

	attemptReopenExited := make(chan struct{})
	go func() {
		assert.False(t, m.attemptReopen(0, mft))
		attemptReopenExited <- struct{}{}
	}()

	select {
	case <-reopenFailedInvoked:
	case <-time.After(testTimeout):
		t.Fail()
	}

	select {
	case <-attemptReopenExited:
	case <-time.After(testTimeout):
		t.Fail()
	}
	mft.AssertExpectations(t)
}

type mockFTransport struct {
	mock.Mock
}

func (m *mockFTransport) Open() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) IsOpen() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockFTransport) RemainingBytes() (num_bytes uint64) {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *mockFTransport) Flush() (err error) {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockFTransport) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockFTransport) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockFTransport) SetRegistry(fr FRegistry) {
	m.Called(fr)
}

func (m *mockFTransport) Register(fc *FContext, fac FAsyncCallback) error {
	args := m.Called(fc, fac)
	return args.Error(0)
}

func (m *mockFTransport) Unregister(fc *FContext) {
	m.Called(fc)
}

func (m *mockFTransport) Closed() <-chan bool {
	args := m.Called()
	return args.Get(0).(<-chan bool)
}

func (m *mockFTransport) SetMonitor(mon *FTransportMonitor) {
	m.Called(mon)
}
