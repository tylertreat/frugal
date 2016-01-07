part of frugal;

var _encoder = new Utf8Encoder();
var _decoder = new Utf8Decoder();

Uint8List encodeHeaders(Map<String, String> headers) {
  var size = 0;
  // Get total frame size headers
  if (headers != null && headers.length > 0) {
    for (var name in headers.keys) {
      // 4 bytes each for name, value length
      size += 8 + name.length + headers[name].length;
    }
  }

  // Header buff = [version (1 byte), size (4 bytes), headers (size bytes)]
  var buff = new Uint8List(5 + size);

  // Write version
  buff[0] = 0x00;

  // Write size
  var bdata = new ByteData.view(buff.buffer);
  bdata.setInt32(1, size);

  // Write headers
  if (headers != null && headers.length > 0) {
    var i = 5;
    for (var name in headers.keys) {
      // Write name length
      bdata.setInt32(i, name.length);
      i += 4;
      // Write name
      buff.setAll(i, _encoder.convert(name));
      i += name.length;

      // Write value length
      var value = headers[name];
      bdata.setInt32(i, value.length);
      i += 4;
      buff.setAll(i, _encoder.convert(value));
      i += value.length;
    }
  }
  return buff;
}

/// Returns the headers from a TTransort
Map<String, String> readHeaders(TTransport transport) {
  // Buffer version
  var buff = new Uint8List(1);
  transport.read(buff, 0, 1);

  // Support more versions when available
  if (buff[0] != 0x00) {
    throw new UnsupportedError("frugal: Unsupported header version ${buff[0]}");
  }

  // Read size
  buff = new Uint8List(4);
  transport.read(buff, 0, 4);
  var bdata = new ByteData.view(buff.buffer);
  var size = bdata.getInt32(0);
  
  return _readHeaderPairs(transport, size);
}

/// Returns the headers from Frugal frame
Map<String, String> decodeHeadersFromFrame(Uint8List frame) {
  if (frame.length < 5) {
    throw new StateError("frugal: invalid frame size ${frame.length}");
  }

  // Support more versions when available
  if (frame[0] != 0x00) {
    throw new UnsupportedError("frugal: Unsupported header version ${frame[0]}");
  }

  var bdata = new ByteData.view(frame.buffer);
  var size = bdata.getInt32(1);
  var transport = new TUint8List(frame.sublist(5, 5+size));
  return _readHeaderPairs(transport, size);
}

Map<String, String> _readHeaderPairs(TTransport transport, int size) {
  var buff = new Uint8List(size);
  transport.read(buff, 0, size);

  var bdata = new ByteData.view(buff.buffer);
  var headers = {};
  for (var i = 0; i < size; i) {
    var nameSize = bdata.getInt32(i);
    i += 4;
    if (i > size || i+nameSize > size) {
      throw new StateError("frugal: invalid protocol header name");
    }
    var name = _decoder.convert(buff, i, i+nameSize);
    i += nameSize;

    var valueSize = bdata.getInt32(i);
    i += 4;
    if (i > size || i+valueSize > size) {
      throw new StateError("frugal: invalid protocol header value");
    }
    var value = _decoder.convert(buff, i, i+valueSize);
    i += valueSize;

    headers[name] = value;
  }

  return headers;
}
