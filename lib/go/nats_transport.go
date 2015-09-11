package frugal

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

// natsTransportFactory is an implementation of the TransportFactory interface
// which creates Transports backed by NATS.
type natsTransportFactory struct {
	conn *nats.Conn
}

// NewNATSTransportFactory returns a TransportFactory which creates Transports
// backed by the NATS messaging system.
func NewNATSTransportFactory(conn *nats.Conn) TransportFactory {
	return &natsTransportFactory{conn: conn}
}

// GetTransport creates a new NATS Transport for the given pub/sub topic.
func (n *natsTransportFactory) GetTransport(topic string) Transport {
	return newNATSTransport(n.conn, topic)
}

// natsTransport is an implementation of the Transport interface backed by the
// NATS messaging system.
type natsTransport struct {
	thriftTransport thrift.TTransport
}

// newNATSTransport creates a new NATS Transport for the given pub/sub topic.
func newNATSTransport(conn *nats.Conn, subject string) Transport {
	return &natsTransport{newNATSThriftTransport(conn, subject)}
}

// Subscribe opens the Transport to receive messages on the subscription.
func (n *natsTransport) Subscribe() error {
	return n.thriftTransport.Open()
}

// Unsubscribe closes the Transport to stop receiving messages on the
// subscription.
func (n *natsTransport) Unsubscribe() error {
	return n.thriftTransport.Close()
}

// ThriftTransport returns the wrapped Thrift TTransport.
func (n *natsTransport) ThriftTransport() thrift.TTransport {
	return n.thriftTransport
}

// ApplyProxy wraps the underlying TTransport with the TTransport returned by
// the given TTransportFactory.
func (n *natsTransport) ApplyProxy(transportFactory thrift.TTransportFactory) {
	n.thriftTransport = transportFactory.GetTransport(n.thriftTransport)
}
