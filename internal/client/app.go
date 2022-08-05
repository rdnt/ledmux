package client

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"

	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/audio"
	"ledctl3/internal/client/controller/video"
	"ledctl3/internal/pkg/events"
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
	conn       *websocket.Conn

	Displays       video.DisplayRepository
	DisplayConfigs [][]video.DisplayConfig

	//cfg config.Config
	//ip       string
	//leds     int
	//mode     string
	//capturer string

	ctl           *controller.Controller
	ServerAddress string

	//displayVisualizer visualizer.Visualizer
	//audioVisualizer   visualizer.Visualizer
}

type Segment struct {
	Id   int
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
	addr := fmt.Sprintf("ws://%s:%d/ws", a.Host, a.Port)

	a.ServerAddress = addr

	displayVisualizer, err := video.New(
		video.WithLedsCount(a.Leds),
		video.WithDisplayRepository(a.Displays),
		video.WithDisplayConfig(a.DisplayConfigs), // TODO @@@@
	)
	if err != nil {
		return nil, err
	}

	segs := []audio.Segment{}
	for _, seg := range a.Segments {
		segs = append(segs, audio.Segment{
			Id:   seg.Id,
			Leds: seg.Leds,
		})
	}

	audioVisualizer, err := audio.New(
		audio.Options{
			Leds:     a.Leds,
			Segments: segs,
		},
	)

	a.ctl, err = controller.New(
		controller.WithLedsCount(a.Leds),
		controller.WithDisplayVisualizer(displayVisualizer),
		controller.WithAudioVisualizer(audioVisualizer),

		controller.WithSegmentsCount(len(segs)),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for b := range a.ctl.Events() {
			err := a.conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	return a, nil
}

func (a *App) Start() error {
	var err error
	a.conn, _, err = websocket.DefaultDialer.Dial(a.ServerAddress, nil)
	if err != nil {
		return err
	}

	err = a.reload()
	if err != nil {
		panic(err)
	}

	return a.ctl.SetMode(a.DefaultMode)
}

func (a *App) reload() error {
	time.Sleep(1 * time.Second)
	segments := []events.SegmentConfig{}

	for _, s := range a.Segments {
		segments = append(
			segments, events.SegmentConfig{
				Id:   s.Id,
				Leds: s.Leds,
			},
		)
	}

	e := events.NewReloadEvent(a.Leds, string(a.StripType), a.GpioPin, a.Brightness, segments)

	b, err := msgpack.Marshal(e)
	if err != nil {
		return err
	}

	err = a.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() error {
	a.conn.Close()

	return a.ctl.SetMode(controller.Reset)
}
