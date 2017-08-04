package frugal

// FSubscription is a subscription to a pub/sub topic created by a scope. The
// topic subscription is actually handled by an FScopeTransport, which the
// FSubscription wraps. Each FSubscription should have its own FScopeTransport.
// The FSubscription is used to unsubscribe from the topic.
type FSubscription struct {
	topic     string
	transport FSubscriberTransport
	errorC    chan error
}

type suspender interface {
	Suspend() error
}

// NewFSubscription creates a new FSubscription to the given topic which should
// be subscribed on the given FScopeTransport. This is to be used by generated
// code and should not be called directly.
func NewFSubscription(topic string, transport FSubscriberTransport) *FSubscription {
	return &FSubscription{
		topic:     topic,
		transport: transport,
		errorC:    make(chan error, 1),
	}
}

// Unsubscribe from the topic.
func (s *FSubscription) Unsubscribe() error {
	return s.transport.Unsubscribe()
}

// Suspend unsubscribes without removing durable information on the server,
// if applicable
func (s *FSubscription) Suspend() error {
	// If the subscriber transport has a suspend method, use it
	// otherwise call unsubscribe
	// TODO 3.0 get rid of this
	if suspender, ok := s.transport.(suspender); ok {
		return suspender.Suspend()
	}
	return s.transport.Unsubscribe()
}

// Topic returns the subscription topic name.
func (s *FSubscription) Topic() string {
	return s.topic
}
