package event

type Effect string

const (
	Rainbow   Effect = "rainbow"
	Christmas Effect = "christmas"
)

type SetEffectEvent struct {
	Event
	Segments []SetEffectEventSegment `json:"segments"`
}

type SetEffectEventSegment struct {
	Id     int    `json:"id"`
	Effect Effect `json:"effect"`
}

func NewSetEffectEvent(segments []SetEffectEventSegment) SetEffectEvent {
	return SetEffectEvent{
		Event: Event{
			Type: SetEffect,
		},
		Segments: segments,
	}
}
