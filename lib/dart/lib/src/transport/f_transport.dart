part of frugal;

/// FTransport is comparable to Thrift's TTransport in that it represent the
/// transport layer for frugal clients. However, frugal is callback based and
/// sends only framed data. Therefore, instead of exposing read, write, and
/// flush, the transport has a simple send method that sends framed frugal
/// messages. To handle callback data, an FTransport also has an FRegistry,
/// so it provides methods for registering and unregistering an FAsyncCallback
/// to an FContext.
abstract class FTransport {
  static const REQUEST_TOO_LARGE = 100;
  static const RESPONSE_TOO_LARGE = 101;

  MonitorRunner _monitor;
  StreamController _closeController = new StreamController.broadcast();
  Stream<Object> get onClose => _closeController.stream;

  FRegistry _registry;
  int _requestSizeLimit;
  int get requestSizeLimit => _requestSizeLimit;

  FTransport({FRegistry registry, int requestSizeLimit})
      : _registry = registry ?? new FRegistryImpl(),
        _requestSizeLimit = requestSizeLimit ?? 0;

  /// Set an FTransportMonitor on the transport
  void set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, this);
  }

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen;

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open();

  /// Closes the transport.
  Future close([Error error]) => _signalClose(error);

  /// Send the given framed frugal payload over the transport.
  /// Throws [TTransportError] if the payload could not be sent.
  Future send(Uint8List payload);

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback) {
    _registry.register(ctx, callback);
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    _registry.unregister(ctx);
  }

  /// Execute a frugal frame (NOTE: this frame must include the frame size).
  void executeFrame(Uint8List frame) {
    _registry.execute(frame.sublist(4));
  }

  Future _signalClose(cause) async {
    _closeController.add(cause);
    await _monitor?.onClose(cause);
  }
}
