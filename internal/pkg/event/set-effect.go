package event

type Effect string

const (
	Rainbow   Effect = "rainbow"
	Christmas Effect = "christmas"
)

type SetEffectEvent struct {
	Event    Type                    `json:"event"`
	Segments []SetEffectEventSegment `json:"segments"`
}

type SetEffectEventSegment struct {
	Id     int    `json:"id"`
	Effect Effect `json:"effect"`
}

func (e SetEffectEvent) Type() Type {
	return SetEffect
}
