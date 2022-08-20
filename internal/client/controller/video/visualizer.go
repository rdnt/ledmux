package video

import (
	"context"
	"errors"
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/bamiaux/rez"

	"ledctl3/internal/client/config"
	"ledctl3/internal/client/visualizer"
)

var (
	ErrConfigNotFound = fmt.Errorf("config not found")
)

type Visualizer struct {
	displayRepo    DisplayRepository
	leds           int
	cancel         context.CancelFunc
	done           chan bool
	events         chan visualizer.UpdateEvent
	displayConfigs [][]DisplayConfig

	displays        []Display
	displayResizers []rez.Converter
}

type DisplayConfig struct {
	Id           int
	SegmentId    int
	Leds         int
	Width        int
	Height       int
	Left         int
	Top          int
	Framerate    int
	BoundsOffset int
	BoundsSize   int
	Bounds       config.Bounds
}

type Bounds struct {
	From Vector2 `yaml:"from" json:"from"`
	To   Vector2 `yaml:"to" json:"to"`
}

type Vector2 struct {
	X int `yaml:"x" json:"x"`
	Y int `yaml:"y" json:"y"`
}

func (d DisplayConfig) String() string {
	return fmt.Sprintf(
		"DisplayConfig{id: %d, segmentId: %d, leds: %d, width: %d, height: %d, left: %d, top: %d, framerate: %d, offset: %d, size: %d, bounds: %+v}",
		d.Id, d.SegmentId, d.Leds, d.Width, d.Height, d.Left, d.Top, d.Framerate, d.BoundsOffset, d.BoundsSize, d.Bounds,
	)
}

func (v *Visualizer) Events() chan visualizer.UpdateEvent {
	return v.events
}

func (v *Visualizer) startCapture(ctx context.Context) error {
	captureCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	v.displays, err = v.displayRepo.All()
	if err != nil {
		return err
	}

	displayConfigs, err := v.matchDisplays(v.displays)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(v.displays))

	for _, d := range v.displays {
		cfg := displayConfigs[d.Id()]

		fmt.Println("########################################")
		fmt.Println(d.Id())
		fmt.Println(d)
		fmt.Println(cfg)
		fmt.Println("########################################")

		go func(d Display) {
			defer wg.Done()
			frames := d.Capture(captureCtx, cfg.Framerate)

			for frame := range frames {
				//fmt.Println(d.Resolution())

				go v.process(d, cfg, frame)
			}

			cancel()
		}(d)
	}

	wg.Wait()
	time.Sleep(3 * time.Second)

	return nil
}

func (v *Visualizer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	v.done = make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("parent ctx done, exiting")
				v.done <- true
				return
			default:
				fmt.Println("STARTING CAPTURE")

				err := v.startCapture(ctx)
				if errors.Is(err, context.Canceled) {
					fmt.Println("capture canceled")

					v.stopCapture()
					return
				} else if err != nil {
					fmt.Println("error starting capture:", err)

					v.stopCapture()
					time.Sleep(3 * time.Second)
				}
			}
		}
	}()

	return nil
}

func (v *Visualizer) stopCapture() {
	for _, d := range v.displays {
		err := d.Close()
		if err != nil {
			fmt.Println(err)
		}
	}

	v.displays = nil
}

func (v *Visualizer) process(d Display, cfg DisplayConfig, pix []byte) {
	//fmt.Println("process:", d.Id())
	//fmt.Println(d)
	//fmt.Println(cfg)

	if v.displayResizers[d.Id()] == nil {
		src := &image.NRGBA{
			Pix:    pix,
			Stride: d.Width() * 4,
			Rect:   image.Rect(0, 0, d.Width(), d.Height()),
		}

		ratio := (d.Width() + d.Height()) / (cfg.Leds / 2)
		ratio = 90

		//fmt.Println("ratio", ratio)
		//fmt.Println(d.Width()/ratio, d.Height()/ratio)

		dst := image.NewNRGBA(image.Rect(0, 0, d.Width()/ratio, d.Height()/ratio))

		//fmt.Println()

		convertCfg, err := rez.PrepareConversion(dst, src)
		if err != nil {
			panic(err)
		}

		converter, err := rez.NewConverter(convertCfg, rez.NewBilinearFilter())
		if err != nil {
			panic(err)
		}

		v.displayResizers[d.Id()] = converter
	}

	src := &image.NRGBA{
		Pix:    pix,
		Stride: d.Width() * 4,
		Rect:   image.Rect(0, 0, d.Width(), d.Height()),
	}

	ratio := (d.Width() + d.Height()) / (cfg.Leds / 2)
	ratio = 90

	dst := image.NewNRGBA(image.Rect(0, 0, d.Width()/ratio, d.Height()/ratio))

	err := v.displayResizers[d.Id()].Convert(dst, src)
	if err != nil {
		panic(err)
	}

	pix = dst.Pix
	//width, height := d.Width(), d.Height()
	width, height := d.Width()/ratio, d.Height()/ratio

	//total := (d.Width()/ratio + d.Height()/ratio) * 2

	if d.Orientation() == Portrait || d.Orientation() == PortraitFlipped {
		width, height = height, width
	}

	pix = getEdges(pix, width, height)
	//pix = rotatePix(pix, d.Orientation())

	fromOffset := calculateOffset(width, height, 8, 21)
	toOffset := calculateOffset(width, height, 31, 21)

	size := getPixSliceSize(width, height, fromOffset, toOffset)

	pix = getBounds(pix, fromOffset*4, size*4)
	pix = averagePix(pix, cfg.Leds)
	pix = adjustWhitePoint(pix, 0, 256)

	v.events <- visualizer.UpdateEvent{
		Segments: []visualizer.Segment{
			{
				Id:  cfg.SegmentId,
				Pix: pix,
			},
		},
	}
}

func getPixSliceSize(width, height, from, to int) int {
	return (width*2 + height*2) - from + to
}

func calculateOffset(width, height, x, y int) int {
	var offset int
	if x == 0 {
		offset = 2*width + height + (height - y)
	} else if x == width-1 {
		offset = width + y
	} else {
		if y == 0 {
			offset = x
		} else if y == height-1 {
			offset = width + height + (width - x)
		} else {
			return 0
		}
	}
	// offset = offset % (d.Width*2 + d.Height*2)
	return offset
}

//func rotatePix(pix []byte, orientation Orientation) []byte {
//
//}

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

// getBounds filters the given edge
func getBounds(edgePix []byte, offset, size int) []byte {
	newBounds := make([]byte, size)

	for i := 0; i < size; i++ {
		newBounds[i] = edgePix[(i+offset)%len(edgePix)]
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

	fmt.Println("stop: waiting for done")

	<-v.done

	fmt.Println("stop: done received")
	return nil
}

// matchDisplays tries to map a display entry from the system to one in the
// config file. It should be re-run whenever the config file changes or a
// display capturer becomes invalid, for example if an app enters fullscreen or
// when a display is (dis)connected.
func (v *Visualizer) matchDisplays(displays []Display) (map[int]DisplayConfig, error) {
	// try to find matching configuration
	var match map[int]DisplayConfig

	fmt.Println("displays", displays)
	fmt.Println("configs", v.displayConfigs)

	for _, cfg := range v.displayConfigs {
		sys2cfg := map[int]int{}
		cfg2sys := map[int]int{}

		for _, displayCfg := range cfg {
			for _, sysd := range displays {
				_, ok1 := cfg2sys[displayCfg.Id]
				_, ok2 := sys2cfg[sysd.Id()]

				if ok1 || ok2 {
					// this display has already been matched with a config entry
					continue
				}

				widthEq := sysd.Width() == displayCfg.Width
				heightEq := sysd.Height() == displayCfg.Height
				leftEq := sysd.X() == displayCfg.Left
				topEq := sysd.Y() == displayCfg.Top

				if widthEq && heightEq && leftEq && topEq {
					// resolution and offset is the same, match found!
					cfg2sys[displayCfg.Id] = sysd.Id()
					sys2cfg[sysd.Id()] = displayCfg.Id

					break
				}
			}
		}

		if len(sys2cfg) != len(displays) {
			// not all displays have been matched, try another config
			continue
		}

		match = map[int]DisplayConfig{}
		for displayId, configId := range sys2cfg {
			match[displayId] = cfg[configId]
		}

		break
	}

	if match == nil {
		return nil, ErrConfigNotFound
	}

	fmt.Println("match", match)

	return match, nil
}

func New(opts ...Option) (*Visualizer, error) {
	v := &Visualizer{
		displayResizers: make([]rez.Converter, 2),
	}

	for _, opt := range opts {
		err := opt(v)
		if err != nil {
			return nil, err
		}
	}

	if v.displayRepo == nil {
		return nil, fmt.Errorf("invalid display repository")
	}

	v.events = make(chan visualizer.UpdateEvent)

	return v, nil
}
