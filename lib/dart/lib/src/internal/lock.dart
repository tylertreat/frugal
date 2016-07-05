part of frugal;

class Lock {
  Completer _lock = null;

  Future lock() async {
    while (_lock != null && !_lock.isCompleted) {
      await _lock.future;
    }
    _lock = new Completer();
  }

  void unlock() {
    _lock.complete();
  }
}
