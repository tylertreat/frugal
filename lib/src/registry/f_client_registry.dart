part of frugal;

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
    var opId = ctx.opId();
    if (_handlers.containsKey(opId)) {
      throw new StateError("frugal: context already registered");
    }
    _handlers[opId] = callback;
  }

  /// Unregister a callback for the given Context.
  void unregister(FContext ctx) {
    _handlers.remove(ctx.opId());
  }

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List frame) {
    var headers = Headers.decodeFromFrame(frame);
    var opId;
    try {
      opId = int.parse(headers[_opid]);
    } catch(e) {
      log.warning("Frame headers does not have an opId");
      return;
    }

    if (!_handlers.containsKey(opId)) {
      log.info("No handler for op $opId}. Dropping frame.");
      return;
    }
    _handlers[opId](new TMemoryTransport.fromUnt8List(frame));
  }
}
