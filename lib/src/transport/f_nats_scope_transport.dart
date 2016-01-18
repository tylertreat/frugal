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

  FByteBuffer _buffer;

  FNatsScopeTransport(Nats this._nats);

  bool get hasReadData => isOpen && _subscriber && _buffer != null;

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen {
    if (!_nats.isConnected) {
      return false;
    }

    // publisher has to have a write buffer
    if (!_subscriber) {
      return _buffer != null;
    }
    return _subject != null && _subscription != null;
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
      _buffer = new FByteBuffer(_NATS_MAX_MESSAGE_SIZE);
      return;
    }

    if(_subject == null || _subject.isEmpty) {
      throw new FError.withMessage('cannot subscribe to empty topic');
    }

    _subscription = await _nats.subscribe(_subject).catchError((e) {
      throw new TTransportError(e);
    });
    _subscription.listen((Message message) {
      // TODO: Do we need some kind of lock until this thing is read?
      _buffer = new FByteBuffer.fromUInt8List(message.payload);
      _signalRead.add(true);
    }, onError: signalSubscriptionErr);
  }

  void signalSubscriptionErr(Error e) {
    _error.addError(e);
    close();
  }

  /// Closes the transport.
  Future close() async {
    _buffer = null;
    if (!_subscriber) {
      return;
    }
    if (_nats.isConnected) {
      await _nats.unsubscribe(_subject);
      _subscription = null;
    }
  }

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    return _buffer.read(buffer, offset, length);
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write(Uint8List buffer, int offset, int length) {
    if (_buffer.writeRemaining < length) {
      _buffer.reset();
      throw new FError.withMessage(
          'Message is too large. Maximum capacity for a NATS message is $_NATS_MAX_MESSAGE_SIZE bytes.'
      );
    }
    _buffer.write(buffer, offset, length);
  }

  /// Flush any pending data out of a transport buffer.
  /// Throws [TTransportError] if there was an error writing out data.
  Future flush() async {
    var bytes = _buffer.asUint8List();
    _buffer.reset();
    _nats.publish(_subject, "", bytes);
  }
}
