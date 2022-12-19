package event

type TurnOffEvent struct {
	Event     Type `json:"event"`
	SegmentId int  `json:"segmentId"`
}

func (e TurnOffEvent) Type() Type {
	return TurnOff
}
