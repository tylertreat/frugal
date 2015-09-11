package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

type Client struct {
	Protocol  thrift.TProtocol
	Transport Transport
}

type ClientProvider map[string]*Client

type TransportFactory interface {
	GetTransport(topic string) Transport
}

type Transport interface {
	Subscribe() error
	Unsubscribe() error
	ThriftTransport() thrift.TTransport
	ApplyProxy(thrift.TTransportFactory)
}
