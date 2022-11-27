package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"ledctl3/internal/pkg/event"
	"ledctl3/internal/server/config"
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

type Controller struct {
	config config.Config
	mux    sync.Mutex
	events chan []byte
	ws     *ws281x.Engine
	leds   int
	//segments  []Segment
	mode      Mode
	rendering bool
	buffer    []byte
}

//type Segment struct {
//	id    int
//	start int
//	end   int
//}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65535, //1024
	WriteBufferSize: 65535,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func New(cfg config.Config) (*Controller, error) {
	events := make(chan []byte, 1)
	ctl := &Controller{
		config: cfg,
		events: events,
		ws:     nil,
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

func (c *Controller) Handle(typ event.Type, b []byte) {
	switch typ {
	case event.Update:
		c.HandleUpdateEvent(b)
	case event.Render:
		c.HandleRenderEvent(b)
	default:
		fmt.Println("unknown event")
	}
}

func (c *Controller) Start() error {
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
					fmt.Println("error during read", err)
					return
				}

				if typ != websocket.TextMessage {
					fmt.Println("invalid message type")
					continue
				}

				var e event.Event
				err = json.Unmarshal(b, &e)
				if err != nil {
					fmt.Println("invalid event format")
					continue
				}

				c.Handle(e.Type, b)
			}
		},
	)

	go http.ListenAndServe(":4197", nil)

	return nil
}

func (c *Controller) reload(gpioPin, ledsCount, brightness int, stripType string) error {
	if c.ws != nil {
		err := c.ws.Clear()
		if err != nil {
			return err
		}

		c.ws.Fini()
	}

	engine, err := ws281x.Init(gpioPin, ledsCount, brightness, stripType)
	if err != nil {
		return err
	}

	c.ws = engine

	//i := 0
	//for _, s := range segments {
	//	c.segments = append(
	//		c.segments, Segment{
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

func (c *Controller) HandleUpdateEvent(b []byte) {
	var evt event.UpdateEvent

	err := json.Unmarshal(b, &evt)
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg := evt.Segments[0]

	err = c.reload(cfg.GpioPin, cfg.Leds, cfg.Brightness, cfg.StripType)
	if err != nil {
		fmt.Println(err)
	}
}

func (c *Controller) HandleRenderEvent(b []byte) {
	if c.ws == nil {
		return
	}

	var evt event.SetLedsEvent
	err := json.Unmarshal(b, &evt)
	if err != nil {
		panic(err)
	}

	if c.mode != Render {
		c.ws.Stop()
		c.mode = Render
	}

	//out := "\n"
	for _, seg := range evt.Segments {
		if len(c.config.Segments) <= seg.Id {
			panic("segment doesn't exist")
		}

		//segment := c.config.Segments[seg.Id]
		//
		//for i := 0; i < (segment.end-segment.start)*4; i += 4 {
		//	// Parse color data for current LED
		//	r := seg.Pix[i]
		//	g := seg.Pix[i+1]
		//	b := seg.Pix[i+2]
		//	// Set the current LED's color
		//	// Not need to check for error
		//	err := c.ws.SetLedColor(i/4+segment.start, r, g, b)
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//
		//	//out += color.RGB(seg.Pix[i], seg.Pix[i+1], seg.Pix[i+2], true).Sprintf(" ")
		//}

	}
	//fmt.Print(out)

	c.mux.Lock()
	if c.rendering {
		c.mux.Unlock()

		return
	}

	c.rendering = true
	c.mux.Unlock()

	go func() {
		defer func() {
			c.mux.Lock()
			c.rendering = false
			c.mux.Unlock()
		}()

		err = c.ws.Render()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// TODO: properly handle out of bounds
	//if c.leds != len(evt.Data) / 4 {
	//	c.leds = len(evt.Data) / 4
	//}

	//if c.rendering {
	//	return
	//}
	//
	//c.rendering = true

	//segment = c.segments[1]
	//
	//for i := 0; i < (segment.end-segment.start)*4; i += 4 {
	//	// Parse color data for current LED
	//	r := evt.Data[i]
	//	g := evt.Data[i+1]
	//	b := evt.Data[i+2]
	//	// Set the current LED's color
	//	// Not need to check for error
	//	err := c.ws.SetLedColor(i/4+segment.start, r, g, b)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}

	//c.rendering = false
	//}

	//fmt.Print(evt.SegmentId)
}
