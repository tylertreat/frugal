part of frugal;

/// FAdapterTransport is an FTransport that executes TSocketTransport
/// frames.
class FAdapterTransport extends FTransport {
  final Logger log = new Logger('FAdapterTransport');
  _TFramedTransport _transport;

  FAdapterTransport(TSocketTransport transport)
      : _transport = new _TFramedTransport(transport.socket) {
    // If there is an error on the socket, close the transport pessimistically.
    // This error is already logged upstream in TSocketTransport.
    transport.socket.onError.listen((e) => closeWithException(e));
    // Forward state changes on to the transport monitor.
    // Note: Just forwarding OPEN on for the time-being.
    transport.socket.onState.listen((state) {
      if (state == TSocketState.OPEN) _monitor?.signalOpen();
    });
  }

  @override
  bool get isOpen => _transport.isOpen && _registry != null;

  @override
  Future open() => _transport.open();

  @override
  Future close() => closeWithException(null);

  // TODO: Remove this override with 2.0
  @override
  int read(Uint8List buffer, int offset, int length) {
    throw new UnsupportedError("Cannot call read on FTransport");
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _transport.write(buffer, offset, length);
  }

  Future closeWithException(cause) async {
    await _transport.close();
    await _signalClose(cause);
  }

  @override
  void setRegistry(FRegistry registry) {
    super.setRegistry(registry);
    _registry = registry;
    _transport.onFrame.listen((_FrameWrapper frame) {
      try {
        var dur = new DateTime.now().difference(frame.timestamp);
        if (dur > _highWatermark) {
          log.warning(
              "frame spent ${dur} in the transport buffer, your consumer might be backed up");
        }
        _registry.execute(frame.frameBytes);
      } catch (e) {
        // TODO: Log the stacktrace
        // Fatal error. Close the transport.
        log.severe("FAsyncCallback had a fatal error ${e.toString()}." +
            "Closing transport.");
        closeWithException(e);
      }
    });
  }
}
