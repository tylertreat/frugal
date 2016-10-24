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
      registry = new MockFRegistry();
      transport = new FAdapterTransport(socketTransport, registry: registry);
    });

    test('test opening transport opens and listens to the socket', () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      // Initialize the registry
      registry.initCompleter();

      // Add a message to the socket
      messageStream.add(new Uint8List.fromList([0, 0, 0, 4, 1, 2, 3, 4]));
      await registry.executeCompleter.future.timeout(new Duration(seconds: 1));
      expect(registry.data[0], equals(new Uint8List.fromList([1, 2, 3, 4])));
    });

    test('test transport sends properly', () async {
      // flush transport before opening
      var framedBuffer = new Uint8List.fromList([0, 0, 0, 4, 1, 2, 3, 4]);
      expect(transport.send(framedBuffer),
          throwsA(new isInstanceOf<TTransportError>()));

      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      verify(socket.open()).called(1);

      // Initialize the registry
      registry.initCompleter();

      // Write to/flush transport
      await transport.send(framedBuffer);
      verify(socket.send(framedBuffer)).called(1);
    });

    test('test socket error triggers transport close', () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      var monitor = new MockTransportMonitor();
      transport.monitor = monitor;
      expect(transport.isOpen, equals(true));

      // Kill the socket with an error
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

    test('test registry error triggers transport close', () async {
      // Open the transport
      when(socket.isClosed).thenReturn(true);
      when(socket.open()).thenReturn(new Future.value());
      await transport.open();
      expect(transport.isOpen, equals(true));

      // Kill the transport with the registry failing
      registry.initCompleter();
      var err = new FError();
      registry.executeError = err;
      var closeCompleter = new Completer();
      transport.onClose.listen((e) {
        closeCompleter.complete(e);
      });

      // Make sure socket gets closed
      when(socket.open()).thenReturn(new Future.value());
      messageStream.add(new Uint8List.fromList([0, 0, 0, 4, 1, 2, 3, 4]));
      var timeout = new Duration(seconds: 1);
      await registry.executeCompleter.future.timeout(timeout);
      expect(registry.data[0], equals(new Uint8List.fromList([1, 2, 3, 4])));

      // Make sure the transport was closed with the correct error
      expect(await closeCompleter.future.timeout(timeout), equals(err));
      expect(transport.isOpen, equals(false));
    });
  });
}

/// Mock socket transport.
class MockSocketTransport extends Mock implements TSocketTransport {}

/// Mock socket.
class MockSocket extends Mock implements TSocket {}
