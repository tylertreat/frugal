part of frugal.src.frugal;

/// Comparable to Thrift's [TTransport] in that it represents the transport
/// layer for frugal clients. However, frugal is callback based and sends only
/// framed data. Therefore, instead of exposing read, write, and flush, the
/// transport has a simple [send] method that sends framed frugal messages. To
/// handle callback data, also has an [FRegistry], so it provides methods for
/// registering and unregistering an [FAsyncCallback] to an [FContext].
abstract class FTransport {

  MonitorRunner _monitor;
  StreamController _closeController = new StreamController.broadcast();
  FRegistry _registry;

  /// Limits the size of requests to the server.
  /// No limit will be enforced if set to a non-positive value (i.e. <1).
  final int requestSizeLimit;

  /// Create an [FTransport] with the optional [FRegistry] and
  /// [requestSizeLimit].
  FTransport({FRegistry registry, this.requestSizeLimit})
      : _registry = registry ?? new FRegistryImpl();

  /// Listen to close events on the transport.
  Stream<Object> get onClose => _closeController.stream;

  /// Set an [FTransportMonitor] on the transport.
  set monitor(FTransportMonitor monitor) {
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

  /// Register an [FAsyncCallback] to the given [FContext].
  void register(FContext ctx, FAsyncCallback callback) {
    _registry.register(ctx, callback);
  }

  /// Unregister any associated [FAsyncCallback] from the given [FContext].
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
