package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

type FServer interface {
	ProcessorFactory() FProcessorFactory
	ServerTransport() thrift.TServerTransport
	InputTransportFactory() thrift.TTransportFactory
	OutputTransportFactory() thrift.TTransportFactory
	InputProtocolFactory() FProtocolFactory
	OutputProtocolFactory() FProtocolFactory

	// Starts the server
	Serve() error
	// Stops the server. This is optional on a per-implementation basis. Not
	// all servers are required to be cleanly stoppable.
	Stop() error
}
