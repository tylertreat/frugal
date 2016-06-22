part of frugal;

/// InvocationHandler processes a method invocation on a proxied method and
/// returns the result. The arguments should match the arity of the proxied
/// method, and have the same types. The first argument will always be the
/// FContext.
typedef Future InvocationHandler(
    String serviceName, String methodName, List<Object> args);

/// Middleware is used to implement interceptor logic around API calls. This can
/// be used, for example, to implement retry policies on service calls, logging,
/// telemetry, or authentication and authorization. Middleware me be applied to
/// both RPC services and pub/sub scopes. Middleware returns an
/// InvocationHandler which proxies the given InvocationHandler.
typedef InvocationHandler Middleware(InvocationHandler);

/// FMethod contains an InvocationHandler used to proxy the given service method
/// This should only be used by generated code.
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
  Future call(List<Object> args) {
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

/// Middleware for debugging that logs the requests and responses when activated with the url param frugal_debug=true
InvocationHandler debugMiddleware(InvocationHandler next) {
  return (String serviceName, String methodName, List<Object> args) async {
    print('frugal called $serviceName.$methodName');
    for (int i = 0; i < args.length; i++) {
      int iHuman = i + 1;
      var obj = args[i];
      String type = obj.runtimeType.toString();
      String json = fObjToJson(obj);
      print(
          'frugal called $serviceName.$methodName: arg #$iHuman: $type: $json');
    }
    Object ret = await next(serviceName, methodName, args);
    String type = ret.runtimeType.toString();
    String json = fObjToJson(ret);
    print('frugal $serviceName.$methodName returned: $type: $json');
    return ret;
  };
}
