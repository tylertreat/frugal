part of frugal;

/// Invoked when a message frame is received.
typedef void AsyncCallback(TTransport);

/// Responsible for multiplexing received messages to the
/// appropriate callback.
abstract class Registry {
  /// Register a callback for the given Context.
  void register(Context ctx, AsyncCallback callback);
  /// Unregister a callback for the given Context.
  void unregister(Context);
  /// Dispatch a single Frugal message frame.
  void execute(Uint8List);
}