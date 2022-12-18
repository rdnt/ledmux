package event

type SetColorEvent struct {
	Event     Type   `json:"event"`
	SegmentId int    `json:"segmentId"`
	Color     string `json:"color"`
}

func (e SetColorEvent) Type() Type {
	return SetColor
}
