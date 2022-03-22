package client

import (
	"fmt"
	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/ambilight"
	"ledctl3/internal/client/interfaces"
	"ledctl3/pkg/udp"
)

type App struct {
	DefaultMode controller.Mode

	Host       string
	Port       int
	Leds       int
	StripType  StripType
	GpioPin    int
	Brightness int
	BlackPoint int
	Segments   []Segment
	conn       udp.Client

	Displays       interfaces.DisplayRepository
	DisplayConfigs [][]ambilight.DisplayConfig

	//cfg config.Config
	//ip       string
	//leds     int
	//mode     string
	//capturer string

	ctl *controller.Controller
	//displayVisualizer interfaces.Visualizer
	//audioVisualizer   interfaces.Visualizer
}

type Server struct {
}

type Segment struct {
	Leds int
}

func New(opts ...Option) (*App, error) {
	a := &App{}

	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return nil, err
		}
	}

	var err error
	a.conn, err = udp.NewClient(fmt.Sprintf("%s:%d", a.Host, a.Port))
	if err != nil {
		return nil, err
	}

	displayVisualizer, err := ambilight.New(
		ambilight.WithLedsCount(a.Leds),
		ambilight.WithDisplayRepository(a.Displays),
		ambilight.WithDisplayConfig(a.DisplayConfigs), // TODO @@@@
	)
	if err != nil {
		return nil, err
	}

	a.ctl, err = controller.New(
		controller.WithLedsCount(a.Leds),
		controller.WithDisplayVisualizer(displayVisualizer),
		controller.WithAudioVisualizer(nil),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for b := range a.ctl.Events() {
			err = a.conn.Send(b)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	return a, nil
}

func (a *App) Start() error {
	err := a.reload()
	if err != nil {
		panic(err)
	}

	return a.ctl.SetMode(a.DefaultMode)
}

func (a *App) reload() error {
	//segments := []events.Segment{}

	// TODO: pass proper matched config
	//for _, d := range a.DisplayConfigs[0] {
	//	segments = append(
	//		segments, events.Segment{
	//			Id:   d.Id,
	//			Leds: d.Leds,
	//		},
	//	)
	//}

	//e := events.NewReloadEvent(a.Leds, string(a.StripType), a.GpioPin, a.Brightness, segments)
	//
	//b, err := msgpack.Marshal(e)
	//if err != nil {
	//	return err
	//}
	//
	//err = a.conn.Send(b)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (a *App) Stop() error {
	return a.ctl.SetMode(controller.Reset)
}
