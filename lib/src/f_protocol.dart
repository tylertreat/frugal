part of frugal;

abstract class FProtocolFactory {
  FProtocol getProtocol(TTransport transport);
}

abstract class FProtocol extends TProtocol {
  FProtocol(TTransport transport)
    : super(transport);

  void writeRequestHeader(Context);
  Context readRequestHeader();

  void writeResponseHeader(Context);
  void readResponseHeader(Context);
}

class FBinaryProtocol extends TBinaryProtocol implements FProtocol {
  FBinaryProtocol(TTransport transport)
    : super(transport);

  writeRequestHeader(Context ctx) {
    _writeHeader(ctx.requestHeaders());
  }

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

  writeResponseHeader(Context ctx) {
    _writeHeader(ctx.requestHeaders());
  }

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
