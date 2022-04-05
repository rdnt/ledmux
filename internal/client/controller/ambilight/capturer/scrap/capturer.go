package scrap

import (
	"github.com/kbinani/screenshot"
	"ledctl3/internal/client/controller/ambilight"
)

var scaleFactor = 8

type dxgiCapturer struct {
	displays []ambilight.Display
}

func (c *dxgiCapturer) All() ([]ambilight.Display, error) {
	ds := []ambilight.Display{}

	count := screenshot.NumActiveDisplays()
	for i := 0; i < count; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		d := &display{
			id:     i,
			width:  bounds.Dx(),
			height: bounds.Dy(),
			x:      bounds.Min.X,
			y:      bounds.Min.Y,
		}

		err := d.reset()
		if err != nil {
			continue
		}

		ds = append(ds, d)
	}

	return ds, nil
}

func New() (*dxgiCapturer, error) {
	return &dxgiCapturer{}, nil
}
