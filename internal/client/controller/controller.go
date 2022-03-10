package controller

import (
	"github.com/vmihailenco/msgpack/v5"
	"ledctl3/internal/client/interfaces"
	"ledctl3/internal/pkg/events"
)

type Controller struct {
	leds       int
	mode       Mode
	visualizer interfaces.Visualizer
	events     chan []byte

	displayVisualizer interfaces.Visualizer
	audioVisualizer   interfaces.Visualizer
}

func New(opts ...Option) (*Controller, error) {
	s := &Controller{}

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}

	s.events = make(chan []byte)

	return s, nil
}

func (s *Controller) Start() error {
	return nil
}

func (s *Controller) Events() chan []byte {
	return s.events
}

func (s *Controller) Stop() error {
	if s.visualizer != nil {
		return s.visualizer.Stop()
	}

	return nil
}

type Mode string

const (
	Reset     Mode = "reset"
	Reload    Mode = "reload"
	Ambilight Mode = "ambilight"
	AudioViz  Mode = "audioviz"
	Rainbow   Mode = "rainbow"
	Static    Mode = "static"
)

var Modes = map[string]Mode{
	"reset":     Reset,
	"reload":    Reload,
	"ambilight": Ambilight,
	"audioviz":  AudioViz,
	"rainbow":   Rainbow,
	"static":    Static,
}

func (s *Controller) SetMode(mode Mode) error {
	if mode == s.mode {
		return nil
	}

	s.mode = mode

	if s.visualizer != nil {
		err := s.visualizer.Stop()
		if err != nil {
			return err
		}
	}

	switch mode {
	case Ambilight:
		s.visualizer = s.displayVisualizer
	case AudioViz:
		s.visualizer = s.audioVisualizer
	case Rainbow, Static:
		s.visualizer = nil
	}

	if s.visualizer != nil {
		err := s.visualizer.Start()
		if err != nil {
			return err
		}

		go func() {
			for evt := range s.visualizer.Events() {
				//if evt.Display.Id() == 1 {
				//	continue
				//}
				// TODO: associate display with strip here
				e := events.NewAmbilightEvent(evt.Display.Id(), evt.Data)

				b, err := msgpack.Marshal(e)
				if err != nil {
					panic(err)
				}

				s.events <- b
			}
		}()
	}

	return nil
}
