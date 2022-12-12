package event

import "fmt"

type SetLedsEvent struct {
	Event Type   `json:"event"`
	Id    int    `json:"id"`
	Pix   []byte `json:"pix"`
}

func (e SetLedsEvent) Type() Type {
	return SetLeds
}

func (e SetLedsEvent) String() string {
	return fmt.Sprint(e.Id, len(e.Pix))
}
