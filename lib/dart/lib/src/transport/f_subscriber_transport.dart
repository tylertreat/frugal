part of frugal;

/// FSubscriberTransport is used exclusively for scope subscribers.
abstract class FSubscriberTransport {
  /// Queries whether the transport is subscribed to a topic.
  /// Returns [true] if the transport is subscribed to a topic.
  bool get isSubscribed;

  /// set the subscribe topic and opens the transport.
  Future subscribe(String topic, FAsyncCallback callback);

  /// unsets the subscribe topic and closes the transport.
  Future unsubscribe();
}

/// FSubscriberTransportFactory produces FSubscriberTransports.
abstract class FSubscriberTransportFactory {
  FSubscriberTransport getTransport();
}
