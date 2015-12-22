package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// Provider produces Frugal FTransports and Thrift TProtocols.
type Provider struct {
	FTransportFactory FTransportFactory
	TTransportFactory thrift.TTransportFactory
	TProtocolFactory  thrift.TProtocolFactory
}

// NewProvider creates a new Provider using the given factories.
func NewProvider(t FTransportFactory, f thrift.TTransportFactory, p thrift.TProtocolFactory) *Provider {
	return &Provider{
		FTransportFactory: t,
		TTransportFactory: f,
		TProtocolFactory:  p,
	}
}

// New returns a new FTransport and TProtocol used for pub/sub.
func (p *Provider) New() (FTransport, thrift.TProtocol) {
	transport := p.FTransportFactory.GetTransport()
	if p.TTransportFactory != nil {
		transport.ApplyProxy(p.TTransportFactory)
	}
	protocol := p.TProtocolFactory.GetProtocol(transport.ThriftTransport())
	return transport, protocol
}
