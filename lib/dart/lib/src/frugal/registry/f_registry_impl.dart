part of frugal.src.frugal;

int _globalOPID = 0;

/// Responsible for multiplexing received client messages to the appropriate
/// callback.
class FRegistryImpl implements FRegistry {
  final Logger _log = new Logger('FRegistryImpl');
  Map<int, FAsyncCallback> _handlers;

  /// Create an [FRegistryImpl] instance.
  FRegistryImpl() {
    _handlers = {};
  }

  @override
  void register(FContext ctx, FAsyncCallback callback) {
    // An FContext can be reused for multiple requests. Because of this, every
    // time an FContext is registered, it must be assigned a new op id to
    // ensure we can properly correlate responses. We use a monotonically
    // increasing atomic uint64 for this purpose. If the FContext already has
    // an op id, it has been used for a request. We check the handlers map to
    // ensure that request is not still in-flight.
    if (_handlers.containsKey(ctx._opID)) {
      throw new StateError("frugal: context already registered");
    }
    var opID = _incrementAndGetOPID();
    ctx._opID = opID;
    _handlers[opID] = callback;
  }

  @override
  void unregister(FContext ctx) {
    _handlers.remove(ctx._opID);
  }

  @override
  void execute(Uint8List frame) {
    var headers = Headers.decodeFromFrame(frame);
    var opID;
    try {
      opID = int.parse(headers[_opid]);
    } catch (e) {
      log.warning("frugal: invalid protocol frame: op id not a uint64", e);
      return;
    }

    if (!_handlers.containsKey(opId)) {
      return;
    }
    _handlers[opID](new TMemoryTransport.fromUint8List(frame));
  }

  static int _incrementAndGetOPID() {
    _globalOPID++;
    return _globalOPID;
  }
}
