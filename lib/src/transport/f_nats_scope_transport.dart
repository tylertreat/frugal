part of frugal;

/// Implementation of FScopeTransport backed by NATS client
class FNatsScopeTransport extends FScopeTransport {
  static const _NATS_MAX_MESSAGE_SIZE = 1024 * 1024;

  Nats _nats;
  String _subject;
  Stream<Message> _subscription;
  bool _subscriber;

  StreamController _signalRead = new StreamController.broadcast();
  Stream get signalRead => _signalRead.stream;

  StreamController _error = new StreamController.broadcast();
  Stream get error => _error.stream;

  WriteByteBuffer _writeBuffer;
  ReadByteBuffer _readBuffer;

  FNatsScopeTransport(Nats this._nats);

  bool get hasReadData => isOpen && _subscriber && _readBuffer.remaining > 0;

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen {
    if (!_nats.isConnected) {
      return false;
    }

    // publisher has to have a write buffer
    if (!_subscriber) {
      return _writeBuffer != null;
    }
    return _subject != null && _subscription != null && _readBuffer != null;
  }

  @override
  void setTopic(String topic) {
    if(_subscriber) {
      throw new FError.withMessage('subscriber cannot set topic');
    }
    this._subject = topic;
  }

  @override
  Future subscribe(String topic) async {
    this._subscriber = true;
    this._subject = topic;
    await open();
  }

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open() async {
    if(!_nats.isConnected) {
      throw new FError.withMessage('NATS is not connected');
    }

    if(!_subscriber) {
      _writeBuffer = new WriteByteBuffer(_NATS_MAX_MESSAGE_SIZE);
      return;
    }

    if(_subject == null || _subject.isEmpty) {
      throw new FError.withMessage('cannot subscribe to empty topic');
    }

    _subscription = await _nats.subscribe(_subject).catchError((e) {
      throw new TTransportError(e);
    });
    _readBuffer = new ReadByteBuffer();
    _subscription.listen((Message message) {
      _readBuffer.add(message.payload);
      _signalRead.add(true);
    }, onError: signalSubscriptionErr);
  }

  void signalSubscriptionErr(Error e) {
    _error.addError(e);
    close();
  }

  /// Closes the transport.
  Future close() async {
    if (!_subscriber) {
      _writeBuffer = null;
      return;
    }
    _readBuffer = null;
    if (_nats.isConnected) {
      await _nats.unsubscribe(_subject);
      _subscription = null;
    }
  }

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    return _readBuffer.read(buffer, offset, length);
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write(Uint8List buffer, int offset, int length) {
    if (_writeBuffer.remaining < length) {
      _writeBuffer.reset();
      throw new FError.withMessage(
          'Message is too large. Maximum capacity for a NATS message is $_NATS_MAX_MESSAGE_SIZE bytes.'
      );
    }
    _writeBuffer.write(buffer, offset, length);
  }

  /// Flush any pending data out of a transport buffer.
  /// Throws [TTransportError] if there was an error writing out data.
  Future flush() async {
    var bytes = _writeBuffer.asUint8List();
    _writeBuffer.reset();
    _nats.publish(_subject, "", bytes);
  }
}
