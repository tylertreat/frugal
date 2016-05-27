part of frugal;

typedef Object InvocationHandler(List<Object> args);
typedef InvocationHandler Middleware(InvocationHandler);

class FMethod {
  InvocationHandler _handler;

  FMethod(f, middleware) {
    this._handler = _compose_middleware(f, middleware);
  }

  Object call(List<Object> args) {
    return this._handler(args);
  }

  InvocationHandler _compose_middleware(f, List<Middleware> middleware) {
    // TODO create the initial handler
    InvocationHandler handler = (args) {
      return Function.apply(f, args);
    };

    if(middleware == null) {
      return handler;
    }
    return middleware.fold(handler, (prev, element) => element(prev));
  }
}
