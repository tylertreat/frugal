package frugal

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

// fNatsTransportFactory is an implementation of the FTransportFactory
// interface which creates FTransports backed by NATS.
type fNatsTransportFactory struct {
	conn *nats.Conn
}

// NewFNatsTransportFactory returns a FTransportFactory which creates
// FTransport backed by the NATS messaging system.
func NewFNatsTransportFactory(conn *nats.Conn) FTransportFactory {
	return &fNatsTransportFactory{conn: conn}
}

// GetTransport creates a new NATS FTransport.
func (n *fNatsTransportFactory) GetTransport() FTransport {
	return newFNatsTransport(n.conn)
}

// fNatsTransport is an implementation of the FTransport interface backed by
// the NATS messaging system.
type fNatsTransport struct {
	tTransport thrift.TTransport
	nats       *tNatsTransport
}

// newFNatsTransport creates a new NATS FTransport for the given pub/sub topic.
func newFNatsTransport(conn *nats.Conn) FTransport {
	tr := newTNatsTransport(conn)
	return &fNatsTransport{
		tTransport: tr,
		nats:       tr,
	}
}

// Subscribe opens the FTransport to receive messages on the subscription.
func (n *fNatsTransport) Subscribe(topic string) error {
	n.nats.SetSubject(topic)
	return n.tTransport.Open()
}

// Unsubscribe closes the FTransport to stop receiving messages on the
// subscription.
func (n *fNatsTransport) Unsubscribe() error {
	return n.tTransport.Close()
}

// PreparePublish prepares the FTransport for publishing to the given topic.
func (n *fNatsTransport) PreparePublish(topic string) {
	n.nats.SetSubject(topic)
}

// ThriftTransport returns the wrapped Thrift TTransport.
func (n *fNatsTransport) ThriftTransport() thrift.TTransport {
	return n.tTransport
}

// ApplyProxy wraps the underlying TTransport with the TTransport returned by
// the given TTransportFactory.
func (n *fNatsTransport) ApplyProxy(transportFactory thrift.TTransportFactory) {
	n.tTransport = transportFactory.GetTransport(n.tTransport)
}
