part of frugal.src.frugal;

/// A default monitor implementation that attempts to reopen a closed transport
/// with exponential backoff behavior and a capped number of retries. Its
/// behavior can be customized by extending this class and overriding desired
/// callbacks.
class BaseFTransportMonitor extends FTransportMonitor {
  static const int defaultMaxReopenAttempts = 60;
  static const int defaultInitialWait = 2000;
  static const int defaultMaxWait = 2000;

  int _maxReopenAttempts;
  int _initialWait;
  int _maxWait;

  StreamController _onConnectController = new StreamController.broadcast();
  StreamController _onDisconnectController = new StreamController.broadcast();

  bool _isConnected = true;

  /// Create a [BaseFTransportMonitor] with default parameters.
  BaseFTransportMonitor(
      {maxReopenAttempts: defaultMaxReopenAttempts,
      initialWait: defaultInitialWait,
      maxWait: defaultMaxWait}) {
    this._maxReopenAttempts = maxReopenAttempts;
    this._initialWait = initialWait;
    this._maxWait = maxWait;
  }

  Stream get onConnect => _onConnectController.stream;
  Stream get onDisconnect => _onDisconnectController.stream;
  bool get isConnected => _isConnected;

  @override
  void onClosedCleanly() {
    _isConnected = false;
    _onDisconnectController.add(null);
  }

  @override
  int onClosedUncleanly(Object cause) {
    _isConnected = false;
    _onDisconnectController.add(cause);

    return _maxReopenAttempts > 0 ? _initialWait : -1;
  }

  @override
  int onReopenFailed(int prevAttempts, int prevWait) {
    if (prevAttempts >= _maxReopenAttempts) {
      return -1;
    }

    return (prevWait * 2).clamp(0, _maxWait);
  }

  @override
  void onReopenSucceeded() {
    _isConnected = true;
    _onConnectController.add(null);
  }
}
