import 'dart:async';
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';
import 'package:mockito/mockito.dart';
import 'f_transport_test.dart' show MockFRegistry, MockTransportMonitor;

void main() {
  group('FAdapterTransport', () {
    StreamController<TSocketState> stateStream;
    StreamController<Object> errorStream;
    StreamController<Uint8List> messageStream;
    MockSocket socket;
    MockSocketTransport socketTransport;
    FTransport transport;
    MockFRegistry registry;

    setUp(() {
      stateStream = new StreamController.broadcast();
      errorStream = new StreamController.broadcast();
      messageStream = new StreamController.broadcast();

      socket = new MockSocket();
      when(socket.onState).thenReturn(stateStream.stream);
      when(socket.onError).thenReturn(errorStream.stream);
      when(socket.onMessage).thenReturn(messageStream.stream);
      socketTransport = new MockSocketTransport();
      when(socketTransport.socket).thenReturn(socket);
      transport = new FAdapterTransport(socketTransport);
      registry = new MockFRegistry();
    });

    test('test transport open, set registry open and listen to the socket',
        () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      // Set the registry
      registry.initCompleter();
      transport.setRegistry(registry);
      messageStream.add(new Uint8List.fromList([0, 0, 0, 4, 1, 2, 3, 4]));
      await registry.executeCompleter.future.timeout(new Duration(seconds: 1));
      expect(registry.data[0], equals(new Uint8List.fromList([1, 2, 3, 4])));
    });

    test('test socket error triggers transport close', () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      transport.setRegistry(registry);
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;
      expect(transport.isOpen, equals(true));

      // Kill the socket with an error
      when(socket.isOpen).thenReturn(false);
      var err = new StateError('');
      var closeCompleter = new Completer();
      transport.onClose.listen((e) {
        closeCompleter.complete(e);
      });
      var monitorCompleter = new Completer();
      when(monitor.onClosedUncleanly(any))
          .thenAnswer((Invocation realInvocation) {
        monitorCompleter.complete(realInvocation.positionalArguments[0]);
        return -1;
      });
      errorStream.add(err);
      var timeout = new Duration(seconds: 1);
      expect(await closeCompleter.future.timeout(timeout), equals(err));
      verify(socket.isOpen).called(1);
      expect(await monitorCompleter.future.timeout(timeout), equals(err));
      expect(transport.isOpen, equals(false));

      // Reopen the socket under the hood
      stateStream.add(TSocketState.OPEN);
      monitorCompleter = new Completer();
      when(monitor.onReopenSucceeded()).thenAnswer((Invocation realInvocation) {
        monitorCompleter.complete();
      });
      await monitorCompleter.future.timeout(timeout);
      expect(transport.isOpen, equals(true));
    });
  });
}

class MockSocketTransport extends Mock implements TSocketTransport {
  noSuchMethod(i) => super.noSuchMethod(i);
}

class MockSocket extends Mock implements TSocket {
  noSuchMethod(i) => super.noSuchMethod(i);
}
