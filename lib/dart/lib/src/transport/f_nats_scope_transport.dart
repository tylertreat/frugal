part of frugal;

/// Implementation of FScopeTransport backed by NATS client
class FNatsScopeTransport extends FScopeTransport {
  static const _NATS_MAX_MESSAGE_SIZE = 1024 * 1024;

  Nats _nats;
  String _subject;
  Stream<Message> _subscription;
  bool _subscriber;

  StreamController _error = new StreamController.broadcast();
  Stream<Error> get error => _error.stream;

  FByteBuffer _writeBuffer;
  FAsyncCallback _callback;

  FNatsScopeTransport(Nats this._nats);

  @override
  bool get isOpen {
    if (!_nats.isConnected) {
      return false;
    }
    // publisher has to have a write buffer
    if (!_subscriber) {
      return _writeBuffer != null;
    }
    return _subject != null && _callback != null && _subscription != null;
  }

  @override
  void setTopic(String topic) {
    if(_subscriber) {
      throw new FError.withMessage('subscriber cannot set topic');
    }
    this._subject = topic;
  }

  @override
  Future subscribe(String topic, FAsyncCallback callback) async {
    this._subscriber = true;
    this._subject = topic;
    this._callback = callback;
    await open();
  }

  @override
  Future open() async {
    if(!_nats.isConnected) {
      throw new FError.withMessage('NATS is not connected');
    }

    if(!_subscriber) {
      _writeBuffer = new FByteBuffer(_NATS_MAX_MESSAGE_SIZE);
      return;
    }

    if(_subject == null || _subject.isEmpty) {
      throw new FError.withMessage('cannot subscribe to empty topic');
    }

    _subscription = await _nats.subscribe(_subject).catchError((e) {
      throw new TTransportError(e);
    });
    _subscription.listen((Message message) {
      this._callback(new TMemoryTransport.fromUnt8List(message.payload));
    }, onError: (Error e) {
      _error.addError(e);
      close();
    });
  }

  @override
  Future close() async {
    if (!_subscriber) {
      _writeBuffer = null;
      return;
    }
    _callback = null;
    if (_nats.isConnected) {
      await _nats.unsubscribe(_subject);
      _subscription = null;
    }
  }

  @override
  void write(Uint8List buffer, int offset, int length) {
    if (_writeBuffer.writeRemaining < length) {
      _writeBuffer.reset();
      throw new FError.withMessage(
          'Message is too large. Maximum capacity for a NATS message is $_NATS_MAX_MESSAGE_SIZE bytes.'
      );
    }
    _writeBuffer.write(buffer, offset, length);
  }

  @override
  Future flush() async {
    var bytes = _writeBuffer.asUint8List();
    _writeBuffer.reset();
    _nats.publish(_subject, "", bytes);
  }
}
