part of frugal.frugal;

/// A subscription to a pub/sub topic created by a scope. The topic subscription
/// is actually handled by an [FSubscriberTransport], which the [FSubscription]
/// wraps. Each [FSubscription] should have its own [FSubscriberTransport]. The
/// [FSubscription] is used to unsubscribe from the topic.
class FSubscription {
  final String topic;
  FSubscriberTransport _transport;

  FSubscription(this.topic, this._transport);

  /// Unsubscribe from the topic.
  Future unsubscribe() => _transport.unsubscribe();
}
