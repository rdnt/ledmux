package client

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/ambilight"
	"ledctl3/internal/client/interfaces"
	"ledctl3/pkg/events"
	"ledctl3/pkg/udp"
)

type App struct {
	DefaultMode  controller.Mode
	BlackPoint int

	Host       string
	Port       int
	Leds       int
	StripType  StripType
	GpioPin    int
	Brightness int

	Displays []ambilight.DisplayConfig

	//cfg config.Config
	//ip       string
	//leds     int
	//mode     string
	//capturer string

	ctl               *controller.Controller
	conn              udp.Client
	displayRepository interfaces.DisplayRepository
	displayVisualizer interfaces.Visualizer
	audioVisualizer   interfaces.Visualizer
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

	a.displayVisualizer, err = ambilight.New(
		ambilight.WithLedsCount(a.Leds),
		ambilight.WithDisplayRepository(a.displayRepository),
		ambilight.WithDisplayConfig(a.Displays),
	)
	if err != nil {
		return nil, err
	}

	a.ctl, err = controller.New(
		controller.WithLedsCount(a.Leds),
		controller.WithDisplayVisualizer(a.displayVisualizer),
		controller.WithAudioVisualizer(a.audioVisualizer),
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
	e := events.NewReloadEvent(a.Leds, string(a.StripType), a.GpioPin, a.Brightness)

	b, err := msgpack.Marshal(e)
	if err != nil {
		panic(err)
	}

	err = a.conn.Send(b)
	if err != nil {
		fmt.Println(err)
	}

	return a.ctl.SetMode(a.DefaultMode)
}

func (a *App) Stop() error {
	return a.ctl.SetMode(controller.Reset)
}
