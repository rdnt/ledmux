package main

import (
	"errors"
	"fmt"
	"github.com/kbinani/screenshot"
	"github.com/kirides/screencapture/d3d"
	"image"
	"image/png"
	"os"
)

func main() {
	n := 1

	device, deviceCtx, err := d3d.NewD3D11Device()
	if err != nil {
		fmt.Printf("Could not create D3D11 Device. %v\n", err)
		return
	}

	defer device.Release()
	defer deviceCtx.Release()

	var ddup *d3d.OutputDuplicator
	defer func() {
		if ddup != nil {
			ddup.Release()
			ddup = nil
		}
	}()

	// Create image that can contain the wanted output (desktop)
	finalBounds := screenshot.GetDisplayBounds(n)
	imgBuf := image.NewRGBA(finalBounds)
	lastBounds := finalBounds

	for {
		bounds := screenshot.GetDisplayBounds(n)
		newBounds := image.Rect(0, 0, int(bounds.Dx()), int(bounds.Dy()))
		if newBounds != lastBounds {
			lastBounds = newBounds
			imgBuf = image.NewRGBA(lastBounds)

			// Throw away old ddup
			if ddup != nil {
				ddup.Release()
				ddup = nil
			}
		}
		// create output duplication if doesn't exist yet (maybe due to resolution change)
		if ddup == nil {
			ddup, err = d3d.NewIDXGIOutputDuplication(device, deviceCtx, uint(n))
			if err != nil {
				fmt.Printf("err: %v\n", err)
				continue
			}
		}

		// Grab an image.RGBA from the current output presenter
		err = ddup.GetImage(imgBuf, 0)
		if err != nil {
			if errors.Is(err, d3d.ErrNoImageYet) {
				// don't update
				continue
			}
			fmt.Printf("Err ddup.GetImage: %v\n", err)
			os.Exit(1)
			// Retry with new ddup, can occur when changing resolution
			ddup.Release()
			ddup = nil
			continue
		}
		//encodeJpeg(buf, imgBuf, opts)

		f, err := os.Create("poc.png")
		if err != nil {
			fmt.Println(err)
		}
		defer f.Close()

		// Encode to `PNG` with `DefaultCompression` level
		// then save to file
		err = png.Encode(f, imgBuf)
		if err != nil {
			panic(err)
		}
		break
	}
}
