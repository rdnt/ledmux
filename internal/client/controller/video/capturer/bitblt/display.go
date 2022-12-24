package bitblt

import (
	"context"
	"fmt"
	"time"

	"github.com/kbinani/screenshot"

	"ledctl3/internal/client/controller/video"
)

type display struct {
	id     int
	width  int
	height int
	x      int
	y      int
}

func (d *display) Id() int {
	return d.id
}

func (d *display) Width() int {
	return d.width
}

func (d *display) Height() int {
	return d.height
}

func (d *display) X() int {
	return d.x
}

func (d *display) Y() int {
	return d.y
}

func (d *display) Resolution() string {
	return fmt.Sprintf("%dx%d", d.width, d.height)
}

func (d *display) Orientation() video.Orientation {
	if d.width > d.height {
		return video.Landscape
	}

	return video.Portrait
}

func (d *display) String() string {
	return fmt.Sprintf("{id: %d, width: %d, height: %d, left: %d, top: %d}", d.id, d.width, d.height, d.x, d.y)
}

func (d *display) Capture(ctx context.Context, framerate int) chan []byte {
	frames := make(chan []byte)

	go func() {
		ticker := time.NewTicker(time.Duration(1000/framerate) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println("close chan")
				close(frames)
				return
			default:
				pix, err := d.nextFrame()
				if err != nil {
					fmt.Println("err in next", err)
					close(frames)
					return
				}

				frames <- pix
			}
		}
	}()

	return frames
}

func (d *display) nextFrame() ([]byte, error) {
	img, err := screenshot.Capture(d.x, d.y, d.width, d.height)
	if err != nil {
		return nil, err
	}

	return img.Pix, nil
}

func (d *display) Close() error {
	return nil
}
