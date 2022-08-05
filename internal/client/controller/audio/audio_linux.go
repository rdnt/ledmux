package audio

import (
	"ledctl3/internal/client/visualizer"
)

type Visualizer struct {
	leds        int
	segments    []Segment
	maxLedCount int
}

type Segment struct {
	Id   int
	Leds int
}

func (v *Visualizer) Start() error {
	panic("not implemented")
}

func (v *Visualizer) Events() chan visualizer.UpdateEvent {
	panic("not implemented")
}

func (v *Visualizer) Stop() error {
	panic("not implemented")
}
