part of frugal;

class _InMemoryTransport extends TTransport {
  final List<int> _buffer;

  _InMemoryTransport(this._buffer);

  @override
  get isOpen => true;

  @override
  Future open() {}

  @override
  Future close() {}

  @override
  Future flush() {}

  @override
  int read(Uint8List buffer, int offset, int length) {}

  /// Writes into this transport's buffer from [buffer].
  @override
  void write(Uint8List buffer, int offset, int length) =>
      _buffer.addAll(buffer.sublist(offset, offset + length));
}

String fObjToJson(Object obj) {
  if (obj is TBase) {
    List<int> buf = [];
    var tt = new _InMemoryTransport(buf);
    var prot = new TJsonProtocol(tt);
    obj.write(prot);
    return new String.fromCharCodes(buf);
  }
  if (obj is FContext) {
    return JSON.encode(obj.requestHeaders());
  }
  return JSON.encode(obj);
}
