package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/vmihailenco/msgpack/v5"

	"ledctl3/internal/client/visualizer"
	"ledctl3/internal/pkg/events"
)

type Controller struct {
	leds       int
	Mode       Mode
	visualizer visualizer.Visualizer
	events     chan []byte

	displayVisualizer visualizer.Visualizer
	audioVisualizer   visualizer.Visualizer
	segmentCount      int

	timingMux sync.Mutex
	timing    timing
}

type timing struct {
	process ewma.MovingAverage
}

type Statistics struct {
	AverageProcessingTime time.Duration
}

func (ctl *Controller) Statistics() Statistics {
	ctl.timingMux.Lock()
	defer ctl.timingMux.Unlock()

	return Statistics{
		AverageProcessingTime: time.Duration(ctl.timing.process.Value()),
	}
}

func New(opts ...Option) (*Controller, error) {
	s := &Controller{
		timing: timing{
			process: ewma.NewMovingAverage(100),
		},
	}

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}

	s.events = make(chan []byte, s.segmentCount)

	return s, nil
}

func (ctl *Controller) Start() error {
	return nil
}

func (ctl *Controller) Events() chan []byte {
	return ctl.events
}

func (ctl *Controller) Stop() error {
	if ctl.visualizer != nil {
		return ctl.visualizer.Stop()
	}

	return nil
}

type Mode string

const (
	Reset   Mode = "reset"
	Reload  Mode = "reload"
	Video   Mode = "video"
	Audio   Mode = "audio"
	Rainbow Mode = "rainbow"
	Static  Mode = "static"
)

var Modes = map[string]Mode{
	"reset":   Reset,
	"reload":  Reload,
	"video":   Video,
	"audio":   Audio,
	"rainbow": Rainbow,
	"static":  Static,
}

func (ctl *Controller) SetMode(mode Mode) error {
	if mode == ctl.Mode {
		return fmt.Errorf("invalid Mode")
	}

	ctl.Mode = mode

	if ctl.visualizer != nil {
		err := ctl.visualizer.Stop()
		if err != nil {
			return err
		}
	}

	switch mode {
	case Video:
		ctl.visualizer = ctl.displayVisualizer
	case Audio:
		ctl.visualizer = ctl.audioVisualizer
	case Rainbow, Static:
		ctl.visualizer = nil
	}

	if ctl.visualizer != nil {
		err := ctl.visualizer.Start()
		if err != nil {
			return err
		}

		go func() {
			for evt := range ctl.visualizer.Events() {
				ctl.timingMux.Lock()
				ctl.timing.process.Add(float64(evt.Duration.Nanoseconds()))
				ctl.timingMux.Unlock()

				segs := []events.Segment{}

				for _, seg := range evt.Segments {
					segs = append(segs, events.Segment{
						Id:  seg.Id,
						Pix: seg.Pix,
					})
				}

				e := events.NewAmbilightEvent(segs)

				b, err := msgpack.Marshal(e)
				if err != nil {
					panic(err)
				}

				ctl.events <- b
			}
		}()
	}

	return nil
}
