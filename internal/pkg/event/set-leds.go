package event

import "fmt"

type SetLedsEvent struct {
	Event     Type   `json:"event"`
	SegmentId int    `json:"segmentId"`
	Pix       []byte `json:"pix"`
}

func (e SetLedsEvent) Type() Type {
	return SetLeds
}

func (e SetLedsEvent) String() string {
	return fmt.Sprint(e.SegmentId, len(e.Pix))
}
