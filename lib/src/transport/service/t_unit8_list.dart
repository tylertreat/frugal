part of frugal;

/// A Uint8List-backed implementation of TTransport.
class TUint8List extends TTransport {

  final Uint8List _buff;
  int _readPos;
  int _writePos;

  TUint8List(this._buff) {
    _readPos = 0;
    _writePos = 0;
  }

  /// A read-only representation of the bytes in the buffer.
  Uint8List get buff => _buff.buffer.asUint8List();

  /// Reset the internal buffer.
  void reset() {
    _readPos = 0;
    _writePos = 0;
  }

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
    // Can read at most what's left in the internal buffer
    var bytesRemaining = _buff.length - _readPos;
    var amtToRead = (length > bytesRemaining ? bytesRemaining : length);
    if (amtToRead > 0) {
      // This range should not throw an exception since we capped above
      var range = _buff.getRange(_readPos, _readPos + amtToRead);
      try {
        buffer.setRange(offset, offset + amtToRead, range);
      } on RangeError catch(e) {
        throw new TTransportError(TTransportErrorType.UNKNOWN, "Unexpected end of input buffer $e");
      }
      _readPos += amtToRead;
    }
    return amtToRead;
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write (Uint8List buffer, int offset, int length) {
    if ((length + _writePos) > _buff.length) {
      throw new TTransportError(TTransportErrorType.UNKNOWN, "Not enough room in output buffer");
    }
    var range;
    try {
      range = buffer.getRange(offset, offset + length);
    } on RangeError catch(e) {
      throw new TTransportError(TTransportErrorType.UNKNOWN, "Unexpected end of input buffer $e");
    }
    // This range should not throw an exception since we capped above
    _buff.setRange(_writePos, _writePos + length, range);
    _writePos += length;
  }

  /// Flush any pending data out of a transport buffer.
  Future flush() {
    return new Future.value();
  }
}