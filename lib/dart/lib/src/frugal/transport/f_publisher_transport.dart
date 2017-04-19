part of frugal.src.frugal;

/// Transport layer for scope publishers.
abstract class FPublisherTransport {
  /// Query whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen;

  /// Open the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open();

  /// Close the transport.
  Future close();

  /// The maximum publish size permitted by the transport. If [publishSizeLimit]
  /// is a non-positive number, the transport is assumed to have no publish size
  /// limit.
  int get publishSizeLimit;

  /// Publish the given framed frugal payload over the transport.
  /// Throws [TTransportError] if publishing the payload failed.
  void publish(String topic, Uint8List payload);
}

/// Produces [FPublisherTransport] instances.
abstract class FPublisherTransportFactory {
  /// Return a new [FPublisherTransport] instance.
  FPublisherTransport getTransport();
}

/// Transport layer for durable scope providers.
abstract class FDurablePublisherTransport {
  /// Query whether the transport is open.
  bool get isOpen;

  /// Open the transport for publishing.
  Future<Null> open();

  /// Close the transport.
  Future<Null> close();

  /// The maximum publish size permitted by the transport. If this is a
  /// non-positive number, the transport is assumed to have no publish size
  /// limit.
  int get publishSizeLimit;

  /// Publish the given framed frugal payload over the transport.
  /// Throws [TTransportError] if publishing the payload failed.
  /// The parameter groupID allows you to assign arbitrary groupings to
  /// messages. A value of null indicates the message isn't part of any group.
  Future<Null> publish(String topic, String groupId, Uint8List payload);
}

/// Produces [FDurablePublisherTransport] instances.
abstract class FDurablePublisherTransportFactory {
  /// Returns a new [FDurablePublisherTransport] instance.
  FDurablePublisherTransport getTransport();
}
