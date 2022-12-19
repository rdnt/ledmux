package event

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Type string

const (
	Connected   Type = "connected"
	Update      Type = "update"
	SetLeds     Type = "setLeds"
	SetColor    Type = "setColor"
	SetEffect   Type = "setEffect"
	SetGradient Type = "setGradient"
	TurnOn      Type = "turnOn"
	TurnOff     Type = "turnOff"
)

type Event interface {
	Type() Type
}

type EventType struct {
	Event Type `json:"event"`
}

// Parse parses a single event object or an array of event objects
// into a slice of Event. The slice will contain at least one element if
// an error is not returned.
func Parse(b []byte) ([]Event, error) {
	b2 := bytes.TrimLeft(b, " \t\r\n")

	var events []Event

	switch {
	case len(b2) > 0 && b2[0] == '{':
		// parse event
		e, err := parseEvent(b)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	case len(b2) > 0 && b2[0] == '[':
		// parse an array of events
		evts, err := parseEventArray(b)
		if err != nil {
			return nil, err
		}

		events = evts
	default:
		return nil, errors.New("invalid message")
	}

	return events, nil
}

func parseEvent(b []byte) (Event, error) {
	// parse once to get the event type
	var et EventType
	err := json.Unmarshal(b, &et)
	if err != nil {
		return nil, err
	}

	e, err := FromJSON(et.Event, b)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func parseEventArray(b []byte) ([]Event, error) {
	var ets []EventType
	err := json.Unmarshal(b, &ets)
	if err != nil {
		return nil, err
	}

	events := make([]Event, len(ets))

	// create new decoder to parse the actual events based on the types
	dec := json.NewDecoder(bytes.NewReader(b))

	// read the square bracket of the JSON array again
	_, _ = dec.Token()

	// for each event, decode it based on the type we parsed earlier
	for i, typ := range ets {
		var rm json.RawMessage

		err = dec.Decode(&rm)
		if err != nil {
			return nil, err
		}

		e, err := FromJSON(typ.Event, rm)
		if err != nil {
			return nil, err
		}

		events[i] = e
	}

	return events, nil
}

func FromJSON(typ Type, b []byte) (Event, error) {
	switch typ {
	case Connected:
		var e ConnectedEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case SetColor:
		var e SetColorEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case SetEffect:
		var e SetEffectEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case SetGradient:
		var e SetGradientEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case SetLeds:
		var e SetLedsEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case TurnOff:
		var e TurnOffEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case TurnOn:
		var e TurnOnEvent
		err := json.Unmarshal(b, &e)
		return e, err
	case Update:
		var e UpdateEvent
		err := json.Unmarshal(b, &e)
		return e, err
	default:
		return nil, errors.New("invalid type")
	}
}
