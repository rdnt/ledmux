package event

type TurnOffEvent struct {
	Event
	Segments []TurnOffEventSegment `json:"segments"`
}

type TurnOffEventSegment struct {
	Id int `json:"id"`
}

func NewTurnOffEvent(segments []TurnOffEventSegment) TurnOffEvent {
	return TurnOffEvent{
		Event: Event{
			Type: TurnOff,
		},
		Segments: segments,
	}
}
