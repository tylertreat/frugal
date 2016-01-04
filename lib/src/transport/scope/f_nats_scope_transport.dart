part of frugal;

/// Implementation of FScopeTransport backed by NATS client
class FNatsScopeTransport extends FScopeTransport {
  Nats nats;
  String topic;
  Stream<Message> subscription;
  bool subscriber;

  StreamController _signalRead = new StreamController.broadcast();

  Stream get signalRead => _signalRead.stream;

  StreamController _error = new StreamController.broadcast();

  Stream get error => _error.stream;

  bool _isOpen;
  List<int> _writeBuffer;
  Iterator<int> _readIterator;


  FNatsScopeTransport(Nats this.nats);

  Uint8List _consumeWriteBuffer() {
    Uint8List buffer = new Uint8List.fromList(_writeBuffer);
    _writeBuffer.clear();
    return buffer;
  }

  void _setReadBuffer(Uint8List readBuffer) {
    _readIterator = readBuffer != null ? readBuffer.iterator : null;
  }

  void _reset({bool isOpen: false}) {
    _isOpen = isOpen;
    if(!subscriber) {
      _writeBuffer.clear();
    }
    _readIterator = null;
  }

  bool get hasReadData => _readIterator != null;

  /// Queries whether the transport is open.
  /// Returns [true] if the transport is open.
  bool get isOpen => _isOpen && ((!subscriber && _writeBuffer != null) || (topic != null && subscription != null));

  @override
  void setTopic(String topic) {
    if(subscriber) {
      throw new StateError('frugal: subscriber cannot lock topic');
    }
    this.topic = topic;
  }

  @override
  Future subscribe(String topic) async {
    this.subscriber = true;
    this.topic = topic;
    await open();
  }

  /// Opens the transport for reading/writing.
  /// Throws [TTransportError] if the transport could not be opened.
  Future open() async {
    if(!nats.isConnected) {
      throw new StateError('NATS is not connected');
    }
    if(!subscriber) {
      _writeBuffer = [];
      return;
    }

    if(topic == null) {
      throw new StateError('cannot subscribe to empty subject');
    }

    _reset(isOpen: true);
    subscription = await nats.subscribe(topic).catchError((e) {
      throw new TTransportError(e);
    });
    subscription.listen((Message message) {
      _setReadBuffer(message.payload);
      _signalRead.add(true);
    }, onError: signalSubscriptionErr);
  }

  void signalSubscriptionErr(Error e) {
    _error.addError(e);
  }

  /// Closes the transport.
  Future close() async {
    if (!isOpen) {
      return;
    }
    _reset(isOpen: false);
    if(subscriber) {
      await nats.unsubscribe(topic);
      subscription = null;
    }
  }

  /// Reads up to [length] bytes into [buffer], starting at [offset].
  /// Returns the number of bytes actually read.
  /// Throws [TTransportError] if there was an error reading data
  int read(Uint8List buffer, int offset, int length) {
    if (buffer == null) {
      throw new ArgumentError.notNull("buffer");
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError("The range exceeds the buffer length");
    }

    if (_readIterator == null || length <= 0) {
      return 0;
    }

    int i = 0;
    while (i < length && _readIterator.moveNext()) {
      buffer[offset + i] = _readIterator.current;
      i++;
    }

    // cleanup iterator when we've reached the end
    if (_readIterator.current == null) {
      _readIterator = null;
    }

    return i;
  }

  /// Writes up to [len] bytes from the buffer.
  /// Throws [TTransportError] if there was an error writing data
  void write(Uint8List buffer, int offset, int length) {
    // TODO: Blow up if you go over 1Mb
    if (buffer == null) {
      throw new ArgumentError.notNull("buffer");
    }

    if (offset + length > buffer.length) {
      throw new ArgumentError("The range exceeds the buffer length");
    }

    _writeBuffer.addAll(buffer.sublist(offset, offset + length));
  }

  /// Flush any pending data out of a transport buffer.
  /// Throws [TTransportError] if there was an error writing out data.
  Future flush() async {
    Uint8List bytes = _consumeWriteBuffer();
    nats.publish(topic, "", bytes);
  }
}
