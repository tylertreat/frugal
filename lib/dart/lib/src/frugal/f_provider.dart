part of frugal.frugal;

/// Produces [FPublisherTransport], [FSubscriberTransport], and [FProtocol]
/// instances for use by pub/sub scopes. It does this by wrapping an
/// [FPublisherTransportFactory], an [FSubscriberTransportFactory], and an
/// [FProtocolFactory].
class FScopeProvider {
  final FPublisherTransportFactory publisherTransportFactory;
  final FSubscriberTransportFactory subscriberTransportFactory;
  final FProtocolFactory protocolFactory;

  FScopeProvider(this.publisherTransportFactory,
      this.subscriberTransportFactory, this.protocolFactory);
}

/// The service equivalent of [FScopeProvider]. It produces [FTransport] and
/// [FProtocol] instances for use by RPC service clients.
class FServiceProvider {
  final FTransport transport;
  final FProtocolFactory protocolFactory;

  FServiceProvider(this.transport, this.protocolFactory);
}
