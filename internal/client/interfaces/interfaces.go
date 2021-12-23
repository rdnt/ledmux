package interfaces

import (
	"context"
	"golang.org/x/image/draw"
)

type DisplayRepository interface {
	All() ([]Display, error)
}

type Display interface {
	Id() int
	Width() int
	Height() int
	X() int
	Y() int
	Scaler() draw.Scaler
	Resolution() string
	String() string
	Capture(ctx context.Context) chan []byte
	SyncCapture(ctx context.Context, frames chan []byte)
}

type AudioSource interface {
	Transformer() Visualizer
}

// Visualizer takes an input and transforms it to an output for the LEDs.
// The output slice length must be a multiple of 4 (RGBA pairs).
type Visualizer interface {
	Start() error
	Events() chan UpdateEvent
	Stop() error
}

type UpdateEvent struct {
	Display Display
	Data []byte
}

type Updater interface {
	Update(mode string, displayId string, b []byte) error
}

type NotifyFunc func(displayId int, b []byte)

type UpdateFunc func(displayId int, b []byte) error
