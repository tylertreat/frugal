part of frugal;

/// FAsyncCallback is an internal callback which is constructed by generated
/// code and invoked by an FRegistry when a RPC response is received. In other
/// words, it's used to complete RPCs. The operation ID on FContext is used to
/// look up the appropriate callback. FAsyncCallback is passed an in-memory
/// TTransport which wraps the complete message. The callback returns an error
/// or throws an exception if an unrecoverable error occurs and the transport
/// needs to be shutdown.
typedef void FAsyncCallback(TTransport);

/// FRegistry is responsible for multiplexing and handling messages received
/// from the server. An FRegistry is used by an FTransport.
///
/// When a request is made, an FAsyncCallback is registered to an FContext. When
/// a response for the FContext is received, the FAsyncCallback is looked up,
/// executed, and unregistered.
abstract class FRegistry {
  /// Register a callback for the given Context.
  void register(FContext ctx, FAsyncCallback callback);

  /// Unregister a callback for the given Context.
  void unregister(FContext);

  /// Dispatch a single Frugal message frame.
  void execute(Uint8List);
}
