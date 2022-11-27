package event

type SetGradientEvent struct {
	Event
	Segments []SetGradientEventSegment `json:"segments"`
}

type SetGradientEventSegment struct {
	Id    int                    `json:"id"`
	Steps []SetGradientEventStep `json:"colors"`
}

type SetGradientEventStep struct {
	Color    [3]byte `json:"colors"`
	Position float64 `json:"position"`
}

func NewSetGradientEvent(segments []SetGradientEventSegment) SetGradientEvent {
	return SetGradientEvent{
		Event: Event{
			Type: SetGradient,
		},
		Segments: segments,
	}
}
