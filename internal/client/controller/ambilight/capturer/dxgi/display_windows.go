package dxgi

import (
	"context"
	"errors"
	"fmt"
	"github.com/kirides/screencapture/d3d"
	"image"
	"time"
)

var ErrNoFrame = fmt.Errorf("no frame")

type display struct {
	index  int
	id     int
	width  int
	height int
	x      int
	y      int
	buf    *image.RGBA
	dev    *d3d.ID3D11Device
	devCtx *d3d.ID3D11DeviceContext
	ddup   *d3d.OutputDuplicator
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

func (d *display) String() string {
	return fmt.Sprintf("Display{id: %d, width: %d, height: %d, left: %d, top: %d}", d.id, d.width, d.height, d.x, d.y)
}

func (d *display) Capture(ctx context.Context, framerate int) chan []byte {
	frames := make(chan []byte)

	go func() {
		ticker := time.NewTicker(time.Duration(1000/framerate) * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println(d.id, "context done")
				close(frames)
				return
			default:
				pix, err := d.nextFrame()
				if errors.Is(err, ErrNoFrame) {
					//fmt.Println(d.id, "no frame")
					continue
				} else if err != nil {
					fmt.Println(d.id, "non-nil error", err)

					err := d.reset()
					if err != nil {
						fmt.Println(d.id, "failed to reset from capture")
					}

					close(frames)
					return
				}

				if pix == nil {
					fmt.Println(d.id, "invalid frame")
					continue
				}

				//fmt.Println(d.id, "dispatch")
				frames <- pix
			}
		}
	}()

	return frames
}

func (d *display) nextFrame() ([]byte, error) {
	err := d.ddup.GetImage(d.buf, 0)
	if errors.Is(err, d3d.ErrNoImageYet) {
		// don't update
		return nil, ErrNoFrame
	} else if err != nil {
		return nil, err
	}

	return d.buf.Pix, nil
}

func (d *display) reset() error {
	_ = d.Close()

	var err error
	d.dev, d.devCtx, err = d3d.NewD3D11Device()
	if err != nil {
		return err
	}

	d.ddup, err = d3d.NewIDXGIOutputDuplication(d.dev, d.devCtx, uint(d.id))
	if err != nil {
		d.dev.Release()
		d.dev = nil

		d.devCtx.Release()
		d.devCtx = nil
		return err
	}

	return nil
}

func (d *display) Close() error {
	if d.dev != nil {
		d.dev.Release()
		d.dev = nil
	}

	if d.devCtx != nil {
		d.devCtx.Release()
		d.devCtx = nil
	}

	if d.ddup != nil {
		d.ddup.Release()
		d.ddup = nil
	}

	return nil
}
