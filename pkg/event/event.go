package event

// Event represents an incoming event from the journal
type Event struct {
	Event string
	Data  []byte
}

func New(e string, b []byte) Event {
	return Event{
		Event: e,
		Data:  b,
	}
}
