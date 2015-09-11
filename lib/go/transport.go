package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// Client contains the Thrift TProtocol and Frugal Transport used by a client
// to publish and receive messages on a particular pub/sub topic.
type Client struct {
	Protocol  thrift.TProtocol
	Transport Transport
}

// TransportFactory is responsible for creating new Frugal Transports.
type TransportFactory interface {
	// GetTransport creates a new Transport for the given pub/sub topic.
	GetTransport(topic string) Transport
}

// Transport wraps a Thrift TTransport which supports pub/sub.
type Transport interface {
	// Subscribe opens the Transport to receive messages on the subscription.
	Subscribe() error

	// Unsubscribe closes the Transport to stop receiving messages on the
	// subscription.
	Unsubscribe() error

	// ThriftTransport returns the wrapped Thrift TTransport.
	ThriftTransport() thrift.TTransport

	// ApplyProxy wraps the underlying TTransport with the TTransport returned
	// by the given TTransportFactory.
	ApplyProxy(thrift.TTransportFactory)
}
