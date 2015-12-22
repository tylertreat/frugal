package frugal

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

var (
	ErrTooLarge     = errors.New("Message is too large")
	ErrEmptyMessage = errors.New("Message is empty")
)

// NATS limits messages to 1MB.
const natsMaxMessageSize = 1024 * 1024

// tNatsTransport is an implementation of thrift.TTransport exclusively
// used for pub/sub via NATS.
type tNatsTransport struct {
	conn        *nats.Conn
	subject     string
	reader      *bufio.Reader
	writer      *io.PipeWriter
	writeBuffer *bytes.Buffer
	sub         *nats.Subscription
}

func newTNatsTransport(conn *nats.Conn) *tNatsTransport {
	buf := make([]byte, 0, natsMaxMessageSize)
	return &tNatsTransport{
		conn:        conn,
		writeBuffer: bytes.NewBuffer(buf),
	}
}

func (n *tNatsTransport) Open() error {
	reader, writer := io.Pipe()
	n.reader = bufio.NewReader(reader)
	n.writer = writer
	sub, err := n.conn.Subscribe(n.subject, func(msg *nats.Msg) {
		n.writer.Write(msg.Data)
	})
	if err != nil {
		return thrift.NewTTransportExceptionFromError(err)
	}
	n.sub = sub
	return nil
}

func (n *tNatsTransport) IsOpen() bool {
	return n.sub != nil
}

func (n *tNatsTransport) Close() error {
	if !n.IsOpen() {
		return nil
	}
	if err := n.sub.Unsubscribe(); err != nil {
		return err
	}
	n.sub = nil
	return thrift.NewTTransportExceptionFromError(n.writer.Close())
}

func (n *tNatsTransport) Read(p []byte) (int, error) {
	num, err := n.reader.Read(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (n *tNatsTransport) Write(p []byte) (int, error) {
	if len(p)+n.writeBuffer.Len() > natsMaxMessageSize {
		n.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}
	num, err := n.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (n *tNatsTransport) Flush() error {
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return ErrEmptyMessage
	}
	err := n.conn.Publish(n.subject, data)
	n.writeBuffer.Reset()
	return thrift.NewTTransportExceptionFromError(err)
}

func (n *tNatsTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used. See thrift-nats impl.
}

func (n *tNatsTransport) SetSubject(subject string) {
	n.subject = subject
}
