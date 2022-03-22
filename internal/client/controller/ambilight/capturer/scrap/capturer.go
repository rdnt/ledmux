package scrap

import (
	"github.com/kbinani/screenshot"
	"golang.org/x/image/draw"
	"ledctl3/internal/client/interfaces"
)

var scaleFactor = 8

type dxgiCapturer struct {
	displays []interfaces.Display
}

func (c *dxgiCapturer) All() ([]interfaces.Display, error) {
	ds := []interfaces.Display{}

	count := screenshot.NumActiveDisplays()
	for i := 0; i < count; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		d := &display{
			id:     i,
			width:  bounds.Dx(),
			height: bounds.Dy(),
			x:      bounds.Min.X,
			y:      bounds.Min.Y,
			scaler: draw.BiLinear.NewScaler(
				bounds.Dx(), bounds.Dy(), bounds.Dx()/scaleFactor, bounds.Dy()/scaleFactor,
			),
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
