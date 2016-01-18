part of frugal;

/**
 * This is an internal-only class. Don't use it!
 */
class WriteByteBuffer {
  final Uint8List _buff;
  int _writeIndex = 0;

  WriteByteBuffer(int capacity)
    : this._buff = new Uint8List(capacity);

  WriteByteBuffer.fromUInt8List(Uint8List buff)
    : this._buff = buff;

  int get remaining => _buff.length - _writeIndex;

  int write(Uint8List buffer, int offset, int length) {
    var amtToCopy = length > remaining ? remaining : length;
    if (amtToCopy > 0) {
      // See how many bytes are source buffer
      var bytesInSource = buffer.length - offset;
      amtToCopy = amtToCopy > bytesInSource ? bytesInSource : amtToCopy;
      // Write the appropriate range
      if (amtToCopy > 0) {
        var range = buffer.getRange(offset, offset + amtToCopy);
        _buff.setRange(_writeIndex, _writeIndex + amtToCopy, range);
      }
    }
    _writeIndex += amtToCopy;
    return amtToCopy;
  }

  Uint8List asUint8List() {
    return new Uint8List.fromList(_buff.sublist(0, _writeIndex));
  }

  void reset() {
    _writeIndex = 0;
  }
}

/**
 * This is an internal-only class. Don't use it!
 */
class ReadByteBuffer {
  final DoubleLinkedQueue _buff = new DoubleLinkedQueue();

  int get remaining => _buff.length;

  int read(Uint8List buffer, int offset, int length) {
    // See how many bytes are source buffer
    var bytesInGiven = buffer.length - offset;
    var amtToRead = length > bytesInGiven ? bytesInGiven : length;
    for (int i = 0; i < amtToRead; i++) {
      buffer[i + offset] = _buff.removeFirst();
    }
    return amtToRead;
  }

  void add(Uint8List buffer) {
    _buff.addAll(buffer);
  }
}
