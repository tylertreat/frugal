package frugal

// FScopeProvider produces FScopeTransports and FProtocols for use by pub/sub
// scopes. It does this by wrapping an FScopeTransportFactory and
// FProtocolFactory. This also provides a shim for adding middleware to a
// publisher or subscriber.
type FScopeProvider struct {
	publisherTransportFactory  FPublisherTransportFactory
	subscriberTransportFactory FSubscriberTransportFactory
	protocolFactory            *FProtocolFactory
	middleware                 []ServiceMiddleware
}

// NewFScopeProvider creates a new FScopeProvider using the given factories.
func NewFScopeProvider(pub FPublisherTransportFactory, sub FSubscriberTransportFactory, prot *FProtocolFactory) *FScopeProvider {
	return &FScopeProvider{
		publisherTransportFactory:  pub,
		subscriberTransportFactory: sub,
		protocolFactory:            prot,
		middleware:                 []ServiceMiddleware{},
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

// AddMiddleware adds the given ServiceMiddleware to the FScopeProvider.
func (p *FScopeProvider) AddMiddleware(middleware ServiceMiddleware) {
	p.middleware = append(p.middleware, middleware)
}

// GetMiddleware returns the ServiceMiddleware added to this FScopeProvider.
func (p *FScopeProvider) GetMiddleware() []ServiceMiddleware {
	return p.middleware
}

// FServiceProvider produces FTransports and FProtocolFactories for use by RPC
// service clients. The main purpose of this is to provide a shim for adding
// middleware to a client.
type FServiceProvider struct {
	transport       FTransport
	protocolFactory *FProtocolFactory
	middleware      []ServiceMiddleware
}

// NewFServiceProvider creates a new FServiceProvider containing the given
// FTransport and FProtocolFactory.
func NewFServiceProvider(transport FTransport, protocolFactory *FProtocolFactory) *FServiceProvider {
	return &FServiceProvider{
		transport:       transport,
		protocolFactory: protocolFactory,
		middleware:      []ServiceMiddleware{},
	}
}

// GetTransport returns the contained FTransport.
func (f *FServiceProvider) GetTransport() FTransport {
	return f.transport
}

// GetProtocolFactory returns the contained FProtocolFactory.
func (f *FServiceProvider) GetProtocolFactory() *FProtocolFactory {
	return f.protocolFactory
}

// AddMiddleware adds the given ServiceMiddleware to the FServiceProvider.
func (f *FServiceProvider) AddMiddleware(middleware ServiceMiddleware) {
	f.middleware = append(f.middleware, middleware)
}

// GetMiddleware returns the ServiceMiddleware added to this FServiceProvider.
func (f *FServiceProvider) GetMiddleware() []ServiceMiddleware {
	return f.middleware
}
