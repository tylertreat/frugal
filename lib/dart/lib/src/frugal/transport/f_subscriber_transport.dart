part of frugal.src.frugal;

/// Transport layer for scope subscribers.
abstract class FSubscriberTransport {
  /// Queries whether the transport is subscribed to a topic.
  /// Returns [true] if the transport is subscribed to a topic.
  bool get isSubscribed;

  /// Sets the subscribe topic and opens the transport.
  Future subscribe(String topic, FAsyncCallback callback);

  /// Unsets the subscribe topic and closes the transport.
  Future unsubscribe();
}

/// Produces [FSubscriberTransport] instances.
abstract class FSubscriberTransportFactory {
  /// Return a new [FSubscriberTransport] instance.
  FSubscriberTransport getTransport();
}
