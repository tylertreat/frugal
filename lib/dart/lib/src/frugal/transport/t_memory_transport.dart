part of frugal.src.frugal;

/// An [FByteBuffer]-backed implementation of [TTransport].
class TMemoryTransport extends TTransport {
  static const _defaultBufferLength = 1024;

  final FByteBuffer _buff;

  TMemoryTransport([int capacity])
      : _buff = new FByteBuffer(capacity ?? _defaultBufferLength);

  TMemoryTransport.fromUint8List(Uint8List buffer)
      : _buff = new FByteBuffer.fromUint8List(buffer);

  @override
  bool get isOpen => true;

  @override
  Future open() async {}

  @override
  Future close() async {}

  @override
  int read(Uint8List buffer, int offset, int length) {
    return _buff.read(buffer, offset, length);
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    _buff.write(buffer, offset, length);
  }

  @override
  Future flush() async {}

  /// The bytes currently stored in the transport.
  Uint8List get buffer => _buff.asUint8List();
}
