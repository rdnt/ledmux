package event

type SetColorEvent struct {
	Event    Type                   `json:"event"`
	Segments []SetColorEventSegment `json:"segments"`
}

type SetColorEventSegment struct {
	Id    int     `json:"id"`
	Color [3]byte `json:"color"`
}

func (e SetColorEvent) Type() Type {
	return SetColor
}
