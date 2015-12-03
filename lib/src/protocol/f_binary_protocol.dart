part of frugal;

/// Binary protocol implementation for Frugal. Extends upon the
/// associated TProtocol.
class FBinaryProtocol extends TBinaryProtocol implements FProtocol {
  FBinaryProtocol(TTransport transport)
    : super(transport);

  /// write the request headers on the given Context
  void writeRequestHeader(Context ctx) {
    _writeHeader(ctx.requestHeaders());
  }

  /// read the requests headers into a new Context
  Context readRequestHeader() {
    // Check version when more are available.
    readByte();

    // Read headers
    var numHeaders = readI16();
    var ctx = new Context("");
    for (var i = 0; i < numHeaders; i++) {
      var key = readString();
      var value = readString();
      ctx.addRequestHeader(key, value);
    }
    return ctx;
  }

  /// write the response headers on the given Context
  void writeResponseHeader(Context ctx) {
    _writeHeader(ctx.requestHeaders());
  }

  /// read the requests headers into the given Context
  void readResponseHeader(Context ctx) {
    // Check version when more are available.
    readByte();

    // Read headers
    var numHeaders = readI16();
    for (var i = 0; i < numHeaders; i++) {
      var key = readString();
      var value = readString();
      ctx.addRequestHeader(key, value);
    }
  }

  _writeHeader(Map<String, String> headers) {
    writeByte(0x00);
    writeI16(headers.length);
    for (var key in headers.keys) {
      writeString(key);
      writeString(headers[key]);
    }
  }
}
