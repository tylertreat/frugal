part of frugal;

/// Extension of TProtocol with the addition of headers
class FProtocol extends TProtocolDecorator {
  final TProtocol _protocol;

  FProtocol(TProtocol protocol)
    : this._protocol = protocol,
      super(protocol);

  /// write the request headers on the given Context
  void writeRequestHeader(Context ctx) {
    _protocol.transport.writeAll(encodeHeaders(ctx.requestHeaders()));
  }

  /// read the requests headers into a new Context
  Context readRequestHeader() {
    return new Context.withRequestHeaders(readHeaders(_protocol.transport));
  }

  /// write the response headers on the given Context
  void writeResponseHeader(Context ctx) {
    _protocol.transport.writeAll(encodeHeaders(ctx.responseHeaders()));
  }

  /// read the requests headers into the given Context
  void readResponseHeader(Context ctx) {
    ctx.addResponseHeaders(readHeaders(_protocol.transport));
  }
}
