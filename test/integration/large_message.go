package integration

import (
	"testing"

	"github.com/Workiva/frugal-go"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/Workiva/thrift/lib/go/thrift"
	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
)

func LargeMessage(
	t *testing.T,
	protocolFactory thrift.TProtocolFactory,
	transportFactory thrift.TTransportFactory,
	conn *nats.Conn,
	// Name is used to display the protocol and transport in each test case
	name string,
) {
	CheckShort(t)

	t.Logf("Testing large messages with %v", name)
	factory := frugal.NewFNatsTransportFactory(conn)
	provider := frugal.NewProvider(factory, transportFactory, protocolFactory)
	publisher := event.NewEventsPublisher(provider)

	evt := event.NewEvent()

	// Publish event
	evt.Message = WarAndPeace
	if err := publisher.PublishEventCreated(name, evt); err != nil {
		assert.Error(t, err, "*event.Event.Message (2) field write error: Message is too large")
	} else {
		t.Errorf("Sending message larger than 1MB succeeded on %v", name)
	}
}
