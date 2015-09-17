package frugal

import "git.apache.org/thrift.git/lib/go/thrift"

// Provider produces Frugal Transports and Thrift TProtocols.
type Provider struct {
	TransportFactory       TransportFactory
	ThriftTransportFactory thrift.TTransportFactory
	ProtocolFactory        thrift.TProtocolFactory
}

// NewProvider creates a new Provider using the given factories.
func NewProvider(t TransportFactory, f thrift.TTransportFactory, p thrift.TProtocolFactory) *Provider {
	return &Provider{
		TransportFactory:       t,
		ThriftTransportFactory: f,
		ProtocolFactory:        p,
	}
}

// New returns a new Transport and TProtocol used for pub/sub.
func (p *Provider) New() (Transport, thrift.TProtocol) {
	transport := p.TransportFactory.GetTransport("")
	if p.ThriftTransportFactory != nil {
		transport.ApplyProxy(p.ThriftTransportFactory)
	}
	protocol := p.ProtocolFactory.GetProtocol(transport.ThriftTransport())
	return transport, protocol
}
