package event

type SetColorEvent struct {
	Event
	Segments []SetColorEventSegment `json:"segments"`
}

type SetColorEventSegment struct {
	Id    int     `json:"id"`
	Color [3]byte `json:"color"`
}

func NewSetColorEvent(segments []SetColorEventSegment) SetColorEvent {
	return SetColorEvent{
		Event: Event{
			Type: SetColor,
		},
		Segments: segments,
	}
}
