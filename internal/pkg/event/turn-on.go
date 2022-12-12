package event

type TurnOnEvent struct {
	Event    Type                 `json:"event"`
	Segments []TurnOnEventSegment `json:"segments"`
}

type TurnOnEventSegment struct {
	Id int `json:"id"`
}

func (e TurnOnEvent) Type() Type {
	return TurnOn
}
