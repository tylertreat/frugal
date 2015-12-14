package frugal

import (
	"log"
	"runtime/debug"

	"git.apache.org/thrift.git/lib/go/thrift"
)

type FSimpleServer struct {
	quit chan struct{}

	processorFactory       FProcessorFactory
	serverTransport        thrift.TServerTransport
	inputTransportFactory  thrift.TTransportFactory
	outputTransportFactory thrift.TTransportFactory
	inputProtocolFactory   FProtocolFactory
	outputProtocolFactory  FProtocolFactory
}

func NewFSimpleServer2(processor FProcessor, serverTransport thrift.TServerTransport) *FSimpleServer {
	return NewFSimpleServerFactory2(NewFProcessorFactory(processor), serverTransport)
}

func NewFSimpleServer4(processor FProcessor, serverTransport thrift.TServerTransport, transportFactory thrift.TTransportFactory, protocolFactory FProtocolFactory) *FSimpleServer {
	return NewFSimpleServerFactory4(NewFProcessorFactory(processor),
		serverTransport,
		transportFactory,
		protocolFactory,
	)
}

func NewFSimpleServer6(processor FProcessor, serverTransport thrift.TServerTransport, inputTransportFactory thrift.TTransportFactory, outputTransportFactory thrift.TTransportFactory, inputProtocolFactory FProtocolFactory, outputProtocolFactory FProtocolFactory) *FSimpleServer {
	return NewFSimpleServerFactory6(NewFProcessorFactory(processor),
		serverTransport,
		inputTransportFactory,
		outputTransportFactory,
		inputProtocolFactory,
		outputProtocolFactory,
	)
}

func NewFSimpleServerFactory2(processorFactory FProcessorFactory, serverTransport thrift.TServerTransport) *FSimpleServer {
	return NewFSimpleServerFactory6(processorFactory,
		serverTransport,
		thrift.NewTTransportFactory(),
		thrift.NewTTransportFactory(),
		NewFBinaryProtocolFactoryDefault(),
		NewFBinaryProtocolFactoryDefault(),
	)
}

func NewFSimpleServerFactory4(processorFactory FProcessorFactory, serverTransport thrift.TServerTransport, transportFactory thrift.TTransportFactory, protocolFactory FProtocolFactory) *FSimpleServer {
	return NewFSimpleServerFactory6(processorFactory,
		serverTransport,
		transportFactory,
		transportFactory,
		protocolFactory,
		protocolFactory,
	)
}

func NewFSimpleServerFactory6(processorFactory FProcessorFactory, serverTransport thrift.TServerTransport, inputTransportFactory thrift.TTransportFactory, outputTransportFactory thrift.TTransportFactory, inputProtocolFactory FProtocolFactory, outputProtocolFactory FProtocolFactory) *FSimpleServer {
	return &FSimpleServer{
		processorFactory:       processorFactory,
		serverTransport:        serverTransport,
		inputTransportFactory:  inputTransportFactory,
		outputTransportFactory: outputTransportFactory,
		inputProtocolFactory:   inputProtocolFactory,
		outputProtocolFactory:  outputProtocolFactory,
		quit: make(chan struct{}, 1),
	}
}

func (p *FSimpleServer) ProcessorFactory() FProcessorFactory {
	return p.processorFactory
}

func (p *FSimpleServer) ServerTransport() thrift.TServerTransport {
	return p.serverTransport
}

func (p *FSimpleServer) InputTransportFactory() thrift.TTransportFactory {
	return p.inputTransportFactory
}

func (p *FSimpleServer) OutputTransportFactory() thrift.TTransportFactory {
	return p.outputTransportFactory
}

func (p *FSimpleServer) InputProtocolFactory() FProtocolFactory {
	return p.inputProtocolFactory
}

func (p *FSimpleServer) OutputProtocolFactory() FProtocolFactory {
	return p.outputProtocolFactory
}

func (p *FSimpleServer) Listen() error {
	return p.serverTransport.Listen()
}

func (p *FSimpleServer) AcceptLoop() error {
	for {
		client, err := p.serverTransport.Accept()
		if err != nil {
			select {
			case <-p.quit:
				return nil
			default:
			}
			return err
		}
		if client != nil {
			go func() {
				if err := p.processRequests(client); err != nil {
					log.Println("error processing request:", err)
				}
			}()
		}
	}
}

func (p *FSimpleServer) Serve() error {
	err := p.Listen()
	if err != nil {
		return err
	}
	p.AcceptLoop()
	return nil
}

func (p *FSimpleServer) Stop() error {
	p.quit <- struct{}{}
	p.serverTransport.Interrupt()
	return nil
}

func (p *FSimpleServer) processRequests(client thrift.TTransport) error {
	processor := p.processorFactory.GetProcessor(client)
	inputTransport := p.inputTransportFactory.GetTransport(client)
	outputTransport := p.outputTransportFactory.GetTransport(client)
	inputProtocol := p.inputProtocolFactory.GetProtocol(inputTransport)
	outputProtocol := p.outputProtocolFactory.GetProtocol(outputTransport)
	defer func() {
		if e := recover(); e != nil {
			log.Printf("panic in processor: %s: %s", e, debug.Stack())
		}
	}()
	if inputTransport != nil {
		defer inputTransport.Close()
	}
	if outputTransport != nil {
		defer outputTransport.Close()
	}
	for {
		ok, err := processor.Process(inputProtocol, outputProtocol)
		if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
			return nil
		} else if err != nil {
			log.Printf("error processing request: %s", err)
			return err
		}
		if !ok {
			break
		}
	}
	return nil
}
