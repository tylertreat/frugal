import "package:test/test.dart";
import "package:thrift/thrift.dart";

import "../../lib/src/frugal.dart";

void main() {
  group('FRegistryImpl', () {
    test(
        'test the registry sucessfully executes an FAsyncCallback registered to an FContext',
        () {
      var ctx = new FContext(correlationID: 'sweet-corid');
      var mockCallback = new MockCallback();
      FRegistry registry = new FRegistryImpl();
      registry.register(ctx, mockCallback.callback);

      var protocolFactory = new FProtocolFactory(new TBinaryProtocolFactory());
      var transport = new TMemoryTransport();
      var oprot = protocolFactory.getProtocol(transport);
      ctx.addResponseHeader('_opid', ctx.requestHeader('_opid'));
      ctx.addResponseHeader('foo', 'bar');
      oprot.writeResponseHeader(ctx);
      var msg = new TMessage("fooMethod", TMessageType.REPLY, 3);
      oprot.writeMessageBegin(msg);
      oprot.writeMessageEnd();

      registry.execute(transport.buffer);

      expect(mockCallback.calls, equals(1));
      var iprot = protocolFactory.getProtocol(mockCallback.transport);
      var iCtx = new FContext();
      iprot.readResponseHeader(iCtx);
      expect(iCtx.responseHeader('foo'), equals(ctx.responseHeader('foo')));
      var iMsg = iprot.readMessageBegin();
      expect(iMsg.name, equals(msg.name));
      iprot.readMessageEnd();

      // Send a frame with no opid, make sure no handler is called
      var badTransport = new TMemoryTransport();
      var badOprot = protocolFactory.getProtocol(badTransport);
      var badCtx = new FContext();
      badOprot.writeResponseHeader(badCtx);
      registry.execute(badTransport.buffer);
      expect(mockCallback.calls, equals(1));

      // Unregister the context and make sure any more responses for the
      // context are dropped
      registry.unregister(ctx);
      registry.execute(transport.buffer);
      expect(mockCallback.calls, equals(1));
    });

    test(
        'test register throws an exception if the contest is already bound to a callback',
        () {
      var ctx = new FContext(correlationID: 'sweet-corid');
      var mockCallback = new MockCallback();
      FRegistry registry = new FRegistryImpl();
      registry.register(ctx, mockCallback.callback);

      expect(() => registry.register(ctx, mockCallback.callback),
          throwsA(new isInstanceOf<StateError>()));
    });
  });
}

/// Mock callback.
class MockCallback {
  /// Number of calls.
  int calls = 0;

  /// Transport used by the callback.
  TTransport transport;

  /// Mock callback function.
  void callback(TTransport transport) {
    calls++;
    this.transport = transport;
  }
}
