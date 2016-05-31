import "package:frugal/frugal.dart";
import "package:test/test.dart";

class MiddlewareTestingService {
  String str;
  int handleSomething(int first, int second, String str) {
    this.str = str;
    return first + second;
  }
}

class MiddlewareDataStruct {
  int arg;
  String methodName;
  String serviceName;
}

void main() {
  Middleware newMiddleware(MiddlewareDataStruct mds) {
    return (InvocationHandler next) {
      return (String serviceName, String methodName, List<Object> args) {
        mds.arg = args[0];
        mds.serviceName = serviceName;
        mds.methodName = methodName;
        args[0] = mds.arg + 1;
        return next(serviceName, methodName, args);
      };
    };
  }

  test('no middleware', () {
    MiddlewareTestingService service = new MiddlewareTestingService();
    FMethod method = new FMethod(service.handleSomething, 'MiddlewareTestingService', 'handleSomething', null);
    expect(method([3, 64, 'foo']), equals(67));
    expect(service.str, equals('foo'));
  });

  test('middleware', () {
    MiddlewareDataStruct mds1 = new MiddlewareDataStruct();
    MiddlewareDataStruct mds2 = new MiddlewareDataStruct();

    MiddlewareTestingService service = new MiddlewareTestingService();
    Middleware middleware1 = newMiddleware(mds1);
    Middleware middleware2 = newMiddleware(mds2);
    FMethod method = new FMethod(service.handleSomething, 'MiddlewareTestingService', 'handleSomething', [middleware1, middleware2]);
    expect(method([3, 64, 'foo']), equals(69));
    expect(mds1.arg, equals(4));
    expect(mds2.arg, equals(3));
    expect(mds1.serviceName, equals('MiddlewareTestingService'));
    expect(mds2.serviceName, equals('MiddlewareTestingService'));
    expect(mds1.methodName, equals('handleSomething'));
    expect(mds2.methodName, equals('handleSomething'));
  });
}
