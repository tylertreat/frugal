package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// FProcessor is a generic object which operates upon an input stream and
// writes to some output stream.
type FProcessor interface {
	Process(in, out *FProtocol) error
}

// FProcessorFunction is a processor for a specific API call.
type FProcessorFunction interface {
	Process(ctx *FContext, in, out *FProtocol) error
}

// FProcessorFactory produces FProcessors used by FServer.
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
