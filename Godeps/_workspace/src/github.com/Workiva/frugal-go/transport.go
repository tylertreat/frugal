package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// FTransportFactory is responsible for creating new Frugal FTransports.
type FTransportFactory interface {
	// GetTransport creates a new FTransport.
	GetTransport() FTransport
}

// FTransport wraps a Thrift TTransport which supports pub/sub.
type FTransport interface {
	// Subscribe opens the FTransport to receive messages on the subscription.
	Subscribe(string) error

	// Unsubscribe closes the FTransport to stop receiving messages on the
	// subscription.
	Unsubscribe() error

	// PreparePublish prepares the FTransport for publishing to the given
	// topic.
	PreparePublish(string)

	// ThriftTransport returns the wrapped Thrift TTransport.
	ThriftTransport() thrift.TTransport

	// ApplyProxy wraps the underlying TTransport with the TTransport returned
	// by the given TTransportFactory.
	ApplyProxy(thrift.TTransportFactory)
}
