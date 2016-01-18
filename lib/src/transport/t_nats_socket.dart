part of frugal;

/// A Service TSocket backed by a NATS client
class TNatsSocket implements TSocket {
  static const String DISCONNECT = "DISCONNECT";
  static const int maxMissedHeartbeats = 3;

  final Nats _natsClient;
  final String _listenTo;
  final String _replyTo;
  final String _heartbeatListen;
  final String _heartbeatReply;
  Duration _heartbeatInterval;
  Timer _heartbeatTimer;
  Stream<Message> _heartbeatListenStream;

  final StreamController<TSocketState> _onStateController;
  Stream<TSocketState> get onState => _onStateController.stream;

  final StreamController<Object> _onErrorController;
  Stream<Object> get onError => _onErrorController.stream;

  final StreamController<List<int>> _onMessageController;
  Stream<List<int>> get onMessage => _onMessageController.stream;

  Stream<Message> _listenStream;

  final List<Uint8List> _requests = [];

  int _missed = 0;

  TNatsSocket(this._natsClient, this._listenTo, this._replyTo,
              this._heartbeatListen, this._heartbeatReply,
              Duration readTimeout, this._heartbeatInterval)
  : _onStateController = new StreamController.broadcast(),
  _onErrorController = new StreamController.broadcast(),
  _onMessageController = new StreamController.broadcast() {
  }

  bool get isOpen => _listenStream != null && _natsClient.isConnected;

  bool get isClosed => _listenStream == null;

  Future open() async {
    if (!isClosed) {
      throw new TTransportError(TTransportErrorType.ALREADY_OPEN, 'Socket already connected');
    }

    if (!_natsClient.isConnected) {
      throw new TTransportError(TTransportErrorType.NOT_OPEN, 'NATS client not open');
    }

    // Listen for errors on the Nats client. Pass them along.
    _natsClient.onError.listen((Error e) {
      _onErrorController.add(e);
    });

    // Subscribe to inbox. Return pass along disconnects.
    _listenStream = await _natsClient.subscribe(_listenTo);
    _listenStream.listen((Message msg) {
      if (msg.reply == DISCONNECT) {
        close();
        return;
      }
      _onMessageController.add(msg.payload);
    });

    // Start listening for heartbeats.
    if (_heartbeatInterval.inMilliseconds > 0) {
      _heartbeatListenStream = await _natsClient.subscribe(_heartbeatListen);
      _heartbeatListenStream.listen((Message message) {
        // Send a heartbeat response and clear missed field
        _natsClient.publish(_heartbeatReply, "", new Uint8List.fromList([]));
        _missed = 0;
        _resetHeartbeat();
      });
      _resetHeartbeat();
    }
    _onStateController.add(TSocketState.OPEN);
  }

  void _resetHeartbeat() {
    if (_heartbeatTimer != null) {
      _heartbeatTimer.cancel();
    }
    _heartbeatTimer = new Timer.periodic(_heartbeatInterval, (Timer timer) {
      // increment missed and check the count
      _missed++;
      if (_missed >= maxMissedHeartbeats) {
        close();
      }
    });
  }

  Future close() async {
    if (_heartbeatTimer != null) {
      _heartbeatTimer.cancel();
    }

    _listenStream = null;
    if(_heartbeatInterval.inMilliseconds > 0) {
      if (_natsClient.isConnected) {
        _natsClient.unsubscribe(_heartbeatListen);
      }
      _heartbeatListenStream = null;
    }

    if (_requests.isNotEmpty) {
      _onErrorController
      .add(new FError.withMessage('socket was closed with pending requests'));
    }
    _requests.clear();
    _onStateController.add(TSocketState.CLOSED);

    if (_natsClient.isConnected) {
      _natsClient.unsubscribe(_listenTo);
    }
  }

  void send(Uint8List data) {
    _requests.add(data);
    _sendRequests();
  }

  void _sendRequests() {
    while (isOpen && _requests.isNotEmpty) {
      Uint8List data = _requests.removeAt(0);
      _natsClient.publish(_replyTo, "", data);
    }
  }
}