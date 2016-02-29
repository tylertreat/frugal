part of frugal;

abstract class FTransportMonitor {
  void onClosedCleanly();
  int onClosedUncleanly(Exception cause);
  int onReopenFailed(int prevAttempts, int prevWait);
  void onReopenSucceeded();
}

