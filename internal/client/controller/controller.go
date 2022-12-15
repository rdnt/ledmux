package controller

import (
	"fmt"
	"sync"
	"time"

	"ledctl3/internal/client/visualizer"
	"ledctl3/internal/pkg/event"

	"github.com/VividCortex/ewma"
)

type Controller struct {
	leds       int
	Mode       Mode
	visualizer visualizer.Visualizer
	events     chan []event.Event

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

	s.events = make(chan []event.Event)

	return s, nil
}

func (ctl *Controller) Start() error {
	return nil
}

func (ctl *Controller) Events() chan []event.Event {
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
				ctl.timing.process.Add(float64(evt.Latency.Nanoseconds()))
				ctl.timingMux.Unlock()

				events := []event.Event{}

				for _, seg := range evt.Segments {
					pix := make([]uint8, 0, len(seg.Pix)*4)
					for _, c := range seg.Pix {
						r, g, b, a := c.RGBA()
						pix = append(pix, uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
					}

					events = append(events, event.SetLedsEvent{
						Event: event.SetLeds,
						Id:    seg.Id,
						Pix:   pix,
					})
				}

				ctl.events <- events
			}
		}()
	}

	return nil
}
