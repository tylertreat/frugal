part of frugal;

/// FSubscription is a subscription to a pub/sub topic created by a scope. The
/// topic subscription is actually handled by an FSubscriberTransport, which
/// the FSubscription wraps. Each FSubscription should have its own
/// FSubscriberTransport. The FSubscription is used to unsubscribe from the
/// topic.
class FSubscription {
  String subject;
  FSubscriberTransport _transport;

  FSubscription(this.subject, this._transport);

  /// Unsubscribe from the subject.
  Future unsubscribe() => _transport.unsubscribe();
}
