package frugal

import (
	"bufio"
	"io"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

// natsThriftTransport is an implementation of thrift.TTransport exclusively
// used for pub/sub via NATS.
type natsThriftTransport struct {
	conn    *nats.Conn
	subject string
	reader  *bufio.Reader
	writer  *io.PipeWriter
	sub     *nats.Subscription
}

func newNATSThriftTransport(conn *nats.Conn, subject string) thrift.TTransport {
	reader, writer := io.Pipe()
	return &natsThriftTransport{
		conn:    conn,
		subject: subject,
		reader:  bufio.NewReader(reader),
		writer:  writer,
	}
}

func (n *natsThriftTransport) Open() error {
	sub, err := n.conn.Subscribe(n.subject, func(msg *nats.Msg) {
		n.writer.Write(msg.Data)
	})
	if err != nil {
		return err
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
	return nil
}

func (n *natsThriftTransport) Read(p []byte) (int, error) {
	return n.reader.Read(p)
}

func (n *natsThriftTransport) Write(p []byte) (int, error) {
	if err := n.conn.Publish(n.subject, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (n *natsThriftTransport) Flush() error {
	return n.conn.Flush()
}

func (n *natsThriftTransport) RemainingBytes() uint64 {
	return ^uint64(0) // We just don't know unless framed
}
