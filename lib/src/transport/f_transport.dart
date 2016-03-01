part of frugal;

/// FTransport is a TTransport for services.
abstract class FTransport extends TTransport {
  static const REQUEST_TOO_LARGE = 100;
  static const RESPONSE_TOO_LARGE = 101;

  TTransport _transport;
  MonitorRunner _monitor;

  void set transport(TTransport transport) {
    _transport = transport;
  }

  void set monitor(FTransportMonitor monitor) {
    _monitor = new MonitorRunner(monitor, _transport);
  }

  @override
  bool get isOpen => _transport.isOpen;

  @override
  Future open() => _transport.open();

  @override
  int read(Uint8List buffer, int offset, int length) {
    return _transport.read(buffer, offset, length);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _transport.write(buffer, offset, length);
  }

  @override
  Future flush() => _transport.flush();

  /// Set the Registry on the transport.
  void setRegistry(FRegistry registry);

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx);

  Future signalClose(cause) async {
    if (_monitor != null) await _monitor.onClose(cause);
  }
}
