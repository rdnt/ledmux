package visualizer

import (
	"image/color"
	"time"
)

// Visualizer processes some input and transforms it to an output for the LED
// strip segments.
type Visualizer interface {
	Start() error
	Events() chan UpdateEvent
	Stop() error
}

// UpdateEvent is emitted whenever one or more segments need to be updated.
type UpdateEvent struct {
	Segments []Segment
	Latency  time.Duration
}

// Segment is an LED strip segment.
type Segment struct {
	// ID is the unique identifier of the Segment.
	Id int
	// Pix contains color data for each LED. Must be a multiple of 4 (RGBA).
	Pix []color.Color
}
