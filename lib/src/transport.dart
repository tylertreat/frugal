part of frugal;

/// Responsible for creating new Frugal Transports.
abstract class TransportFactory {
  Transport getTransport();
}

/// Wraps a Thrift TTransport which supports pub/sub.
abstract class Transport {
  /// Stream that signals to transport listener that data is available on
  /// the thrift transport.
  Stream get signalRead;

  /// Open the Transport to receive messages on the subscription.
  Future subscribe(String);

  /// Close the Transport to stop receiving messages on the subscription.
  Future unsubscribe();

  /// Prepare the Transport for publishing to the given topic.
  void preparePublish(String);

  /// Return the wrapped Thrift TTransport.
  TTransport thriftTransport();

  /// Wrap the underlying TTransport with the TTransport returned by the
  /// given TTransportFactory.
  Future applyProxy(TTransportFactory);
}