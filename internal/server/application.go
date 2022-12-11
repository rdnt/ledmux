package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"ledctl3/internal/pkg/event"
	"ledctl3/internal/server/config"
	"ledctl3/pkg/ws281x"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
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

	leds       int
	stripType  string
	gpioPin    int
	brightness int
	segments   []Segment
}

type Segment struct {
	id    int
	start int
	end   int
	leds  int
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

//func (ctl *Application) Handle(e events.EventWithType) {
//	switch e.EventWithType {
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

func (a *Application) Handle(typ event.Type, b []byte) {
	switch typ {
	case event.Update:
		a.HandleUpdateEvent(b)
	case event.SetLeds:
		a.HandleSetLedsEvent(b)
	default:
		fmt.Println("unknown event")
	}
}

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

				evts, err := ParseMessage(b)
				if err != nil {
					fmt.Println("invalid message format")
					continue
				}

				a.ProcessEvents(evts)
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

func (a *Application) HandleUpdateEvent(b []byte) {
	var evt event.UpdateEvent

	err := json.Unmarshal(b, &evt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg := evt.Segments[0]

	err = a.reload(cfg.GpioPin, cfg.Leds, cfg.Brightness, cfg.StripType)
	if err != nil {
		fmt.Println(err)
	}
}

func (a *Application) HandleSetLedsEvent(b []byte) {
	if a.ws == nil {
		return
	}

	var evt event.SetLedsEvent
	err := json.Unmarshal(b, &evt)
	if err != nil {
		panic(err)
	}

	if a.mode != Render {
		a.ws.Stop()
		a.mode = Render
	}

	//out := "\n"
	for _, seg := range evt.Segments {
		idx := slices.IndexFunc(a.segments, func(s Segment) bool {
			return s.id == seg.Id
		})

		if idx == -1 {
			fmt.Println("Segment doesn't exist:", seg.Id)
			return
		}

		segment := a.segments[idx]

		for i := 0; i < (segment.end-segment.start)*4; i += 4 {
			// Parse color data for current LED
			r := seg.Pix[i]
			g := seg.Pix[i+1]
			b := seg.Pix[i+2]
			aa := seg.Pix[i+3]
			// Set the current LED's color
			// Not need to check for error
			err := a.ws.SetLedColor(i/4+segment.start, r, g, b, aa)
			if err != nil {
				fmt.Println(err)
			}

			//out += color.RGB(seg.Pix[i], seg.Pix[i+1], seg.Pix[i+2], true).Sprintf(" ")
		}

	}
	//fmt.Print(out)

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

		err = a.ws.Render()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// TODO: properly handle out of bounds
	//if a.leds != len(evt.Data) / 4 {
	//	a.leds = len(evt.Data) / 4
	//}

	//if a.rendering {
	//	return
	//}
	//
	//a.rendering = true

	//segment = a.segments[1]
	//
	//for i := 0; i < (segment.end-segment.start)*4; i += 4 {
	//	// Parse color data for current LED
	//	r := evt.Data[i]
	//	g := evt.Data[i+1]
	//	b := evt.Data[i+2]
	//	// Set the current LED's color
	//	// Not need to check for error
	//	err := a.ws.SetLedColor(i/4+segment.start, r, g, b)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}

	//a.rendering = false
	//}

	//fmt.Print(evt.SegmentId)
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

func (a *Application) ProcessEvents(evts []event.Event) {
	//for _, e := range evts {
	//	switch e.EventWithType {
	//	case event.SetLeds:
	//		a.HandleSetLedsEvent(e)
	//	}
	//}
}
