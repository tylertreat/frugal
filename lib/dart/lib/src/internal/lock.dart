part of frugal;

/// Lock acts like a mutex, disallowing execution of critical sections of code
/// concurrently. A `Completer` can be incorrect because all sleeping threads
/// are woken when the completer completes, allowing multiple threads to
/// enter the critical section. This class addresses that problem.
class Lock {
  Completer _lock = null;

  /// Locks the mutex, disallowing another thread from acquiring the lock
  /// until unlock() is called.
  Future lock() async {
    while (_lock != null && !_lock.isCompleted) {
      await _lock.future;
    }
    _lock = new Completer();
  }

  /// Unlocks the mutex.
  void unlock() {
    _lock.complete();
  }
}
