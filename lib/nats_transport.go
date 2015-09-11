package frugal

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/nats-io/nats"
)

type natsTransportFactory struct {
	conn *nats.Conn
}

func NewNATSTransportFactory(conn *nats.Conn) TransportFactory {
	return &natsTransportFactory{conn: conn}
}

func (n *natsTransportFactory) GetTransport(topic string) Transport {
	return NewNATSTransport(n.conn, topic)
}

type natsTransport struct {
	thriftTransport thrift.TTransport
}

func NewNATSTransport(conn *nats.Conn, subject string) Transport {
	return &natsTransport{newNATSThriftTransport(conn, subject)}
}

func (n *natsTransport) Subscribe() error {
	return n.thriftTransport.Open()
}

func (n *natsTransport) Unsubscribe() error {
	return n.thriftTransport.Close()
}

func (n *natsTransport) ThriftTransport() thrift.TTransport {
	return n.thriftTransport
}

func (n *natsTransport) ApplyProxy(transportFactory thrift.TTransportFactory) {
	n.thriftTransport = transportFactory.GetTransport(n.thriftTransport)
}
