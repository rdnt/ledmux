package event

type SetGradientEvent struct {
	Event    Type                      `json:"event"`
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

func (e SetGradientEvent) Type() Type {
	return SetGradient
}
