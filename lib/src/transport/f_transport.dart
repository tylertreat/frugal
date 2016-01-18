part of frugal;

/// FTransport is a TTransport for services.
abstract class FTransport extends TTransport {
  TTransport _transport;

  void set transport(TTransport transport) {
    _transport = transport;
  }

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen => _transport.isOpen;

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open() => _transport.open();

  /// Closes the transport.
  Future close() => _transport.close();

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    return _transport.read(buffer, offset, length);
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write(Uint8List buffer, int offset, int length) {
    _transport.write(buffer, offset, length);
  }

  /// Flush any pending data out of a transport buffer.
  /// Throws [TTransportError] if there was an error writing out data.
  Future flush() => _transport.flush();

  /// Set the Registry on the transport.
  void setRegistry(FRegistry registry);

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx);
}
