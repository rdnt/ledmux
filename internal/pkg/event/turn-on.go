package event

type TurnOnEvent struct {
	Event     Type `json:"event"`
	SegmentId int  `json:"segmentId"`
}

func (e TurnOnEvent) Type() Type {
	return TurnOn
}
