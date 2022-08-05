package dxgi

import (
	"fmt"
	"image"

	"github.com/kbinani/screenshot"

	"ledctl3/internal/client/controller/video"
)

var scaleFactor = 8

type dxgiCapturer struct {
	displays []video.Display
}

func (c *dxgiCapturer) All() ([]video.Display, error) {
	ds := []video.Display{}

	count := screenshot.NumActiveDisplays()
	for i := 0; i < count; i++ {
		d := &display{
			id: i,
		}

		err := d.reset()
		if err != nil {
			fmt.Println("failed to reset display from All", i, err)
			continue
		}

		bounds := d.ddup.Bounds()
		//
		d.width = bounds.Dx()
		d.height = bounds.Dy()
		d.x = bounds.Min.X
		d.y = bounds.Min.Y

		//bounds = image.Rect(0, 0, 2560, 1440)
		//d.width = 2560
		//d.height = 1440
		//d.x = 0
		//d.y = 0

		d.buf = image.NewNRGBA(bounds)

		ds = append(ds, d)
	}

	return ds, nil
}

func New() (*dxgiCapturer, error) {
	return &dxgiCapturer{}, nil
}
