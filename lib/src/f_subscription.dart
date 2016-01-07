part of frugal;

/// Frugal Subscription
class FSubscription {
  String subject;
  FScopeTransport _transport;
  StreamController _errorControler = new StreamController.broadcast();
  Stream<Error> get error => _errorControler.stream;

  FSubscription(this.subject, this._transport);

  /// Unsubscribe from the subject.
  Future unsubscribe() => _transport.unsubscribe();

  signal(Error err) { _errorControler.add(err); }
}
