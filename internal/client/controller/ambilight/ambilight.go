package ambilight

import (
	"context"
	"fmt"
	"ledctl3/internal/client/interfaces"
	"sync"
)

type Visualizer struct {
	displays   interfaces.DisplayRepository
	leds       int
	cancel     context.CancelFunc
	done       chan bool
	events     chan interfaces.UpdateEvent
	displayCfg []DisplayConfig
}

type DisplayConfig struct {
	Id           int
	Width        int
	Height       int
	Leds         int
	BoundsOffset int
	BoundsSize   int
}

func (v *Visualizer) Events() chan interfaces.UpdateEvent {
	return v.events
}

func (v *Visualizer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	displays, err := v.displays.All()
	if err != nil {
		return err
	}

	v.done = make(chan bool, len(displays))

	for _, d := range displays {
		pixChan := d.Capture(ctx)

		go func(d interfaces.Display) {
			defer fmt.Println("done")
			for {
				select {
				case pix := <-pixChan:
					// async processing of incoming pix
					go v.process(d, pix)
				case <-ctx.Done():
					if v.done != nil {
						v.done <- true
						close(v.done)
						v.done = nil
					}
					return
				}
			}
		}(d)
	}

	return nil
}

func (v *Visualizer) process(d interfaces.Display, pix []byte) {
	if len(pix) == 0 {
		//fmt.Println("invalid frame", d.Id())
		return
	}

	//if d.Id() != 0 {
	//	return
	//}

	if d.Id() >= len(v.displayCfg) {
		// skip as this display is not in the config
		return
	}

	cfg := v.displayCfg[d.Id()]

	if cfg.Width != d.Width() || cfg.Height != d.Height() {
		// skip as this is an invalid config for this display
		return
	}

	//// bounds are per-display
	//boundsOffset := 5860
	////boundsOffset := 0
	//boundsSize := 2560*2 + 1440*2 // whole screen

	pix = getEdges(pix, d.Width(), d.Height())
	// TODO: do this outside this package
	pix = getBounds(pix, cfg.BoundsOffset*4, cfg.BoundsSize*4)

	pix = averagePix(pix, cfg.Leds)
	pix = adjustWhitePoint(pix, 16, 256)

	v.events <- interfaces.UpdateEvent{
		Display: d,
		Data:    pix,
	}
}

// adjustWhitePoint adjusts the white point for each color individually.
func adjustWhitePoint(pix []byte, bp, wp float64) []byte {
	for i := 0; i < len(pix); i++ {
		pix[i] = awp(pix[i], bp, wp)
	}

	return pix
}

func awp(color byte, min, max float64) byte {
	c := float64(color)
	res := c/256*(max-min) + min
	return byte(res)
}

func averagePix(src []byte, ledsCount int) []byte {
	pixels := len(src) / 4
	pixelsPerLed := pixels / ledsCount
	dst := make([]byte, ledsCount*4)

	for i := 0; i < ledsCount; i++ {
		// Initialize the color values to zero
		var r, g, b, a = 0, 0, 0, 0
		// Loop all pixels in the current segment
		offset := pixelsPerLed * i * 4
		if i == ledsCount-1 {
			// Grab the remaining n pixels
			// They will be at most len(pix) % count
			pixelsPerLed = pixels - (pixelsPerLed * (ledsCount - 1))
		}
		for j := 0; j < pixelsPerLed*4; j += 4 {
			// Calculate the offset (based on current segment)
			// Add the casted color integer to the last value
			r += int(src[offset+j])
			g += int(src[offset+j+1])
			b += int(src[offset+j+2])
			a += int(src[offset+j+3])
			// r = int(data[offset + j * 3]);
			// g = int(data[offset + j * 3 + 1]);
			// b = int(data[offset + j * 3 + 2]);
			// fmt.Println(offset + j)
		}
		// Get the average by dividing the accumulated color value with the
		// count of the pixels in the segment
		r = r / pixelsPerLed
		g = g / pixelsPerLed
		b = b / pixelsPerLed
		a = a / pixelsPerLed

		// Modify the correct bytes on the LED data
		// Leaving the first byte untouched
		dst[i*4] = uint8(r)
		dst[i*4+1] = uint8(g)
		dst[i*4+2] = uint8(b)
		dst[i*4+3] = uint8(a)
	}

	return dst

}

func getBounds(edg []byte, offset, size int) []byte {
	newBounds := make([]byte, size) // 3 times the size (R G B bytes)
	for i := 0; i < size; i++ {
		newBounds[i] = edg[(i+offset)%len(edg)]
	}
	return newBounds
}

// getEdges decodes the pixel data from the specified image, stores the
// border pixels in four arrays, averages the borders based on the specified
// length of the strip, sets the operation mode to 'A' (Ambilight) and returns
// the color data as a bytes array
func getEdges(pix []byte, width int, height int) []byte {
	// index from stride and coords: y*Stride + x*4
	// Initialize new waitgroup
	var wg sync.WaitGroup
	wg.Add(4)
	// Two horizontal two vertical, 4 bytes for each pixel (RGBA)
	b := make([]byte, (width*2+height*2)*4)
	// Create a wait group and add a goroutine for each edge
	// Top edge
	go func() {
		// Once complete set as done
		defer wg.Done()
		// Offset is 0 for the top edge, we are going clockwise
		// Loop all the pixels
		copy(b[0:width*4], pix[0:width*4])
	}()
	// Right edge
	go func() {
		defer wg.Done()
		// Offset is 4 times the width of the display,
		// since we need 4 bytes per pixel (RGB values)
		offset := width * 4
		for y := 0; y < height*4; y += 4 {
			i := y*width + (width-1)*4
			b[offset+y] = pix[i]
			b[offset+y+1] = pix[i+1]
			b[offset+y+2] = pix[i+2]
			b[offset+y+3] = pix[i+3]
		}
	}()

	// TODO: Bottom edge

	go func() {
		defer wg.Done()
		offset := (width + height) * 4
		for x := 0; x < width*4; x += 4 {
			i := (width * (height - 1) * 4) - x
			b[offset+x] = pix[i]
			b[offset+x+1] = pix[i+1]
			b[offset+x+2] = pix[i+2]
			b[offset+x+3] = pix[i+3]
		}
	}()
	// Left edge
	go func() {
		defer wg.Done()
		offset := (width*2 + height) * 4
		for y := 0; y < height*4; y += 4 {
			i := (height*4 - y - 4) * width
			b[offset+y] = pix[i]
			b[offset+y+1] = pix[i+1]
			b[offset+y+2] = pix[i+2]
			b[offset+y+3] = pix[i+3]
		}
	}()
	// Wait until all routines are complete
	wg.Wait()
	// Return the bounding pixels
	return b
}

func (v *Visualizer) Stop() error {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}

	<-v.done

	return nil
}

func New(opts ...Option) (*Visualizer, error) {
	v := &Visualizer{}

	for _, opt := range opts {
		err := opt(v)
		if err != nil {
			return nil, err
		}
	}

	v.events = make(chan interfaces.UpdateEvent)

	return v, nil
}
