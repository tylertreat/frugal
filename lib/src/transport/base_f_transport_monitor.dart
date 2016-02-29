part of frugal;

class BaseFTransportMonitor extends FTransportMonitor {
  static const int DEFAULT_MAX_REOPEN_ATTEMPTS = 60;
  static const int DEFAULT_INITIAL_WAIT = 2000;
  static const int DEFAULT_MAX_WAIT = 2000;

  int _maxReopenAttempts;
  int _initialWait;
  int _maxWait;

  BaseFTransportMonitor({maxReopenAttempts: DEFAULT_MAX_REOPEN_ATTEMPTS,
                        initialWait: DEFAULT_INITIAL_WAIT,
                        maxWait: DEFAULT_MAX_WAIT}) {
    this._maxReopenAttempts = maxReopenAttempts;
    this._initialWait = initialWait;
    this._maxWait = maxWait;
  }

  @override
  void onClosedCleanly() {}

  @override
  int onClosedUncleanly(Exception cause) {
    return _maxReopenAttempts > 0 ? _initialWait : -1;
  }

  @override
  int onReopenFailed(int prevAttempts, int prevWait) {
    if(prevAttempts >= _maxReopenAttempts) {
      return -1;
    }

    return (prevWait * 2).clamp(0, _maxWait);
  }

  @override
  void onReopenSucceeded() {}
}
