package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"ledctl3/internal/pkg/event"
)

type EventWithType struct {
	Event string `json:"event"`
}

type Events []EventWithType

var ErrInvalidMessage = errors.New("invalid message")

func ParseMessage(b []byte) ([]event.Event, error) {
	//var evts Events
	//err := json.Unmarshal(b, &evts)
	//if err == nil {
	//	return evts, nil
	//}

	r := bytes.NewReader(b)
	dec := json.NewDecoder(r)

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		return nil, ErrInvalidMessage
	}

	var events []event.Event

	switch delim {
	case json.Delim('['):
		// read the type for each event
		var eventTypes []EventWithType
		err = json.Unmarshal(b, &eventTypes)
		if err != nil {
			return nil, err
		}

		// seek to the beginning
		_, err = r.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		// create new decoder to parse the actual events based on the types
		dec = json.NewDecoder(r)

		// read the square bracket again
		_, _ = dec.Token()

		// for each event, decode it based on the type we parsed earlier
		for _, typ := range eventTypes {
			e, err := decodeEvent(dec, event.Type(typ.Event))
			if err != nil {
				return nil, err
			}

			events = append(events, e)
		}
	case json.Delim('{'):
		e, err := parseEvent(b)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	default:
		return nil, ErrInvalidMessage
	}

	return events, nil
}

func parseEvent(b []byte) (event.Event, error) {
	// parse once to get the event type
	var evt EventWithType
	err := json.Unmarshal(b, &evt)
	if err != nil {
		return nil, err
	}

	switch event.Type(evt.Event) {
	case event.SetLeds:
		var e event.SetLedsEvent
		err := json.Unmarshal(b, &e)
		if err != nil {
			return nil, err
		}

		return e, nil
	default:
		return nil, ErrInvalidMessage
	}
}

//func decodeEventTypes(dec *json.Decoder, r io.ReadSeeker) ([]event.Type, error) {
//	// parse once to get the event type
//	var types []event.Type
//
//	var evt EventWithType
//	err := dec.Decode(&evt)
//	if err != nil {
//		return nil, err
//	}
//
//	types = append(types, event.Type(evt.EventWithType))
//}

func decodeEvent(dec *json.Decoder, typ event.Type) (event.Event, error) {
	switch typ {
	case event.TurnOn:
		var e event.TurnOnEvent
		err := dec.Decode(&e)
		if err != nil {
			return nil, err
		}

		return e, nil
	case event.SetLeds:
		var e event.SetLedsEvent
		err := dec.Decode(&e)
		if err != nil {
			return nil, err
		}

		return e, nil

	default:
		return nil, ErrInvalidMessage
	}
}
