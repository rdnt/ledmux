package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"ledctl3/internal/pkg/event"
	"ledctl3/internal/server/config"
	"ledctl3/pkg/color"
	"ledctl3/pkg/ws281x"

	"github.com/gorilla/websocket"
)

type Mode string

const (
	Idle   Mode = "idle"
	Render Mode = "render"
)

//
//const (
//	Ambilight Mode = "video"
//	AudioViz  Mode = "audio"
//	Rainbow   Mode = "rainbow"
//	Static    Mode = "static"
//	Reload    Mode = "reload"
//)
//
//var modes = map[string]Mode{
//	"video":   Ambilight,
//	"audio":   AudioViz,
//	"rainbow": Rainbow,
//	"static":  Static,
//	"reload":  Reload,
//}

type Application struct {
	mux       sync.Mutex
	events    chan []byte
	ws        *ws281x.Engine
	mode      Mode
	rendering bool
	buffer    []byte

	leds        int
	stripType   string
	gpioPin     int
	brightness  int
	segments    map[int]Segment
	calibration map[int]Calibration
}

type Segment struct {
	id    int
	start int
	end   int
	leds  int
}

type Calibration struct {
	Start int
	End   int
	Red   float64
	Green float64
	Blue  float64
	White float64
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65535, //1024
	WriteBufferSize: 65535,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func New(c config.Config) (*Application, error) {
	err := validateConfig(c)
	if err != nil {
		return nil, err
	}

	a := &Application{
		events: make(chan []byte, 1),
		ws:     nil,
	}

	err = a.applyConfig(c)
	if err != nil {
		return nil, err
	}

	err = a.reload(a.gpioPin, a.leds, a.brightness, a.stripType)
	if err != nil {
		fmt.Println(err)
	}

	return a, nil
}

//func (ctl *Application) Handle(e events.EventType) {
//	switch e.EventType {
//	case events.Ambilight:
//		if ctl.mode != Ambilight {
//			ctl.ws.Stop()
//			ctl.mode = Ambilight
//			fmt.Println("amb")
//		}
//		evt := events.AmbilightEvent(e)
//		if ctl.leds != len(evt.Data) / 4 {
//			ctl.leds = len(evt.Data) / 4
//		}
//		for i := 0; i < ctl.leds; i++ {
//			// Parse color data for current LED
//			r = uint8(data[i*3])
//			g = uint8(data[i*3+1])
//			b = uint8(data[i*3+2])
//			// Set the current LED's color
//			// Not need to check for error
//			_ = ws.SetLedColor(i, r, g, b)
//		}
//
//		ctl.ws.SetLedColor()
//	case events.Reload:
//		err := ctl.reload(
//			//data.LedsCount, data.GPIOPin, data.Brightness,
//			200, 18, 100,
//		)
//		if err != nil {
//			panic(err)
//		}
//	}
//}

func (a *Application) Start() error {
	http.HandleFunc(
		"/ws", func(w http.ResponseWriter, req *http.Request) {
			wsconn, err := upgrader.Upgrade(w, req, nil)
			if err != nil {
				fmt.Println(err)
				return
			}

			a.HandleConnected(wsconn)

			wsconn.EnableWriteCompression(true)

			for {
				typ, b, err := wsconn.ReadMessage()
				if err != nil {
					fmt.Println("error during read", err)
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
		},
	)

	go http.ListenAndServe(":4197", nil)

	return nil
}

func (a *Application) reload(gpioPin, ledsCount, brightness int, stripType string) error {
	if a.ws != nil {
		err := a.ws.Clear()
		if err != nil {
			return err
		}

		a.ws.Fini()
	}

	engine, err := ws281x.Init(gpioPin, ledsCount, brightness, stripType)
	if err != nil {
		return err
	}

	a.ws = engine

	//i := 0
	//for _, s := range segments {
	//	a.segments = append(
	//		a.segments, Segment{
	//			id:    s.Id,
	//			start: i,
	//			end:   i + s.Leds,
	//		},
	//	)
	//
	//	i += s.Leds
	//}

	return nil
}

func (a *Application) HandleUpdateEvent(e event.UpdateEvent) {
	leds := 0
	for _, seg := range e.Segments {
		leds += seg.Leds
	}

	err := a.reload(e.GpioPin, leds, e.Brightness, e.StripType)
	if err != nil {
		fmt.Println(err)
	}
}

func (a *Application) HandleSetLedsEvent(e event.SetLedsEvent) {
	seg, ok := a.segments[e.SegmentId]
	if !ok {
		fmt.Println("Segment doesn't exist:", e.SegmentId)
		return
	}

	for i := seg.start; i < seg.end; i++ {
		// Parse color data for current LED
		offset := i * 4

		r := e.Pix[offset]
		g := e.Pix[offset+1]
		b := e.Pix[offset+2]
		aa := e.Pix[offset+3]

		// Set the current LED's color
		// Not need to check for error
		err := a.setLedColor(i, r, g, b, aa)
		if err != nil {
			fmt.Println(err)
		}

		//out += color.RGB(seg.Pix[i], seg.Pix[i+1], seg.Pix[i+2], true).Sprintf(" ")
	}
}

func (a *Application) HandleSetColorEvent(e event.SetColorEvent) {
	seg, ok := a.segments[e.SegmentId]
	if !ok {
		fmt.Println("Segment doesn't exist:", e.SegmentId)
		return
	}

	clr, err := color.FromString(e.Color)
	if err != nil {
		fmt.Println(err)
		return
	}

	r, g, b, aa := clr.RGBA()

	for i := seg.start; i < seg.end; i++ {
		err := a.setLedColor(i, uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(aa>>8))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (a *Application) HandleTurnOffEvent(e event.TurnOffEvent) {
	seg, ok := a.segments[e.SegmentId]
	if !ok {
		fmt.Println("Segment doesn't exist:", e.SegmentId)
		return
	}

	for i := seg.start; i < seg.end; i++ {
		err := a.setLedColor(i, 0, 0, 0, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func (a *Application) HandleConnected(wsconn *websocket.Conn) {
	segs := make([]event.ConnectedEventSegment, len(a.segments))
	for i, seg := range a.segments {
		segs[i] = event.ConnectedEventSegment{
			Id:   seg.id,
			Leds: seg.leds,
		}
	}

	e := event.ConnectedEvent{
		Event:      event.Connected,
		Brightness: a.brightness,
		GpioPin:    a.gpioPin,
		StripType:  a.stripType,
		Segments:   segs,
	}

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	err = wsconn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		panic(err)
	}
}

func (a *Application) setLedColor(id int, r, g, b, aa uint8) error {
	calib, ok := a.calibration[id]
	if ok {
		r = uint8(float64(r) * calib.Red)
		g = uint8(float64(g) * calib.Green)
		b = uint8(float64(b) * calib.Blue)
		aa = uint8(float64(aa) * calib.White)
	}

	return a.ws.SetLedColor(id, r, g, b, aa)
}

func (a *Application) ProcessEvents(events ...event.Event) {
	for _, e := range events {
		//fmt.Printf("<- %s\n", e)

		switch e := e.(type) {
		case event.SetColorEvent:
			a.HandleSetColorEvent(e)
		case event.SetEffectEvent:
			fmt.Println("setEffect event: no handler")
		case event.SetLedsEvent:
			a.HandleSetLedsEvent(e)
		case event.TurnOffEvent:
			a.HandleTurnOffEvent(e)
		case event.TurnOnEvent:
			fmt.Println("turnOn event: no handler")
		case event.UpdateEvent:
			a.HandleUpdateEvent(e)
		default:
			fmt.Println("unknown event", e)
		}
	}

	if a.ws == nil {
		return
	}

	a.mux.Lock()
	if a.rendering {
		a.mux.Unlock()

		return
	}

	a.rendering = true
	a.mux.Unlock()

	go func() {
		defer func() {
			a.mux.Lock()
			a.rendering = false
			a.mux.Unlock()
		}()

		err := a.ws.Render()
		if err != nil {
			fmt.Println(err)
		}
	}()
}
