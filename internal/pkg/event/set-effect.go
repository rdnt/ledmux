package event

type Effect string

const (
	Rainbow   Effect = "rainbow"
	Christmas Effect = "christmas"
)

type SetEffectEvent struct {
	Event     Type   `json:"event"`
	SegmentId int    `json:"segmentId"`
	Effect    Effect `json:"effect"`
}

func (e SetEffectEvent) Type() Type {
	return SetEffect
}
