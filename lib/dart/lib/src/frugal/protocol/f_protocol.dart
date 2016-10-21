part of frugal.frugal;

/// Frugal's equivalent of Thrift's [TProtocol]. It defines the serialization
/// protocol used for messages, such as JSON, binary, etc. [FProtocol] actually
/// extends [TProtocol] and adds support for serializing [FContext]. In
/// practice, [FProtocol] simply wraps a [TProtocol] and uses [Thrift]'s
/// built-in serialization. [FContext] is encoded before the [TProtocol]
/// serialization of the message using a simple binary protocol. See the
/// protocol documentation for more details.
class FProtocol extends TProtocolDecorator {
  final TTransport _transport;

  FProtocol(TProtocol protocol)
      : this._transport = protocol.transport,
        super(protocol);

  /// Write the request headers on the given [FContext].
  void writeRequestHeader(FContext ctx) {
    transport.writeAll(Headers.encode(ctx.requestHeaders()));
  }

  /// Read the requests headers into a new [FContext].
  FContext readRequestHeader() {
    return new FContext.withRequestHeaders(Headers.read(transport));
  }

  /// Write the response headers on the given [FContext].
  void writeResponseHeader(FContext ctx) {
    transport.writeAll(Headers.encode(ctx.responseHeaders()));
  }

  /// Read the requests headers into the given [FContext].
  void readResponseHeader(FContext ctx) {
    ctx.addResponseHeaders(Headers.read(transport));
  }
}
