package integration

import (
	"sync"
	"testing"
	"time"

	"github.com/Workiva/frugal/example/go/gen-go/event"
	"github.com/stretchr/testify/assert"
)

type expectedMessages struct {
	sync.RWMutex
	messageList map[string]bool
}

func CheckShort(t *testing.T) {
	if testing.Short() {
		t.Log("Skipping integration test in short mode")
		t.Skip()
	}
}

func messageHandler(
	t *testing.T,
	subscriber *event.EventsSubscriber,
	// Channel closed once the subscriber is started
	started chan bool,
	// Channel closed after waiting for messages
	wait chan bool,
	// Channel closed at end of function
	done chan struct{},
	// Map of all messages we expect to receive
	expected *expectedMessages,
	// Protocol and Transport, used for channel name
	name string,
) {
	defer close(done)

	// At the end of the test we should have received all of the messages in the
	// expected messages map
	defer func() {
		expected.RLock()
		for expectedMsg, wasReceived := range expected.messageList {
			assert.True(t, wasReceived, "%v was not received", expectedMsg)
		}
		expected.RUnlock()
	}()
	close(started)

	t.Logf("Testing with %v", name)

	_, err := subscriber.SubscribeEventCreated(name, func(e *event.Event) {

		expectedMsgKey := e.GetMessage()

		expected.RLock()
		expectedMsg, ok := expected.messageList[expectedMsgKey]
		if !ok {
			t.Errorf(`unexpected message on %v`, name)
			return
		}
		expected.RUnlock()
		if expectedMsg == true {
			return
		}

		expected.Lock()
		expected.messageList[expectedMsgKey] = true
		expected.Unlock()

		// There is a race condition here.
		// Not sure of a better way to handle this yet.
		allDone := true
		for _, hasBeenReceived := range expected.messageList {
			allDone = allDone && (hasBeenReceived == true)
		}
		if allDone {
			wait <- true
		}
	})
	if err != nil {
		panic(err)
	}

	select {
	case <-wait:
		return
	case <-time.After(time.Second * 4):
		t.Errorf("Test timed out.")
	}

	return
}
