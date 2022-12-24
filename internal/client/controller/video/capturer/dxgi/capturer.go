package dxgi

import (
	"image"

	"ledctl3/internal/client/controller/video"
)

var scaleFactor = 8

type dxgiCapturer struct {
	displays []video.Display
}

func (c *dxgiCapturer) All() ([]video.Display, error) {
	ds := []video.Display{}

	i := 0
	for {
		d := &display{
			id: i,
		}

		err := d.reset()
		if err != nil {
			break
		}

		bounds := d.ddup.Bounds()
		d.width = bounds.Dx()
		d.height = bounds.Dy()
		d.x = bounds.Min.X
		d.y = bounds.Min.Y

		d.buf = image.NewNRGBA(bounds)

		ds = append(ds, d)

		i++
	}

	return ds, nil
}

func New() (*dxgiCapturer, error) {
	return &dxgiCapturer{}, nil
}
