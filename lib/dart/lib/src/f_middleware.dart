part of frugal;

typedef Object InvocationHandler(String serviceName, String methodName, List<Object> args);
typedef InvocationHandler Middleware(InvocationHandler);

class FMethod {
  String _serviceName;
  String _methodName;
  InvocationHandler _handler;

  FMethod(dynamic f, String serviceName, String methodName, List<Middleware> middleware) {
    this._serviceName = serviceName;
    this._methodName = methodName;
    this._handler = _composeMiddleware(f, middleware);
  }

  Object call(List<Object> args) {
    return this._handler(this._serviceName, this._methodName, args);
  }

  InvocationHandler _composeMiddleware(f, List<Middleware> middleware) {
    InvocationHandler handler = (serviceName, methodName, args) {
      return Function.apply(f, args);
    };

    if(middleware == null) {
      return handler;
    }
    return middleware.fold(handler, (prev, element) => element(prev));
  }
}
