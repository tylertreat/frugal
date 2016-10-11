part of frugal;

/// FScopeProvider produces FPublisherTransports, FSubscriberTransports, and
/// FProtocols for use by pub/sub scopes. It does this by wrapping an
/// FPublisherTransportFactory, an FSubscriberTransportFactory, and an
/// FProtocolFactory.
class FScopeProvider {
  final FPublisherTransportFactory publisherTransportFactory;
  final FSubscriberTransportFactory subscriberTransportFactory;
  final FProtocolFactory protocolFactory;

  FScopeProvider(this.publisherTransportFactory,
      this.subscriberTransportFactory, this.protocolFactory);
}

/// FServiceProvider is the service equivalent of FScopeProvider. It produces
/// FTransports and FProtocols for use by RPC service clients.
class FServiceProvider {
  final FTransport transport;
  final FProtocolFactory protocolFactory;

  FServiceProvider(this.transport, this.protocolFactory);
}
