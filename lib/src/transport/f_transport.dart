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

/// FMultiplexedTransport is a multiplexed Transport that routes frames to the
/// correct callbacks.
class FMultiplexedTransport extends FTransport {
  final Logger log = new Logger('FTransport');
  _TFramedTransport _transport;
  FRegistry _registry;

  FMultiplexedTransport(TSocketTransport transport)
    : _transport = new _TFramedTransport(transport.socket) {
    super.transport = _transport;
  }

  /// Set the Registry on the transport and starts listening for frames.
  void setRegistry(FRegistry registry) {
    _registry = registry;
    _transport.onFrame.listen((Uint8List frame) {
      try {
        _registry.execute(frame);
      } catch(e) {
        // Fatal error. Close the transport.
        log.severe("FAsyncCallback had a fatal error ${e.toString()}." +
          "Closing transport.");
        close();
      }
    });
  }

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new StateError("frugal: transport registry not set");
    }
    _registry.register(ctx, callback);
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new StateError("frugal: transport registry not set");
    }
    _registry.unregister(ctx);
  }
}
