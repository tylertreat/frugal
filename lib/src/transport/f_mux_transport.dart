part of frugal;

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

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen => _transport.isOpen && _registry != null;

  // TODO: Throw error if direct read

  /// Set the Registry on the transport and starts listening for frames.
  void setRegistry(FRegistry registry) {
    if (registry == null) {
      throw new FError.withMessage("registry cannot be null");
    }
    if (_registry != null) {
      throw new FError.withMessage("registry alread set");
    }

    _registry = registry;
    _transport.onFrame.listen((Uint8List frame) {
      try {
        _registry.execute(frame);
      } catch(e) {
        // TODO: Log the stacktrace
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
      throw new FError.withMessage("transport registry not set");
    }
    _registry.register(ctx, callback);
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage("frugal: transport registry not set");
    }
    _registry.unregister(ctx);
  }
}
