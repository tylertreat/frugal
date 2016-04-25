package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic
// object which operates upon an input stream and writes to an output stream.
// Specifically, an FProcessor is provided to an FServer in order to wire up a
// service handler to process requests.
type FProcessor interface {
	Process(in, out *FProtocol) error
}

// FProcessorFunction is used internally by generated code. An FProcessor
// registers an FProcessorFunction for each service method. Like FProcessor, an
// FProcessorFunction exposes a single process call, which is used to handle a
// method invocation.
type FProcessorFunction interface {
	Process(ctx *FContext, in, out *FProtocol) error
}

// FProcessorFactory produces FProcessors and is used by an FServer. It takes a
// TTransport and returns an FProcessor wrapping it.
type FProcessorFactory interface {
	GetProcessor(trans thrift.TTransport) FProcessor
}

type fProcessorFactory struct {
	processor FProcessor
}

func NewFProcessorFactory(p FProcessor) FProcessorFactory {
	return &fProcessorFactory{processor: p}
}

func (p *fProcessorFactory) GetProcessor(trans thrift.TTransport) FProcessor {
	return p.processor
}
