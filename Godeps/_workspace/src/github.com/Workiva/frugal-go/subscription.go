package frugal

// Subscription to a pub/sub topic.
type Subscription struct {
	Topic      string
	fTransport FTransport
	errorC     chan error
}

// NewSubscription creates a new Subscription to the given topic which should
// be subscribed on the given FTransport.
func NewSubscription(topic string, transport FTransport) *Subscription {
	return &Subscription{
		Topic:      topic,
		fTransport: transport,
		errorC:     make(chan error, 1),
	}
}

// Unsubscribe from the topic.
func (s *Subscription) Unsubscribe() error {
	return s.fTransport.Unsubscribe()
}

// Error returns a channel which is signaled when something went wrong with the
// subscription. If an error is returned on this channel, the Subscription has
// been closed.
func (s *Subscription) Error() <-chan error {
	return s.errorC
}

// Signal is used to indicate an error on the subscription. This is to be used
// by generated code.
func (s *Subscription) Signal(err error) {
	s.errorC <- err
	close(s.errorC)
}
