package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gookit/color"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"

	"ledctl3/internal/pkg/events"
	"ledctl3/pkg/ws281x"
)

// TODO: merge with client mode

type Mode string

const (
	Ambilight Mode = "ambilight"
	AudioViz  Mode = "audioviz"
	Rainbow   Mode = "rainbow"
	Static    Mode = "static"
	Reload    Mode = "reload"
)

var modes = map[string]Mode{
	"ambilight": Ambilight,
	"audioviz":  AudioViz,
	"rainbow":   Rainbow,
	"static":    Static,
	"reload":    Reload,
}

type Controller struct {
	mux       sync.Mutex
	msgs      chan []byte
	ws        *ws281x.Engine
	leds      int
	segments  []Segment
	mode      Mode
	rendering bool
	buffer    []byte
}

type Segment struct {
	id    int
	start int
	end   int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65535, //1024
	WriteBufferSize: 65535,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func New() (*Controller, error) {
	msgs := make(chan []byte, 1)

	http.HandleFunc(
		"/ws", func(w http.ResponseWriter, req *http.Request) {
			wsconn, err := upgrader.Upgrade(w, req, nil)
			if err != nil {
				fmt.Println(err)
				return
			}

			wsconn.EnableWriteCompression(true)

			for {
				typ, b, err := wsconn.ReadMessage()
				if err != nil {
					fmt.Println("err")
					return
				}

				if typ != websocket.BinaryMessage {
					fmt.Println("skip")
					continue
				}

				msgs <- b
			}
		},
	)

	go http.ListenAndServe(":4197", nil)

	ctl := &Controller{
		msgs: msgs,
		ws:   nil,
	}

	return ctl, nil
}

//func (ctl *Controller) Handle(e events.Event) {
//	switch e.Type {
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

func (ctl *Controller) Route(e events.Event, b []byte) {
	switch e.Type {
	case events.Ambilight:
		ctl.HandleAmbilightEvent(b)
	case events.Reload:
		ctl.HandleReloadEvent(b)
	default:
		fmt.Println("unknown event")
	}
}

func (ctl *Controller) Start() error {
	//ws, err := ws281x.Init(18, 83, 255, "GRB")
	//if err != nil {
	//	return err
	//}

	//ctl.ws = ws

	go func() {
		for b := range ctl.msgs {
			var e events.Event

			err := msgpack.Unmarshal(b, &e)
			if err != nil {
				fmt.Println("m")
				//fmt.Println(err)
				continue
			}

			ctl.Route(e, b)
		}
	}()
	return nil
}

func (ctl *Controller) reload(gpioPin, ledsCount, brightness int, stripType string, segments []events.SegmentConfig) error {
	if ctl.ws != nil {
		err := ctl.ws.Clear()
		if err != nil {
			return err
		}

		ctl.ws.Fini()
	}

	engine, err := ws281x.Init(gpioPin, ledsCount, brightness, stripType)
	if err != nil {
		return err
	}

	ctl.ws = engine
	ctl.segments = []Segment{}

	i := 0
	for _, s := range segments {
		ctl.segments = append(
			ctl.segments, Segment{
				id:    s.Id,
				start: i,
				end:   i + s.Leds,
			},
		)

		i += s.Leds
	}

	return nil
}

func (ctl *Controller) HandleReloadEvent(b []byte) {
	var evt events.ReloadEvent

	err := msgpack.Unmarshal(b, &evt)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ctl.reload(evt.GpioPin, evt.Leds, evt.Brightness, evt.StripType, evt.Segments)
	if err != nil {
		fmt.Println(err)
	}
}

func (ctl *Controller) HandleAmbilightEvent(b []byte) {
	if ctl.ws == nil {
		return
	}

	var evt events.AmbilightEvent
	err := msgpack.Unmarshal(b, &evt)
	if err != nil {
		panic(err)
	}

	if ctl.mode != Ambilight {
		ctl.ws.Stop()
		ctl.mode = Ambilight
	}

	out := "\n"
	for _, seg := range evt.Segments {
		if len(ctl.segments) <= seg.Id {
			panic("segment doesn't exist")
		}

		segment := ctl.segments[seg.Id]

		for i := 0; i < (segment.end-segment.start)*4; i += 4 {
			// Parse color data for current LED
			r := seg.Pix[i]
			g := seg.Pix[i+1]
			b := seg.Pix[i+2]
			// Set the current LED's color
			// Not need to check for error
			err := ctl.ws.SetLedColor(i/4+segment.start, r, g, b)
			if err != nil {
				fmt.Println(err)
			}

			out += color.RGB(seg.Pix[i], seg.Pix[i+1], seg.Pix[i+2], true).Sprintf(" ")
		}

	}
	fmt.Print(out)

	ctl.mux.Lock()
	if ctl.rendering {
		ctl.mux.Unlock()

		return
	}

	ctl.rendering = true
	ctl.mux.Unlock()

	go func() {
		defer func() {
			ctl.mux.Lock()
			ctl.rendering = false
			ctl.mux.Unlock()
		}()

		err = ctl.ws.Render()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// TODO: properly handle out of bounds
	//if ctl.leds != len(evt.Data) / 4 {
	//	ctl.leds = len(evt.Data) / 4
	//}

	//if ctl.rendering {
	//	return
	//}
	//
	//ctl.rendering = true

	//segment = ctl.segments[1]
	//
	//for i := 0; i < (segment.end-segment.start)*4; i += 4 {
	//	// Parse color data for current LED
	//	r := evt.Data[i]
	//	g := evt.Data[i+1]
	//	b := evt.Data[i+2]
	//	// Set the current LED's color
	//	// Not need to check for error
	//	err := ctl.ws.SetLedColor(i/4+segment.start, r, g, b)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}

	//ctl.rendering = false
	//}

	//fmt.Print(evt.SegmentId)
}
