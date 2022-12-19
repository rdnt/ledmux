package event

type SetGradientEvent struct {
	Event     Type                   `json:"event"`
	SegmentId int                    `json:"segmentId"`
	Steps     []SetGradientEventStep `json:"steps"`
}

type SetGradientEventStep struct {
	Color    string  `json:"color"`
	Position float64 `json:"position"`
}

func (e SetGradientEvent) Type() Type {
	return SetGradient
}
