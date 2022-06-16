package broker

import (
	"fmt"
	"testing"
)

type EventType string

const (
	Connected EventType = "connected"
)

type ConnectedEvent struct {
	connected bool
}

func TestBroker(t *testing.T) {
	conEvt := ConnectedEvent{
		connected: true,
	}

	//disEvt := DisconnectedEvent{
	//	disconnecte: true,
	//}

	dConnect := New[ConnectedEvent]()
	//dDisconnect := New[EventType, DisconnectedEvent]()

	dConnect.Subscribe(func(e ConnectedEvent) {
		fmt.Println("Event received!", e.connected)
	})

	dConnect.Publish(conEvt)

	// output: Event received! test-event
}
