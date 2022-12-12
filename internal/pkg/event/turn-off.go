package event

type TurnOffEvent struct {
	Event    Type                  `json:"event"`
	Segments []TurnOffEventSegment `json:"segments"`
}

type TurnOffEventSegment struct {
	Id int `json:"id"`
}

func (e TurnOffEvent) Type() Type {
	return TurnOff
}
