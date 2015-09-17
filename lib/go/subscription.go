package frugal

// Subscription to a pub/sub topic.
type Subscription struct {
	Topic     string
	transport Transport
}

// NewSubscription creates a new Subscription to the given topic which should
// be subscribed on the given Transport.
func NewSubscription(topic string, transport Transport) *Subscription {
	return &Subscription{Topic: topic, transport: transport}
}

// Unsubscribe from the topic.
func (s *Subscription) Unsubscribe() error {
	return s.transport.Unsubscribe()
}
