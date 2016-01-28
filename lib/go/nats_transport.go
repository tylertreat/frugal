package frugal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

const (
	// NATS limits messages to 1MB.
	natsMaxMessageSize = 1024 * 1024
	disconnect         = "DISCONNECT"
)

// natsServiceTTransport implements thrift.TTransport.
type natsServiceTTransport struct {
	conn              *nats.Conn
	listenTo          string
	writeTo           string
	reader            *io.PipeReader
	writer            *io.PipeWriter
	writeBuffer       *bytes.Buffer
	sub               *nats.Subscription
	heartbeatSub      *nats.Subscription
	heartbeatListen   string
	heartbeatReply    string
	heartbeatInterval time.Duration
	recvHeartbeat     chan struct{}
	closed            chan struct{}
	isOpen            bool
	mutex             sync.RWMutex
	connectSubject    string
	connectTimeout    time.Duration
}

// NewNatsServiceTTransport returns a new thrift TTransport which uses
// the NATS messaging system as the underlying transport. It performs a
// handshake with a server listening on the given NATS subject upon open.
// This TTransport can only be used with FNatsServer.
func NewNatsServiceTTransport(conn *nats.Conn, subject string,
	timeout time.Duration) thrift.TTransport {

	return &natsServiceTTransport{
		conn:           conn,
		connectSubject: subject,
		connectTimeout: timeout,
	}
}

// newNatsServiceTTransportServer returns a new thrift TTransport which uses
// the NATS messaging system as the underlying transport. This TTransport can
// only be used with FNatsServer.
func newNatsServiceTTransportServer(conn *nats.Conn, listenTo, writeTo string) thrift.TTransport {
	return &natsServiceTTransport{
		conn:     conn,
		listenTo: listenTo,
		writeTo:  writeTo,
	}
}

// Open handshakes with the server (if this is a client transport) initializes
// the write buffer and reader/writer pipe, subscribes to the specified
// subject, and starts heartbeating.
func (n *natsServiceTTransport) Open() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.conn.Status() != nats.CONNECTED {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			fmt.Sprintf("frugal: NATS not connected, has status %d", n.conn.Status()))
	}

	if n.isOpen {
		return thrift.NewTTransportException(thrift.ALREADY_OPEN, "frugal: NATS transport already open")
	}

	// Handshake if this is a client.
	if n.connectSubject != "" {
		if err := n.handshake(); err != nil {
			return thrift.NewTTransportExceptionFromError(err)
		}
	}

	if n.listenTo == "" || n.writeTo == "" {
		return thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"frugal: listenTo and writeTo cannot be empty")
	}

	n.closed = make(chan struct{})
	n.writeBuffer = bytes.NewBuffer(make([]byte, 0, natsMaxMessageSize))

	n.reader, n.writer = io.Pipe()

	sub, err := n.conn.Subscribe(n.listenTo, func(msg *nats.Msg) {
		if msg.Reply == disconnect {
			// Remote client is disconnecting.
			n.Close()
			return
		}
		n.writer.Write(msg.Data)
	})
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = sub

	// Handle heartbeats.
	if n.heartbeatInterval > 0 {
		hbSub, err := n.conn.Subscribe(n.heartbeatListen, func(msg *nats.Msg) {
			select {
			case n.recvHeartbeat <- struct{}{}:
			default:
			}
			n.conn.Publish(n.heartbeatReply, nil)
		})
		if err != nil {
			n.Close()
			return thrift.NewTTransportExceptionFromError(err)
		}
		n.heartbeatSub = hbSub
		go func() {
			missed := 0
			for {
				select {
				case <-time.After(n.heartbeatInterval):
					missed++
					if missed >= maxMissedHeartbeats {
						log.Println("frugal: server heartbeat expired")
						n.Close()
						return
					}
				case <-n.recvHeartbeat:
					missed = 0
				case <-n.closed:
					return
				}
			}
		}()
	}
	n.isOpen = true
	return nil
}

func (n *natsServiceTTransport) handshake() error {
	msg, err := n.conn.Request(n.connectSubject, nil, n.connectTimeout)
	if err != nil {
		return err
	}

	if msg.Reply == "" {
		return errors.New("frugal: no reply subject on connect")
	}

	// Connect message consists of "[heartbeat subject] [heartbeat reply subject] [expected interval ms]"
	subjects := strings.Split(string(msg.Data), " ")
	if len(subjects) != 3 {
		return errors.New("frugal: invalid connect message")
	}
	var (
		heartbeatListen = subjects[0]
		heartbeatReply  = subjects[1]
		deadline, err2  = strconv.ParseInt(subjects[2], 10, 64)
	)
	if err2 != nil {
		return err2
	}
	var interval time.Duration
	if deadline > 0 {
		interval = time.Millisecond * time.Duration(deadline)
	}

	n.heartbeatListen = heartbeatListen
	n.heartbeatReply = heartbeatReply
	n.heartbeatInterval = interval
	n.recvHeartbeat = make(chan struct{}, 1)
	n.listenTo = msg.Subject
	n.writeTo = msg.Reply
	return nil
}

func (n *natsServiceTTransport) IsOpen() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.conn.Status() == nats.CONNECTED && n.isOpen
}

// Close unsubscribes, signals the remote peer, and stops heartbeating.
func (n *natsServiceTTransport) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if !n.isOpen {
		return nil
	}

	// Signal remote peer for a graceful disconnect.
	n.conn.PublishRequest(n.writeTo, disconnect, nil)
	if err := n.sub.Unsubscribe(); err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	if n.heartbeatSub != nil {
		if err := n.heartbeatSub.Unsubscribe(); err != nil {
			return thrift.NewTTransportExceptionFromError(err)
		}
	}
	n.sub = nil
	n.heartbeatSub = nil
	close(n.closed)
	n.isOpen = false
	return thrift.NewTTransportExceptionFromError(n.writer.Close())
}

func (n *natsServiceTTransport) Read(p []byte) (int, error) {
	if !n.IsOpen() {
		return 0, thrift.NewTTransportException(thrift.NOT_OPEN, "transport not open")
	}
	num, err := n.reader.Read(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

// Write the bytes to a buffer. If the buffer reaches 1MB, flush the message.
func (n *natsServiceTTransport) Write(p []byte) (int, error) {
	if !n.IsOpen() {
		return 0, thrift.NewTTransportException(thrift.NOT_OPEN, "transport not open")
	}
	remaining := natsMaxMessageSize - n.writeBuffer.Len()
	if remaining < len(p) {
		n.writeBuffer.Write(p[0:remaining])
		if err := n.Flush(); err != nil {
			return 0, thrift.NewTTransportExceptionFromError(err)
		}
		b, err := n.Write(p[remaining:])
		return b, thrift.NewTTransportExceptionFromError(err)
	}
	b, err := n.writeBuffer.Write(p)
	return b, thrift.NewTTransportExceptionFromError(err)
}

// Flush sends the buffered bytes over NATS.
func (n *natsServiceTTransport) Flush() error {
	if !n.IsOpen() {
		return thrift.NewTTransportException(thrift.NOT_OPEN, "transport not open")
	}
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}
	err := n.conn.Publish(n.writeTo, data)
	n.writeBuffer.Reset()
	return thrift.NewTTransportExceptionFromError(err)
}

func (n *natsServiceTTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used.
}
