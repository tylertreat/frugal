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

// natsThriftTransport is an implementation of thrift.TTransport exclusively
// used for pub/sub via NATS.
type natsThriftTransport struct {
	conn        *nats.Conn
	subject     string
	reader      *bufio.Reader
	writer      *io.PipeWriter
	writeBuffer *bytes.Buffer
	sub         *nats.Subscription
}

func newNATSThriftTransport(conn *nats.Conn) *natsThriftTransport {
	buf := make([]byte, 0, natsMaxMessageSize)
	return &natsThriftTransport{
		conn:        conn,
		writeBuffer: bytes.NewBuffer(buf),
	}
}

func (n *natsThriftTransport) Open() error {
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

func (n *natsThriftTransport) IsOpen() bool {
	return n.sub != nil
}

func (n *natsThriftTransport) Close() error {
	if !n.IsOpen() {
		return nil
	}
	if err := n.sub.Unsubscribe(); err != nil {
		return err
	}
	n.sub = nil
	return thrift.NewTTransportExceptionFromError(n.writer.Close())
}

func (n *natsThriftTransport) Read(p []byte) (int, error) {
	num, err := n.reader.Read(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (n *natsThriftTransport) Write(p []byte) (int, error) {
	if len(p)+n.writeBuffer.Len() > natsMaxMessageSize {
		n.writeBuffer.Reset() // Clear any existing bytes.
		return 0, ErrTooLarge
	}
	num, err := n.writeBuffer.Write(p)
	return num, thrift.NewTTransportExceptionFromError(err)
}

func (n *natsThriftTransport) Flush() error {
	data := n.writeBuffer.Bytes()
	if len(data) == 0 {
		return ErrEmptyMessage
	}
	err := n.conn.Publish(n.subject, data)
	n.writeBuffer.Reset()
	return thrift.NewTTransportExceptionFromError(err)
}

func (n *natsThriftTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We don't know unless framed is used. See thrift-nats impl.
}

func (n *natsThriftTransport) SetSubject(subject string) {
	n.subject = subject
}
