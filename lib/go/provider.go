package frugal

// FScopeProvider produces Frugal FScopeTransports and FProtocols used by
// pub/sub scopes.
type FScopeProvider struct {
	scopeTransportFactory FScopeTransportFactory
	protocolFactory       *FProtocolFactory
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(t FScopeTransportFactory, p *FProtocolFactory) *FScopeProvider {
	return &FScopeProvider{
		scopeTransportFactory: t,
		protocolFactory:       p,
	}
}

// New returns a new FScopeTransport and TProtocol used by pub/sub scopes.
func (p *FScopeProvider) New() (FScopeTransport, *FProtocol) {
	transport := p.scopeTransportFactory.GetTransport()
	protocol := p.protocolFactory.GetProtocol(transport)
	return transport, protocol
}

// FServiceProvider produces a FTransport and FProtocolFactory used by
// service clients.
type FServiceProvider struct {
	transport       FTransport
	protocolFactory *FProtocolFactory
}

// NewFServiceProvider creates a new FServiceProvider used by service clients.
func NewFServiceProvider(t FTransport, p *FProtocolFactory) *FServiceProvider {
	return &FServiceProvider{
		transport:       t,
		protocolFactory: p,
	}
}

// Transport returns the FTransport.
func (f *FServiceProvider) Transport() FTransport {
	return f.transport
}

// ProtocolFactory returns the FProtocolFactory.
func (f *FServiceProvider) ProtocolFactory() *FProtocolFactory {
	return f.protocolFactory
}
