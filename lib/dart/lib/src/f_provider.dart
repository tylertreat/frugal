part of frugal;

/// FScopeProvider produces FPublisherTransports, FSubscriberTransports, and
/// FProtocols for use by pub/sub scopes. It does this by wrapping an
/// FPublisherTransportFactory, an FSubscriberTransportFactory, and an
/// FProtocolFactory.
class FScopeProvider {
  final FPublisherTransportFactory fPublisherTransportFactory;
  final FSubscriberTransportFactory fSubscriberTransportFactory;
  final FProtocolFactory fProtocolFactory;

  FScopeProvider(this.fPublisherTransportFactory,
      this.fSubscriberTransportFactory, this.fProtocolFactory);
}

/// FServiceProvider is the service equivalent of FScopeProvider. It produces
/// FTransports and FProtocols for use by RPC service clients.
class FServiceProvider {
  final FTransport fTransport;
  final FProtocolFactory fProtocolFactory;

  FServiceProvider(this.fTransport, this.fProtocolFactory);
}
