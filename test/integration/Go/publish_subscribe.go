package integration

import (
	"strconv"
	"testing"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/Workiva/frugal-go"
	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/nats-io/nats"
)

func PublishSubscribe(
	t *testing.T,
	protocolFactory thrift.TProtocolFactory,
	transportFactory thrift.TTransportFactory,
	conn *nats.Conn,
	// Name is used to display the protocol and transport in each test case
	name string,
) {

	CheckShort(t)

	factory := frugal.NewFNatsTransportFactory(conn)
	provider := frugal.NewProvider(factory, transportFactory, protocolFactory)
	publisher := event.NewEventsPublisher(provider)
	subscriber := event.NewEventsSubscriber(provider)

	evt := event.NewEvent()

	started := make(chan bool)
	wait := make(chan bool)
	done := make(chan struct{})

	expected := new(expectedMessages)
	expected.messageList = make(map[string]bool)

	for i := 1; i < 6; i++ {
		expected.messageList["Foo, Bar"+strconv.Itoa(i)] = false
	}

	go messageHandler(t, subscriber, started, wait, done, expected, name)
	<-started

	// Publish events
	for i := 1; i < 6; i++ {
		evt.Message = "Foo, Bar" + strconv.Itoa(i)
		if err := publisher.PublishEventCreated(name, evt); err != nil {
			panic(err)
		}
	}

	<-done
}
