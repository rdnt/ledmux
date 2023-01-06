package video

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"ledctl3/internal/client/visualizer"

	"golang.org/x/image/draw"

	"github.com/lucasb-eyer/go-colorful"
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

	displays []Display
	scalers  map[int]draw.Scaler
}

type DisplayConfig struct {
	Id        int
	Width     int
	Height    int
	Left      int
	Top       int
	Framerate int
	Segments  []Segment
}

type Segment struct {
	Id      int
	Leds    int
	From    Vector2
	To      Vector2
	Reverse bool
}

type Vector2 struct {
	X int
	Y int
}

func (d DisplayConfig) String() string {
	//return fmt.Sprintf(
	//	"DisplayConfig{id: %d, segmentId: %d, leds: %d, width: %d, height: %d, left: %d, top: %d, framerate: %d, offset: %d, size: %d, bounds: %+v}",
	//	d.Id, d.SegmentId, d.Leds, d.Width, d.Height, d.Left, d.Top, d.Framerate, d.BoundsOffset, d.BoundsSize, d.Bounds,
	//)
	return fmt.Sprintf("%#v", d)
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

	v.scalers = make(map[int]draw.Scaler)

	for _, cfg := range displayConfigs {
		for _, seg := range cfg.Segments {
			rect := image.Rect(seg.From.X, seg.From.Y, seg.To.X, seg.To.Y)

			// TODO: only allow cube (Dx == Dy) if segment is only 1 led

			var width, height int

			if rect.Dx() > rect.Dy() {
				// horizontal
				width = seg.Leds
				height = 2
			} else {
				// vertical
				width = 2
				height = seg.Leds
			}

			v.scalers[seg.Id] = draw.BiLinear.NewScaler(width, height, cfg.Width, cfg.Height)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(v.displays))

	for _, d := range v.displays {
		cfg := displayConfigs[d.Id()]

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
				v.stopCapture()
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
	now := time.Now()
	if len(cfg.Segments) == 0 {
		return
	}

	src := &image.RGBA{
		Pix:    pix,
		Stride: d.Width() * 4,
		Rect:   image.Rect(0, 0, d.Width(), d.Height()),
	}

	segs := make([]visualizer.Segment, len(cfg.Segments))

	var wg sync.WaitGroup
	wg.Add(len(cfg.Segments))

	for i, seg := range cfg.Segments {
		go func(i int, seg Segment) {
			defer wg.Done()

			rect := image.Rect(seg.From.X, seg.From.Y, seg.To.X, seg.To.Y)

			sub := src.SubImage(rect)

			var dst *image.RGBA

			if rect.Dx() > rect.Dy() {
				// horizontal
				dst = image.NewRGBA(image.Rect(0, 0, seg.Leds, 1))
			} else {
				// vertical
				dst = image.NewRGBA(image.Rect(0, 0, 1, seg.Leds))
			}

			v.scalers[seg.Id].Scale(dst, dst.Bounds(), sub, sub.Bounds(), draw.Over, nil)

			colors := []color.Color{}

			for i := 0; i < len(dst.Pix); i += 4 {
				clr, _ := colorful.MakeColor(color.NRGBA{
					R: dst.Pix[i],
					G: dst.Pix[i+1],
					B: dst.Pix[i+2],
					A: dst.Pix[i+3],
				})

				colors = append(colors, clr)
			}

			if seg.Reverse {
				reverse(colors)
			}

			segs[i] = visualizer.Segment{
				Id:  seg.Id,
				Pix: colors,
			}
		}(i, seg)

	}

	wg.Wait()

	v.events <- visualizer.UpdateEvent{
		Segments: segs,
		Latency:  time.Since(now),
	}
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
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
	v := &Visualizer{}

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
