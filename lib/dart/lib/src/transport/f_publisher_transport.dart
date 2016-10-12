part of frugal;

/// FPublisherTransport extends Thrift's TTransport and is used exclusively
/// for scope publishers.
abstract class FPublisherTransport extends TTransport {
  @override
  int read(Uint8List buffer, int offset, int length) {
    throw new TTransportError(
        TTransportErrorType.UNKNOWN, 'Cannot read from FPublisherTransport');
  }

  /// set the publish topic.
  void setTopic(String topic);
}

/// FPublisherTransportFactory produces FPublisherTransports.
abstract class FPublisherTransportFactory {
  FPublisherTransport getTransport();
}
