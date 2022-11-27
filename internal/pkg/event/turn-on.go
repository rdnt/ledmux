package event

type TurnOnEvent struct {
	Event
	Segments []TurnOnEventSegment `json:"segments"`
}

type TurnOnEventSegment struct {
	Id int `json:"id"`
}

func NewTurnOnEvent(segments []TurnOnEventSegment) TurnOnEvent {
	return TurnOnEvent{
		Event: Event{
			Type: TurnOn,
		},
		Segments: segments,
	}
}
