part of frugal;

/// InvocationHandler processes a method invocation on a proxied method and
/// returns the result. The arguments should match the arity of the proxied
/// method, and have the same types. The first argument will always be the
/// FContext.
typedef Object InvocationHandler(
    String serviceName, String methodName, List<Object> args);

/// Middleware is used to implement interceptor logic around API calls. This can
/// be used, for example, to implement retry policies on service calls, logging,
/// telemetry, or authentication and authorization. Middleware me be applied to
/// both RPC services and pub/sub scopes. Middleware returns an
/// InvocationHandler which proxies the given InvocationHandler.
typedef InvocationHandler Middleware(InvocationHandler);

/// FMethod contains an InvocationHandler used to proxy the given service method
class FMethod {
  String _serviceName;
  String _methodName;
  InvocationHandler _handler;

  FMethod(dynamic f, String serviceName, String methodName,
      List<Middleware> middleware) {
    this._serviceName = serviceName;
    this._methodName = methodName;
    this._handler = _composeMiddleware(f, middleware);
  }

  /// Call invokes the proxied InvocationHandler with the given arguments
  /// and returns the results.
  Object call(List<Object> args) {
    return this._handler(this._serviceName, this._methodName, args);
  }

  /// ComposeMiddleware applies the Middleware to the provided method.
  InvocationHandler _composeMiddleware(f, List<Middleware> middleware) {
    InvocationHandler handler = (serviceName, methodName, args) {
      return Function.apply(f, args);
    };

    if (middleware == null) {
      return handler;
    }
    return middleware.fold(handler, (prev, element) => element(prev));
  }
}
