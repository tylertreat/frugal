package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// A processor is a generic object which operates upon an input stream and
// writes to some output stream.
type FProcessor interface {
	Process(in, out FProtocol) (bool, thrift.TException)
}

type FProcessorFunction interface {
	Process(ctx Context, seqId int32, in, out FProtocol) (bool, thrift.TException)
}

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
