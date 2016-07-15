package frugal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

// statelessNatsFTransport implements FTransport and, until the next major
// release, thrift.TTransport that may be wrapped with fMuxTransport
// (DEPRECATED). This is a "stateless" transport in the sense that there is no
// connection with a server. A request is simply published to a subject and
// responses are received on another subject. This assumes requests/responses
// fit within a single NATS message.
type statelessNatsFTransport struct {
	*fBaseTransport
	conn          *nats.Conn
	subject       string
	inbox         string
	requestBuffer *bytes.Buffer
	sub           *nats.Subscription

	// TODO: Remove these with 2.0
	isTTransport bool
	frameBuffer  chan []byte
	currentFrame []byte
	closeChan    chan struct{}
	mu           sync.RWMutex
}

// NewStatelessNatsTTransport returns a new Thrift TTransport which uses the
// NATS messaging system as the underlying transport. Unlike the TTransport
// created by NewNatsServiceTTransport, this TTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply published to a subject and responses are received on another
// subject. This requires requests and responses to fit within a single NATS
// message.
// DEPRECATED - Use NewNatsFTransport to create an FTransport directly.
// TODO: Remove this with 2.0
func NewStatelessNatsTTransport(conn *nats.Conn, subject, inbox string) thrift.TTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &statelessNatsFTransport{
		fBaseTransport: newFBaseTransport(natsMaxMessageSize),
		conn:           conn,
		subject:        subject,
		inbox:          inbox,
		frameBuffer:    make(chan []byte, frameBufferSize),
		requestBuffer:  bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize)),
		isTTransport:   true,
	}
}

// NewStatelessNatsFTransport returns a new FTransport which uses the
// NATS messaging system as the underlying transport. Unlike the FTransport
// created by NewNatsServiceFTransport, this FTransport is stateless in that
// there is no connection maintained between the client and server. A request
// is simply published to a subject and responses are received on another
// subject. This requires requests and responses to fit within a single NATS
// message.
func NewStatelessNatsFTransport(conn *nats.Conn, subject, inbox string) FTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &statelessNatsFTransport{
		fBaseTransport: newFBaseTransport(natsMaxMessageSize),
		conn:           conn,
		subject:        subject,
		inbox:          inbox,
		frameBuffer:    make(chan []byte, frameBufferSize),
		requestBuffer:  bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize)),
	}
}

// Open subscribes to the configured inbox subject.
func (f *statelessNatsFTransport) Open() error {
	// TODO: Remove locking with 2.0
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", f.conn.Status()))
	}
	if f.sub != nil {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	handler := f.handler
	// TODO: Remove this with 2.0
	if f.isTTransport {
		handler = f.tTransportHandler
	}

	sub, err := f.conn.Subscribe(f.inbox, handler)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = sub

	// TODO: Remove this with 2.0
	f.closeChan = make(chan struct{})
	return nil
}

// handler receives a NATS message and executes the frame
func (f *statelessNatsFTransport) handler(msg *nats.Msg) {
	if err := f.fBaseTransport.Execute(msg.Data); err != nil {
		log.Warn("Could not execute frame", err)
	}
}

// tTransportHandler receives a NATS message and places it on the frame buffer
// for reading.
// TODO: Remove this with 2.0
func (f *statelessNatsFTransport) tTransportHandler(msg *nats.Msg) {
	select {
	case f.frameBuffer <- msg.Data:
	case <-f.closeChan:
	}
}

func (f *statelessNatsFTransport) IsOpen() bool {
	// TODO: Remove locking with 2.0
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.sub != nil && f.conn.Status() == nats.CONNECTED
}

// Close unsubscribes from the inbox subject.
func (f *statelessNatsFTransport) Close() error {
	// TODO: Remove locking with 2.0
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.sub == nil {
		return nil
	}
	if err := f.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = nil

	// TODO: Remove this with 2.0
	close(f.closeChan)

	return nil
}

// Read up to len(buf) bytes into buf.
// TODO: This should just return an error with 2.0
func (f *statelessNatsFTransport) Read(buf []byte) (int, error) {
	if !f.isTTransport {
		return 0, errors.New("Cannot read on FTransport")
	}

	// TODO: Remove all read logic with 2.0
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("read:")
	}
	if len(f.currentFrame) == 0 {
		select {
		case frame := <-f.frameBuffer:
			f.currentFrame = frame
		case <-f.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(buf, f.currentFrame)
	f.currentFrame = f.currentFrame[num:]
	return num, nil
}

// Write the bytes to a buffer. Returns ErrTooLarge if the buffer exceeds 1MB.
func (f *statelessNatsFTransport) Write(buf []byte) (int, error) {
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("write:")
	}
	return f.fBaseTransport.Write(buf)
}

// Flush sends the buffered bytes over NATS.
func (f *statelessNatsFTransport) Flush() error {
	if !f.IsOpen() {
		return f.getClosedConditionError("flush:")
	}
	data := f.fBaseTransport.GetRequestBytes()
	if len(data) == 0 {
		return nil
	}

	// TODO: Remove this check in 2.0
	if !f.isTTransport {
		data = prependFrameSize(data)
	}

	err := f.conn.PublishRequest(f.subject, f.inbox, data)
	return thrift.NewTTransportExceptionFromError(err)
}

// This is a no-op for statelessNatsFTransport
func (f *statelessNatsFTransport) SetMonitor(monitor FTransportMonitor) {
}

func (f *statelessNatsFTransport) getClosedConditionError(prefix string) error {
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s stateless NATS client not connected (has status code %d)", prefix, f.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s stateless NATS service TTransport not open", prefix))
}
