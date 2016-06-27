package frugal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/go-nats-streaming"
)

type FNatsStreamingScopeTransportFactory struct {
	conn    stan.Conn
	queue   string
	options []stan.SubscriptionOption
}

func NewFNatsStreamingScopeTransportFactory(conn stan.Conn,
	options ...stan.SubscriptionOption) *FNatsStreamingScopeTransportFactory {
	return &FNatsStreamingScopeTransportFactory{
		conn:    conn,
		options: options,
	}
}

func NewFNatsStreamingScopeTransportFactoryWithQueue(conn stan.Conn, queue string,
	options ...stan.SubscriptionOption) *FNatsStreamingScopeTransportFactory {
	return &FNatsStreamingScopeTransportFactory{
		conn:    conn,
		queue:   queue,
		options: options,
	}
}

func (f *FNatsStreamingScopeTransportFactory) GetTransport() FScopeTransport {
	return NewFNatsStreamingScopeTransportWithQueue(f.conn, f.queue, f.options...)
}

type fNatsStreamingScopeTransport struct {
	conn         stan.Conn
	queue        string
	options      []stan.SubscriptionOption
	subscriber   bool
	topicMu      sync.Mutex
	topic        string
	openMu       sync.RWMutex
	isOpen       bool
	writeBuffer  *bytes.Buffer
	sizeBuffer   []byte
	frameBuffer  chan []byte
	closeChan    chan struct{}
	sub          stan.Subscription
	currentFrame []byte
}

func NewFNatsStreamingScopeTransport(conn stan.Conn,
	options ...stan.SubscriptionOption) FScopeTransport {
	return &fNatsStreamingScopeTransport{
		conn:    conn,
		options: options,
	}
}

func NewFNatsStreamingScopeTransportWithQueue(conn stan.Conn, queue string,
	options ...stan.SubscriptionOption) FScopeTransport {
	return &fNatsStreamingScopeTransport{
		conn:    conn,
		queue:   queue,
		options: options,
	}
}

func (f *fNatsStreamingScopeTransport) LockTopic(topic string) error {
	if f.subscriber {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"subscriber cannot unlock topic")
	}
	f.topicMu.Lock()
	f.topic = topic
	return nil
}

func (f *fNatsStreamingScopeTransport) UnlockTopic() error {
	if f.subscriber {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"subscriber cannot unlock topic")
	}
	f.topic = ""
	f.topicMu.Unlock()
	return nil
}

func (f *fNatsStreamingScopeTransport) Subscribe(topic string) error {
	f.subscriber = true
	f.topic = topic
	return f.Open()
}

func (f *fNatsStreamingScopeTransport) Open() error {
	f.openMu.Lock()
	defer f.openMu.Unlock()
	// TODO: check conn status
	if f.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS streaming transport already open")
	}

	if !f.subscriber {
		f.writeBuffer = bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize))
		f.sizeBuffer = make([]byte, 4)
		f.isOpen = true
		return nil
	}

	if f.topic == "" {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"cannot subscribe to empty subject")
	}

	f.closeChan = make(chan struct{})
	f.frameBuffer = make(chan []byte, frameBufferSize)

	sub, err := f.conn.QueueSubscribe(f.formattedSubject(), f.queue, f.handleMessage, f.options...)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = sub
	f.isOpen = true
	return nil
}

func (f *fNatsStreamingScopeTransport) handleMessage(msg *stan.Msg) {
	if len(msg.Data) < 4 {
		log.Warn("frugal: Discarding invalid scope message frame")
		return
	}
	// Discard frame size.
	select {
	case f.frameBuffer <- msg.Data[4:]:
	case <-f.closeChan:
	}
}

func (f *fNatsStreamingScopeTransport) IsOpen() bool {
	f.openMu.RLock()
	defer f.openMu.RUnlock()
	// TODO: check conn status
	return f.isOpen
}

func (f *fNatsStreamingScopeTransport) Close() error {
	f.openMu.Lock()
	defer f.openMu.Unlock()
	if !f.isOpen {
		return nil
	}

	if !f.subscriber {
		f.isOpen = false
		return nil
	}

	if err := f.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	f.sub = nil
	close(f.closeChan)
	f.isOpen = false
	return nil
}

func (f *fNatsStreamingScopeTransport) Read(p []byte) (int, error) {
	if !f.IsOpen() {
		return 0, thrift.NewTTransportExceptionFromError(io.EOF)
	}
	if len(f.currentFrame) == 0 {
		select {
		case frame := <-f.frameBuffer:
			f.currentFrame = frame
		case <-f.closeChan:
			return 0, thrift.NewTTransportExceptionFromError(io.EOF)
		}
	}
	num := copy(p, f.currentFrame)
	// TODO: We could be more efficient here. If the provided buffer isn't
	// full, we could attempt to get the next frame.

	f.currentFrame = f.currentFrame[num:]
	return num, nil
}

// DiscardFrame discards the current message frame the transport is reading, if
// any. After calling this, a subsequent call to Read will read from the next
// frame. This must be called from the same goroutine as the goroutine calling
// Read.
func (f *fNatsStreamingScopeTransport) DiscardFrame() {
	f.currentFrame = nil
}

// Write bytes to publish. If buffered bytes exceeds 1MB, ErrTooLarge is
// returned.
func (f *fNatsStreamingScopeTransport) Write(p []byte) (int, error) {
	if !f.IsOpen() {
		return 0, f.getClosedConditionError("write:")
	}

	// Include 4 bytes for frame size.
	if len(p)+f.writeBuffer.Len()+4 > natsMaxMessageSize {
		f.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}

	num, err := f.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Flush publishes the buffered message. Returns ErrTooLarge if the buffered
// message exceeds 1MB.
func (f *fNatsStreamingScopeTransport) Flush() error {
	if !f.IsOpen() {
		return f.getClosedConditionError("flush:")
	}
	defer f.writeBuffer.Reset()
	data := f.writeBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	// Include 4 bytes for frame size.
	if len(data)+4 > natsMaxMessageSize {
		return ErrTooLarge
	}
	binary.BigEndian.PutUint32(f.sizeBuffer, uint32(len(data)))
	err := f.conn.Publish(f.formattedSubject(), append(f.sizeBuffer, data...))
	return thrift.NewTTransportExceptionFromError(err)
}

func (f *fNatsStreamingScopeTransport) RemainingBytes() uint64 {
	return ^uint64(0)
}

func (f *fNatsStreamingScopeTransport) formattedSubject() string {
	return fmt.Sprintf("%s%s", frugalPrefix, f.topic)
}

func (f *fNatsStreamingScopeTransport) getClosedConditionError(prefix string) error {
	// TODO: check conn status
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s NATS streaming FScopeTransport not open", prefix))
}
