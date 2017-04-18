part of frugal.src.frugal;

/// An internal callback which is constructed by generated code and invoked when
/// a pub/sub message is received. An FAsyncCallback is passed an
/// in-memory [TTransport] which wraps the complete message. The callback
/// returns an error or throws an exception if an unrecoverable error occurs and
/// the transport needs to be shutdown.
typedef void FAsyncCallback(TTransport transport);

/// Transport layer for scope subscribers.
abstract class FSubscriberTransport {
  /// Queries whether the transport is subscribed to a topic.
  /// Returns [true] if the transport is subscribed to a topic.
  bool get isSubscribed;

  /// Sets the subscribe topic and opens the transport.
  Future<Null> subscribe(String topic, FAsyncCallback callback);

  /// Unsets the subscribe topic and closes the transport.
  Future<Null> unsubscribe();
}

/// Produces [FSubscriberTransport] instances.
abstract class FSubscriberTransportFactory {
  /// Return a new [FSubscriberTransport] instance.
  FSubscriberTransport getTransport();
}

/// An internal callback which is constructed by generated code and invoked when
/// a durable pub/sub message is received. This function is passed an
/// in-memory [TTransport] which wraps the complete message and the groupId
/// of the message.
typedef void FDurableAsyncCallback(TTransport transport, String groupId);

/// Transport layer for durable scope subscribers.
abstract class FDurableSubscriberTransport {
  /// Queries whether the transport is subscribed to a topic.
  bool get isSubscribed;

  /// Sets the subscription topic and opens the transports.
  Future<Null> subscribe(String topic, FDurableAsyncCallback callback);

  /// Unsets the subscription topic and closes the transport.
  Future<Null> unsubscribe();
}

/// Produces [FDurableSubscriberTransport] instances.
abstract class FDurableSubscriberTransportFactory {
  /// Returns a new [FSubscriberTransport] instance.
  FDurableSubscriberTransport getTransport();
}
