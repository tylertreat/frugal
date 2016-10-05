package frugal

// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
// scopes. It does this by wrapping an FScopeTransportFactory and
// FProtocolFactory.
type FScopeProvider struct {
	//scopeTransportFactory FScopeTransportFactory
	publisherTransportFactory FPublisherTransportFactory
	subscriberTransportFactory FSubscriberTransportFactory
	protocolFactory       *FProtocolFactory
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(pub FPublisherTransportFactory, sub FSubscriberTransportFactory, prot *FProtocolFactory) *FScopeProvider {
	return &FScopeProvider{
		publisherTransportFactory: pub,
		subscriberTransportFactory: sub,
		protocolFactory: prot,
	}
}

// NewPublisher returns a new FPublisherTransport and FProtocol used by
// scope publishers.
func (p *FScopeProvider) NewPublisher() (FPublisherTransport, *FProtocol) {
	transport := p.publisherTransportFactory.GetTransport()
	protocol := p.protocolFactory.GetProtocol(transport)
	return transport, protocol
}

// NewSubscriber returns a new FSubscriberTransport and FProtocolFactory used by
// scope subscribers.
func (p *FScopeProvider) NewSubscriber() (FSubscriberTransport, *FProtocolFactory) {
	transport := p.subscriberTransportFactory.GetTransport()
	return transport, p.protocolFactory
}
