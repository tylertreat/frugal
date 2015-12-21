part of frugal;

/// Responsible for multiplexing received client messages to the
/// appropriate callback.
class ClientRegistry implements Registry {
  Map<int, AsyncCallback> _handlers;

  ClientRegistry() {
    _handlers = {};
  }

  /// Register a callback for the given Context.
  void register(Context ctx, AsyncCallback callback) {
    var opId = ctx.opId();
    if (_handlers.containsKey(opId)) {
      throw new StateError("frugal: context already registered");
    }
    _handlers[opId] = callback;
  }

  /// Unregister a callback for the given Context.
  void unregister(Context ctx) {
    _handlers.remove(ctx.opId());
  }

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List frame) {
    var headers = decodeHeadersFromFrame(frame);
    var opId;
    try {
      opId = int.parse(headers[_opid]);
    } catch(e) {
      throw new StateError("frugal: frame header does not have opId");
    }

    if (!_handlers.containsKey(opId)) {
      // Unknown handler. Drop the frame.
      return;
    }
    _handlers[opId](new _FTransport(new TUint8List(frame)));

  }
}
