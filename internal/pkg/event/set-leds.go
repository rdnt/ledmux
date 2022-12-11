package event

type SetLedsEvent struct {
	Event    Type                  `json:"event"`
	Segments []SetLedsEventSegment `json:"segments"`
}

type SetLedsEventSegment struct {
	Id  int    `json:"id"`
	Pix []byte `json:"pix"`
}

func (e SetLedsEvent) Type() Type {
	return SetLeds
}
