part of frugal;

/// A Uint8List-backed implementation of TTransport.
class TMemoryBuffer extends TTransport {

  final FByteBuffer _buff;

  TMemoryBuffer(Uint8List buff)
    : _buff =new FByteBuffer.fromUInt8List(buff);

  /// A read-only representation of the bytes in the buffer.
  Uint8List get buff => _buff.asUint8List();

  /// Queries whether the transport is open.
  bool get isOpen => true;

  /// Opens the transport for reading/writing.
  Future open() {
    return new Future.value();
  }

  /// Closes the transport.
  Future close() {
    return new Future.value();
  }

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    return _buff.read(buffer, offset, length);
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write(Uint8List buffer, int offset, int length) {
    _buff.write(buffer, offset, length);
  }

  /// Flush any pending data out of a transport buffer.
  Future flush() {
    return new Future.value();
  }
}