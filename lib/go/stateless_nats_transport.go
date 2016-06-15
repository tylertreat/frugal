package frugal

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

type statelessNatsTTransport struct {
	conn          *nats.Conn
	subject       string
	inbox         string
	frameBuffer   chan []byte
	currentFrame  []byte
	requestBuffer *bytes.Buffer
	sub           *nats.Subscription
	mu            sync.RWMutex
	closeChan     chan struct{}
}

func NewStatelessNatsTTransport(conn *nats.Conn, subject, inbox string) thrift.TTransport {
	if inbox == "" {
		inbox = nats.NewInbox()
	}
	return &statelessNatsTTransport{
		conn:          conn,
		subject:       subject,
		inbox:         inbox,
		frameBuffer:   make(chan []byte, frameBufferSize),
		requestBuffer: bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize)),
	}
}

func (f *statelessNatsTTransport) Open() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", f.conn.Status()))
	}
	if f.sub != nil {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}
	sub, err := f.conn.Subscribe(f.inbox, f.handler)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = sub
	f.closeChan = make(chan struct{})
	return nil
}

func (f *statelessNatsTTransport) handler(msg *nats.Msg) {
	select {
	case f.frameBuffer <- msg.Data:
	case <-f.closeChan:
	}
}

func (f *statelessNatsTTransport) IsOpen() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sub != nil && f.conn.Status() == nats.CONNECTED
}

func (f *statelessNatsTTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sub == nil {
		return nil
	}
	if err := f.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = nil
	close(f.closeChan)
	return nil
}

func (f *statelessNatsTTransport) Read(buf []byte) (int, error) {
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("read:")
	}
	if f.currentFrame == nil {
		select {
		case frame := <-f.frameBuffer:
			f.currentFrame = frame
		case <-f.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(buf, f.currentFrame)
	// TODO: We could be more efficient here. If the provided buffer isn't
	// full, we could attempt to get the next frame.

	f.currentFrame = f.currentFrame[num:]
	if len(f.currentFrame) == 0 {
		// The entire frame was copied, clear it.
		f.currentFrame = nil
	}
	return num, nil
}

func (f *statelessNatsTTransport) Write(buf []byte) (int, error) {
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("write:")
	}
	if len(buf)+f.requestBuffer.Len() > natsMaxMessageSize {
		f.requestBuffer.Reset()
		return 0, ErrTooLarge
	}
	num, err := f.requestBuffer.Write(buf)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (f *statelessNatsTTransport) Flush() error {
	if !f.IsOpen() {
		return f.getClosedConditionError("flush:")
	}
	defer f.requestBuffer.Reset()
	data := f.requestBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	if len(data) > natsMaxMessageSize {
		return ErrTooLarge
	}
	err := f.conn.PublishRequest(f.subject, f.inbox, data)
	return thrift.NewTTransportExceptionFromError(err)
}

func (f *statelessNatsTTransport) RemainingBytes() uint64 {
	return ^uint64(0)
}

func (f *statelessNatsTTransport) getClosedConditionError(prefix string) error {
	if f.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s stateless NATS client not connected (has status code %d)", prefix, f.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s stateless NATS service TTransport not open", prefix))
}
