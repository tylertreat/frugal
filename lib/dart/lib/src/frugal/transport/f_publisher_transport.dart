part of frugal.src.frugal;

/// Transport layer for scope publishers.
abstract class FPublisherTransport {
  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen;

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open();

  /// Closes the transport.
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
  FPublisherTransport getTransport();
}
