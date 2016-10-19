package frugal

import (
	"bytes"
	"encoding/binary"
	"errors"

	"git.apache.org/thrift.git/lib/go/thrift"
)

const (
	// REQUEST_TOO_LARGE is a TTransportException error type indicating the
	// request exceeded the size limit.
	REQUEST_TOO_LARGE = 100

	// RESPONSE_TOO_LARGE is a TTransportException error type indicating the
	// response exceeded the size limit.
	RESPONSE_TOO_LARGE = 101
)

// ErrTransportClosed is returned by service calls when the transport is
// unexpectedly closed, perhaps as a result of the transport entering an
// invalid state. If this is returned, the transport should be reinitialized.
var ErrTransportClosed = errors.New("frugal: transport was unexpectedly closed")

// ErrTooLarge is returned when attempting to write a message which exceeds the
// transport's message size limit.
var ErrTooLarge = thrift.NewTTransportException(REQUEST_TOO_LARGE,
	"request was too large for the transport")

// IsErrTooLarge indicates if the given error is an ErrTooLarge.
func IsErrTooLarge(err error) bool {
	if err == ErrTooLarge {
		return true
	}
	if e, ok := err.(thrift.TTransportException); ok {
		return e.TypeId() == REQUEST_TOO_LARGE || e.TypeId() == RESPONSE_TOO_LARGE
	}
	return false
}

// FPublisherTransportFactory produces FPublisherTransports and is typically
// used by an FScopeProvider.
type FPublisherTransportFactory interface {
	GetTransport() FPublisherTransport
}

// FPublisherTransport extends Thrift's TTransport and is used exclusively for
// pub/sub scopes. Publishers use it to publish to a topic.
type FPublisherTransport interface {
	thrift.TTransport

	// LockTopic sets the publish topic and locks the transport for exclusive
	// access.
	LockTopic(string) error

	// UnlockTopic unsets the publish topic and unlocks the transport.
	UnlockTopic() error
}

// FSubscriberTransportFactory produces FSubscriberTransports and is typically
// used by an FScopeProvider.
type FSubscriberTransportFactory interface {
	GetTransport() FSubscriberTransport
}

// FSubscriberTransport extends Thrift's TTransport and is used exclusively for
// pub/sub scopes. Subscribers use it to subscribe to a pub/sub topic.
type FSubscriberTransport interface {
	// Subscribe sets the subscribe topic and opens the transport.
	Subscribe(string, FAsyncCallback) error

	// Unsubscribe unsubscribes from the topic and closes the transport.
	Unsubscribe() error
}

// FTransport is Frugal's equivalent of Thrift's TTransport. FTransport is
// comparable to Thrift's TTransport in that it represents the transport layer
// for frugal clients. However, frugal is callback based and sends only framed
// data. Due to this, instead of read, write, and flush methods, FTransport has
// a send method that sends framed frugal messages. To handle callback data, an
// FTransport also has an FRegistry, so it provides methods for registering
// and unregistering an FAsyncCallback to an FContext.
type FTransport interface {
	// Register a callback for the given Context.
	Register(*FContext, FAsyncCallback) error

	// Unregister a callback for the given Context.
	Unregister(*FContext)

	// SetMonitor starts a monitor that can watch the health of, and reopen,
	// the transport.
	SetMonitor(FTransportMonitor)

	// Closed channel receives the cause of an FTransport close (nil if clean
	// close).
	Closed() <-chan error

	// Open prepares the transport to send data.
	Open() error

	// IsOpen returns true if the transport is open, false otherwise.
	IsOpen() bool

	// Close closes the transport.
	Close() error

	// Send transmits the given data.
	Send([]byte) error

	// GetRequestSizeLimit returns the maximum number of bytes that can be
	// transmitted. Returns a non-positive number to indicate an unbounded
	// allowable size.
	GetRequestSizeLimit() uint
}

// FTransportFactory produces FTransports by wrapping a provided TTransport.
type FTransportFactory interface {
	GetTransport(tr thrift.TTransport) FTransport
}

type fBaseTransport struct {
	requestSizeLimit uint
	writeBuffer      bytes.Buffer
	registry         FRegistry
	closed           chan error
}

// Initialize a new fBaseTransport
func newFBaseTransport(requestSizeLimit uint) *fBaseTransport {
	return &fBaseTransport{
		requestSizeLimit: requestSizeLimit,
		registry: NewFRegistry(),
	}
}

// Intialize the close channels
func (f *fBaseTransport) Open() {
	f.closed = make(chan error, 1)
}

// Close the close channels
func (f *fBaseTransport) Close(cause error) {
	select {
	case f.closed <- cause:
	default:
		logger().Warnf("frugal: unable to put close error '%s' on fBaseTransport closed channel", cause)
	}
	close(f.closed)
}

// Execute a frugal frame (NOTE: this frame must include the frame size).
func (f *fBaseTransport) ExecuteFrame(frame []byte) error {
	return f.registry.Execute(frame[4:])
}

// Register a callback for the given Context. Only called by generated code.
func (f *fBaseTransport) Register(ctx *FContext, callback FAsyncCallback) error {
	return f.registry.Register(ctx, callback)
}

// Unregister a callback for the given Context. Only called by generated code.
func (f *fBaseTransport) Unregister(ctx *FContext) {
	f.registry.Unregister(ctx)
}

// Closed channel is closed when the FTransport is closed.
func (f *fBaseTransport) Closed() <-chan error {
	return f.closed
}

func prependFrameSize(buf []byte) []byte {
	frame := make([]byte, 4)
	binary.BigEndian.PutUint32(frame, uint32(len(buf)))
	return append(frame, buf...)
}
