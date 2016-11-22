part of frugal;

int _globalOpId = 0;

/// Responsible for multiplexing received client messages to the
/// appropriate callback.
class FClientRegistry implements FRegistry {
  final Logger log = new Logger('ClientRegistry');
  Map<int, FAsyncCallback> _handlers;

  FClientRegistry() {
    _handlers = {};
  }

  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback) {
    // An FContext can be reused for multiple requests. Because of this, every
    // time an FContext is registered, it must be assigned a new op id to
    // ensure we can properly correlate responses. We use a monotonically
    // increasing atomic uint64 for this purpose. If the FContext already has
    // an op id, it has been used for a request. We check the handlers map to
    // ensure that request is not still in-flight.
    if (_handlers.containsKey(ctx._opId())) {
      throw new StateError("frugal: context already registered");
    }
    var opId = _incrementAndGetOpId();
    ctx._setOpId(opId);
    _handlers[opId] = callback;
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    _handlers.remove(ctx._opId());
  }

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List frame) {
    var headers = Headers.decodeFromFrame(frame);
    var opId;
    try {
      opId = int.parse(headers[_opidHeader]);
    } catch (e) {
      log.warning("frugal: invalid protocol frame: op id not a uint64", e);
      return;
    }

    if (!_handlers.containsKey(opId)) {
      return;
    }
    _handlers[opId](new TMemoryTransport.fromUint8List(frame));
  }

  static int _incrementAndGetOpId() {
    _globalOpId++;
    return _globalOpId;
  }
}
