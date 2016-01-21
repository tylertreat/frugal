part of frugal;

/// Wraps a Thrift TTransport. Used for frugal Scopes.
abstract class FScopeTransport extends TTransport {

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    throw new FError.withMessage('Cannot read directly from FScopeTransport');
  }

  /// set the publish topic.
  void setTopic(String topic);

  /// set the subscribe topic and opens the transport.
  Future subscribe(String topic, FAsyncCallback callback);

  /// Signals errors that occur on the transport. An error here
  /// indicates that the transport is now closed.
  Stream<Error> get error;
}