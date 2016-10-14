package frugal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

// frameBufferSize is the number of message frames to buffer on the subscriber.
const frameBufferSize = 5

// FNatsPublisherTransportFactory creates FNatsPublisherTransports.
type FNatsPublisherTransportFactory struct {
	conn  *nats.Conn
}

// NewFNatsPublisherTransportFactory creates an FNatsPublisherTransportFactory using
// the provided NATS connection.
func NewFNatsPublisherTransportFactory(conn *nats.Conn) *FNatsPublisherTransportFactory {
	return &FNatsPublisherTransportFactory{conn: conn}
}

// GetTransport creates a new NATS FPublisherTransport.
func (n *FNatsPublisherTransportFactory) GetTransport() FPublisherTransport {
	return NewNatsFPublisherTransport(n.conn)
}

// fNatsPublisherTransport implements FPublisherTransport.
type fNatsPublisherTransport struct {
	conn         *nats.Conn
	subject      string
	queue        string
	closeChan    chan struct{}
	writeBuffer  *bytes.Buffer
	topicMu      sync.Mutex
	openMu       sync.RWMutex
	isOpen       bool
	sizeBuffer   []byte
}

// NewNatsFPublisherTransport creates a new FPublisherTransport which is used for
// publishing with scopes
func NewNatsFPublisherTransport(conn *nats.Conn) FPublisherTransport {
	return &fNatsPublisherTransport{conn: conn}
}

// LockTopic sets the publish topic and locks the transport for exclusive
// access.
func (n *fNatsPublisherTransport) LockTopic(topic string) error {
	n.topicMu.Lock()
	n.subject = topic
	return nil
}

// UnlockTopic unsets the publish topic and unlocks the transport.
func (n *fNatsPublisherTransport) UnlockTopic() error {
	n.subject = ""
	n.topicMu.Unlock()
	return nil
}

// Open initializes the transport based on whether it's a publisher or
// subscriber. If Open is called before Subscribe, the transport is assumed to
// be a publisher.
func (n *fNatsPublisherTransport) Open() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", n.conn.Status()))
	}

	if n.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	n.writeBuffer = bytes.NewBuffer([]byte{})
	n.sizeBuffer = make([]byte, 4)
	n.isOpen = true
	return nil
}

// IsOpen returns true if the transport is open, false otherwise.
func (n *fNatsPublisherTransport) IsOpen() bool {
	n.openMu.RLock()
	defer n.openMu.RUnlock()
	return n.conn.Status() == nats.CONNECTED && n.isOpen
}

func (n *fNatsPublisherTransport) getClosedConditionError(prefix string) error {
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s NATS client not connected (has status code %d)", prefix, n.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s NATS FPublisherTransport not open", prefix))
}

// Close unsubscribes in the case of a subscriber and clears the buffer in the
// case of a publisher.
func (n *fNatsPublisherTransport) Close() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if !n.isOpen {
		return nil
	}

	n.isOpen = false
	return nil
}

// Read is an invalid operation for publisher transports, so will always
// return an error.
func (n *fNatsPublisherTransport) Read(p []byte) (int, error) {
	return 0, errors.New("publisher: can't call Read")
}

// Write bytes to publish. If buffered bytes exceeds 1MB, ErrTooLarge is
// returned.
func (n *fNatsPublisherTransport) Write(p []byte) (int, error) {
	// Include 4 bytes for frame size.
	if len(p)+n.writeBuffer.Len()+4 > natsMaxMessageSize {
		n.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}

	num, err := n.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Flush publishes the buffered message.
func (n *fNatsPublisherTransport) Flush() error {
	if !n.IsOpen() {
		return n.getClosedConditionError("flush:")
	}
	defer n.writeBuffer.Reset()
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	binary.BigEndian.PutUint32(n.sizeBuffer, uint32(len(data)))
	err := n.conn.Publish(n.formattedSubject(), append(n.sizeBuffer, data...))
	return thrift.NewTTransportExceptionFromError(err)
}

// RemainingBytes returns the number of bytes left to be read. Read is an
// invalid operation though, so this shouldn't be called.
func (n *fNatsPublisherTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used.
}

func (n *fNatsPublisherTransport) formattedSubject() string {
	return fmt.Sprintf("%s%s", frugalPrefix, n.subject)
}


// FNatsSubscriberTransportFactory creates FNatsSubscriberTransports.
type FNatsSubscriberTransportFactory struct {
	conn  *nats.Conn
	queue string
}

// NewFNatsSubscriberTransportFactory creates an FNatsSubscriberTransportFactory using
// the provided NATS connection. Subscribers using this transport will not use
// a queue.
func NewFNatsSubscriberTransportFactory(conn *nats.Conn) *FNatsSubscriberTransportFactory {
	return &FNatsSubscriberTransportFactory{conn: conn}
}

// NewFNatsSubscriberTransportFactoryWithQueue creates an FNatsSubscriberTransportFactory
// using the provided NATS connection. Subscribers using this transport will
// subscribe to the provided queue, forming a queue group. When a queue group
// is formed, only one member receives the message.
func NewFNatsSubscriberTransportFactoryWithQueue(conn *nats.Conn, queue string) *FNatsSubscriberTransportFactory {
	return &FNatsSubscriberTransportFactory{conn: conn, queue: queue}
}

// GetTransport creates a new NATS FSubscriberTransport.
func (n *FNatsSubscriberTransportFactory) GetTransport() FSubscriberTransport {
	return NewNatsFSubscriberTransportWithQueue(n.conn, n.queue)
}

// fNatsSubscriberTransport implements FSubscriberTransport.
type fNatsSubscriberTransport struct {
	conn         *nats.Conn
	subject      string
	queue        string
	callback     FAsyncCallback
	sub          *nats.Subscription
	openMu       sync.RWMutex
	isOpen       bool
	sizeBuffer   []byte
}

// NewNatsFSubscriberTransport creates a new FSubscriberTransport which is used for
// pub/sub. Subscribers using this transport will not use a queue.
func NewNatsFSubscriberTransport(conn *nats.Conn) FSubscriberTransport {
	return &fNatsSubscriberTransport{conn: conn}
}

// NewNatsFSubscriberTransportWithQueue creates a new FSubscriberTransport which is used
// for pub/sub. Subscribers using this transport will subscribe to the provided
// queue, forming a queue group. When a queue group is formed, only one member
// receives the message.
func NewNatsFSubscriberTransportWithQueue(conn *nats.Conn, queue string) FSubscriberTransport {
	return &fNatsSubscriberTransport{conn: conn, queue: queue}
}

// Subscribe sets the subscribe topic and opens the transport.
func (n *fNatsSubscriberTransport) Subscribe(topic string, callback FAsyncCallback) error {
	n.subject = topic
	n.callback = callback
	return n.open()
}

// Open initializes the transport based on whether it's a publisher or
// subscriber. If Open is called before Subscribe, the transport is assumed to
// be a publisher.
func (n *fNatsSubscriberTransport) open() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", n.conn.Status()))
	}

	if n.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	if n.subject == "" {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"cannot subscribe to empty subject")
	}

	sub, err := n.conn.QueueSubscribe(n.formattedSubject(), n.queue, n.handleMessage)
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = sub
	n.isOpen = true
	return nil
}

func (n *fNatsSubscriberTransport) handleMessage(msg *nats.Msg) {
	if len(msg.Data) < 4 {
		logger().Warn("frugal: Discarding invalid scope message frame")
		return
	}
	transport := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(msg.Data[4:])}
	if err := n.callback(transport); err != nil {
		logger().Warn("frugal: error executing callback: ", err)
	}
}

// IsOpen returns true if the transport is open, false otherwise.
func (n *fNatsSubscriberTransport) IsOpen() bool {
	n.openMu.RLock()
	defer n.openMu.RUnlock()
	return n.conn.Status() == nats.CONNECTED && n.isOpen
}

func (n *fNatsSubscriberTransport) getClosedConditionError(prefix string) error {
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.NOT_OPEN,
			fmt.Sprintf("%s NATS client not connected (has status code %d)", prefix, n.conn.Status()))
	}
	return thrift.NewTTransportException(thrift.NOT_OPEN,
		fmt.Sprintf("%s NATS FSubscriberTransport not open", prefix))
}

// Close unsubscribes in the case of a subscriber and clears the buffer in the
// case of a publisher.
func (n *fNatsSubscriberTransport) Unsubscribe() error {
	n.openMu.Lock()
	defer n.openMu.Unlock()
	if !n.isOpen {
		return nil
	}

	if err := n.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = nil
	n.isOpen = false
	return nil
}

func (n *fNatsSubscriberTransport) formattedSubject() string {
	return fmt.Sprintf("%s%s", frugalPrefix, n.subject)
}
