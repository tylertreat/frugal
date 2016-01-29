package frugal

import "log"

// FSimpleServer is a simple, single-threaded FServer.
type FSimpleServer struct {
	quit             chan struct{}
	processorFactory FProcessorFactory
	serverTransport  FServerTransport
	transportFactory FTransportFactory
	protocolFactory  *FProtocolFactory
}

func NewFSimpleServerFactory5(
	processorFactory FProcessorFactory,
	serverTransport FServerTransport,
	transportFactory FTransportFactory,
	protocolFactory *FProtocolFactory) *FSimpleServer {

	return &FSimpleServer{
		processorFactory: processorFactory,
		serverTransport:  serverTransport,
		transportFactory: transportFactory,
		protocolFactory:  protocolFactory,
		quit:             make(chan struct{}, 1),
	}
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

// Serve starts the server.
func (p *FSimpleServer) Serve() error {
	err := p.Listen()
	if err != nil {
		return err
	}
	p.AcceptLoop()
	return nil
}

// Stop the server.
func (p *FSimpleServer) Stop() error {
	close(p.quit)
	p.serverTransport.Interrupt()
	return nil
}

func (p *FSimpleServer) processRequests(client FTransport) error {
	processor := p.processorFactory.GetProcessor(client)
	transport := p.transportFactory.GetTransport(client)
	protocol := p.protocolFactory.GetProtocol(transport)
	transport.SetRegistry(NewServerRegistry(processor, p.protocolFactory, protocol))

	select {
	case <-p.quit:
		transport.Close()
	case <-client.Closed():
	}

	return nil
}
