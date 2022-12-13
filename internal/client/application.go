package client

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"sync"
	"time"

	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/audio"
	"ledctl3/internal/client/controller/video"
	"ledctl3/internal/pkg/event"

	"github.com/gorilla/websocket"
)

type Application struct {
	DefaultMode controller.Mode

	Host       string
	Port       int
	Leds       int
	StripType  StripType
	GpioPin    int
	Brightness int
	BlackPoint float64
	Segments   []Segment

	connMux sync.Mutex
	conn    *websocket.Conn

	Displays       video.DisplayRepository
	DisplayConfigs [][]video.DisplayConfig

	Colors     []color.Color
	WindowSize int

	//cfg config.Config
	//ip       string
	//leds     int
	//mode     string
	//capturer string

	ctl           *controller.Controller
	ServerAddress string

	Calibration map[int]Calibration

	//displayVisualizer visualizer.Visualizer
	//audioVisualizer   visualizer.Visualizer
}

type Segment struct {
	Id   int
	Leds int
}

type Calibration struct {
	Red   float64
	Green float64
	Blue  float64
	White float64
}

func New(opts ...Option) (*Application, error) {
	a := &Application{}

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
		video.WithDisplayConfig(a.DisplayConfigs),
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
		audio.WithLedsCount(a.Leds),
		audio.WithSegments(segs),
		audio.WithColors(a.Colors...),
		audio.WithWindowSize(a.WindowSize),
		audio.WithBlackPoint(a.BlackPoint),
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
		for events := range a.ctl.Events() {
			a.connMux.Lock()
			conn := a.conn
			a.connMux.Unlock()

			if conn == nil {
				fmt.Println("no connection")
				continue
			}

			//for _, e := range events {
			//	fmt.Printf("-> %s\n", e)
			//}

			b, err := json.Marshal(events)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				a.connMux.Lock()
				a.conn = nil
				a.connMux.Unlock()
			}
		}
	}()

	// go func() {
	// 	for {
	// 		time.Sleep(1 * time.Second)
	// 		fmt.Printf("\r%s", a.ctl.Statistics().AverageProcessingTime)
	// 	}
	// }()

	return a, nil
}

func (a *Application) Start() error {
	var err error

	go func() {
		for {
			// try to re-establish connection if lost
			defer time.Sleep(1 * time.Second)

			a.connMux.Lock()
			conn := a.conn
			a.connMux.Unlock()

			if conn == nil {
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel()

					conn, _, err = websocket.DefaultDialer.DialContext(ctx, a.ServerAddress, nil)
					if err != nil {
						fmt.Println(err)
						return
					}

					a.connMux.Lock()
					a.conn = conn
					a.connMux.Unlock()

					for {
						typ, b, err := conn.ReadMessage()
						if err != nil {
							fmt.Println("error during read", err)
							a.connMux.Lock()
							a.conn = nil
							a.connMux.Unlock()
							return
						}

						if typ != websocket.TextMessage {
							fmt.Println("invalid message type")
							continue
						}

						events, err := event.Parse(b)
						if err != nil {
							fmt.Println(err)
							continue
						}

						a.ProcessEvents(events...)
					}

					//err = a.reload()
					//if err != nil {
					//	fmt.Println(err)
					//	return
					//}
				}()

			}
		}
	}()

	err = a.ctl.SetMode(a.DefaultMode)
	if err != nil {
		panic(err)
	}

	return nil
}

//func (a *Application) reload() error {
//	segments := []event.SegmentConfig{}
//
//	for _, s := range a.Segments {
//		segments = append(
//			segments, event.SegmentConfig{
//				Id:   s.Id,
//				Leds: s.Leds,
//			},
//		)
//	}
//
//	e := event.NewUpdateEvent(a.Leds, string(a.StripType), a.GpioPin, a.Brightness, segments)
//
//	b, err := json.Marshal(e)
//	if err != nil {
//		return err
//	}
//
//	a.connMux.Lock()
//	conn := a.conn
//	a.connMux.Unlock()
//
//	if conn != nil {
//		err = a.conn.WriteMessage(websocket.TextMessage, b)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func (a *Application) Stop() error {
	a.connMux.Lock()
	conn := a.conn
	a.connMux.Unlock()

	if conn != nil {
		conn.Close()
	}

	return a.ctl.SetMode(controller.Reset)
}

func (a *Application) Handle(t event.Type, b []byte) {
	switch t {
	case event.Connected:

	}
}

func (a *Application) ProcessEvents(events ...event.Event) {
	//for _, e := range events {
	//	fmt.Printf("<- %s\n", e)
	//}
}
