import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:mockito/mockito.dart';

void main() {
  group('FTransport', () {
    int requestSizeLimit = 5;
    MockFRegistry registry;
    FTransportImpl transport;
    FContext context;

    setUp(() {
      registry = new MockFRegistry();
      transport = new FTransportImpl(requestSizeLimit, registry);
      context = new FContext();
    });

    test('test register/unregister call through to registry', () async {
      expect(registry.context, isNull);
      expect(registry.callback, isNull);
      transport.register(context, callback);
      expect(registry.context, equals(context));
      expect(registry.callback, equals(callback));
      transport.unregister(context);
      expect(registry.context, isNull);
      expect(registry.callback, isNull);
    });

    test('test executeFrame calls through to registry execute', () async {
      var response = new Uint8List.fromList([1, 2, 3, 4, 5]);
      var responseFramed = new Uint8List.fromList([0, 0, 0, 5, 1, 2, 3, 4, 5]);
      transport.executeFrame(responseFramed);
      expect(registry.data[0], equals(response));
    });

    test(
        'test closeWithException adds the exeption to the onClose stream and properly triggers the transport monitor',
        () async {
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;
      transport.errors = [null, new FError.withMessage('reopen failed'), null];

      var completer = new Completer<Error>();
      var err = new TTransportError();
      transport.onClose.listen((e) {
        completer.complete(e);
      });

      // Open the transport
      await transport.open();

      // Close the transport with an error
      when(monitor.onClosedUncleanly(any)).thenReturn(1);
      when(monitor.onReopenFailed(any, any)).thenReturn(1);
      await transport.close(err);

      var timeout = new Duration(seconds: 1);
      expect(await completer.future.timeout(timeout), equals(err));
      verify(monitor.onClosedUncleanly(err)).called(1);
      verify(monitor.onReopenFailed(1, 1)).called(1);
      verify(monitor.onReopenSucceeded()).called(1);
    });
  });
}

void callback(TTransport transport) {
  return;
}

class FTransportImpl extends FTransport {
  // Default implementations of non-implemented methods
  List<Error> errors = [];
  int openCalls = 0;

  FTransportImpl(int requestSizeLimit, FRegistry registry)
      : super(requestSizeLimit: requestSizeLimit, registry: registry);

  @override
  Future send(Uint8List payload) => new Future.value();

  @override
  Future open() async {
    if (openCalls <= errors.length) {
      if (errors[openCalls] != null) {
        openCalls++;
        throw errors[openCalls];
      }
    }
    openCalls++;
  }

  bool get isOpen => false;
}

class MockFRegistry extends FRegistry {
  List<Uint8List> data;
  FContext context;
  FAsyncCallback callback;
  Completer executeCompleter;
  Error executeError;

  MockFRegistry() {
    data = new List();
  }

  void initCompleter() {
    executeCompleter = new Completer();
  }

  void register(FContext ctx, FAsyncCallback callback) {
    this.context = ctx;
    this.callback = callback;
  }

  void unregister(FContext ctx) {
    if (this.context == ctx) {
      this.context = null;
      this.callback = null;
    }
  }

  void execute(Uint8List data) {
    this.data.add(data);
    if (executeCompleter != null && !executeCompleter.isCompleted) {
      executeCompleter.complete();
    }
    if (executeError != null) {
      throw executeError;
    }
  }
}

class MockTransportMonitor extends Mock implements FTransportMonitor {
  noSuchMethod(i) => super.noSuchMethod(i);
}
