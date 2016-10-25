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
func (p *FScopeProvider) NewPublisher() (FPublisherTransport, *FProtocolFactory) {
	transport := p.publisherTransportFactory.GetTransport()
	return transport, p.protocolFactory
}

// NewSubscriber returns a new FSubscriberTransport and FProtocolFactory used by
// scope subscribers.
func (p *FScopeProvider) NewSubscriber() (FSubscriberTransport, *FProtocolFactory) {
	transport := p.subscriberTransportFactory.GetTransport()
	return transport, p.protocolFactory
}
