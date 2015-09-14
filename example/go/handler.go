package main

import (
	"fmt"

	"github.com/Workiva/frugal/example/go/gen-go/event"
)

type EventPubSubHandler struct{}

func NewEventPubSubHandler() *EventPubSubHandler {
	return &EventPubSubHandler{}
}

func (e *EventPubSubHandler) EventCreated(event *event.Event) error {
	fmt.Println("received", event)
	return nil
}
