library frugal.subscription;

import "dart:async";
import "transport/transport.dart";

class Subscription {
  String subject;
  Transport transport;
  StreamController _errorControler = new StreamController.broadcast();
  Stream<Error> get error => _errorControler.stream;

  Subscription(this.subject, this.transport);

  Future unsubscribe() => transport.unsubscribe();

  signal(Error err) { _errorControler.add(err); }
}
