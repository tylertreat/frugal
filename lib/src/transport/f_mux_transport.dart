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

  @override
  bool get isOpen => _transport.isOpen && _registry != null;

  // TODO: Throw error if direct read

  @override
  void setRegistry(FRegistry registry) {
    if (registry == null) {
      throw new FError.withMessage("registry cannot be null");
    }
    if (_registry != null) {
      return;
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

  @override
  void register(FContext ctx, FAsyncCallback callback) {
    if (_registry == null) {
      throw new FError.withMessage("transport registry not set");
    }
    _registry.register(ctx, callback);
  }

  @override
  void unregister(FContext ctx) {
    if (_registry == null) {
      throw new FError.withMessage("frugal: transport registry not set");
    }
    _registry.unregister(ctx);
  }
}
