part of frugal;

/// FAdapterTransport returns an FTransport which uses the given
/// TSocketTransport for send/callback operations in a way that is compatible
/// with Frugal. Messages received on the TSocket (i.e. Frugal frames) are
/// routed to the FRegistry's execute method.
class FAdapterTransport extends FTransport {
  final Logger log = new Logger('FAdapterTransport');
  _TFramedTransport _framedTransport;

  FAdapterTransport(TSocketTransport transport, {FRegistry registry})
      : super(registry: registry),
        _framedTransport = new _TFramedTransport(transport.socket) {
    // If there is an error on the socket, close the transport pessimistically.
    // This error is already logged upstream in TSocketTransport.
    transport.socket.onError.listen((e) => close(e));
    // Forward state changes on to the transport monitor.
    // Note: Just forwarding OPEN on for the time-being.
    transport.socket.onState.listen((state) {
      if (state == TSocketState.OPEN) _monitor?.signalOpen();
    });
  }

  @override
  bool get isOpen => _framedTransport.isOpen;

  @override
  Future open() async {
    await _framedTransport.open();
    _framedTransport.onFrame.listen((_FrameWrapper frame) {
      try {
        _registry.execute(frame.frameBytes);
      } catch (e) {
        // Fatal error. Close the transport.
        log.severe("FAsyncCallback had a fatal error ${e.toString()}." +
            "Closing transport.");
        close(e);
      }
    });
  }

  @override
  Future close([Error error]) async {
    await _framedTransport.close();
    await super.close(error);
  }

  @override
  Future send(Uint8List payload) async {
    if (!isOpen) {
      throw new TTransportError(TTransportErrorType.NOT_OPEN);
    }
    // We need to write to the wrapped TSocket, not the framed transport, since
    // data given to send is already framed.
    _framedTransport.socket.send(payload);
  }
}
