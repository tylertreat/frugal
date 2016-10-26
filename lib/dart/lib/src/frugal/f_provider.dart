part of frugal.src.frugal;

/// Produces [FPublisherTransport], [FSubscriberTransport], and [FProtocol]
/// instances for use by pub/sub scopes. It does this by wrapping an
/// [FPublisherTransportFactory], an [FSubscriberTransportFactory], and an
/// [FProtocolFactory].
class FScopeProvider {
  /// [FPublisherTransportFactory] used by the scope.
  final FPublisherTransportFactory publisherTransportFactory;

  /// [FSubscriberTransportFactory] used by the scope.
  final FSubscriberTransportFactory subscriberTransportFactory;

  /// [FProtocolFactory] used by the scope.
  final FProtocolFactory protocolFactory;

  /// Creates a new [FScopeProvider].
  FScopeProvider(this.publisherTransportFactory,
      this.subscriberTransportFactory, this.protocolFactory);
}

/// The service equivalent of [FScopeProvider]. It produces [FTransport] and
/// [FProtocol] instances for use by RPC service clients.
class FServiceProvider {
  /// [FTransport] used by the service.
  final FTransport transport;

  /// [FProtocolFactory] used by the service.
  final FProtocolFactory protocolFactory;

  /// Creates a new [FServiceProvider].
  FServiceProvider(this.transport, this.protocolFactory);
}
