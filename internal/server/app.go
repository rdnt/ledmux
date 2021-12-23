package server

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"ledctl3/pkg/events"
	"ledctl3/pkg/udp"
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
	conn udp.Server
	ws   *ws281x.Engine
	leds int
	mode Mode
}

func New() (*Controller, error) {
	server, err := udp.NewServer(":4197")
	if err != nil {
		return nil, err
	}

	ctl := &Controller{
		conn: server,
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
	}
}

func (ctl *Controller) Start() error {
	//ws, err := ws281x.Init(18, 83, 255, "GRB")
	//if err != nil {
	//	return err
	//}

	//ctl.ws = ws

	go func() {
		for b := range ctl.conn.Receive() {
			var e events.Event

			err := msgpack.Unmarshal(b, &e)
			if err != nil {
				panic(err)
			}

			ctl.Route(e, b)
		}
	}()
	return nil
}

func (ctl *Controller) reload(gpioPin, ledsCount, brightness int, stripType string) error {
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

	return nil
}

func (ctl *Controller) HandleReloadEvent(b []byte) {
	var evt events.ReloadEvent

	err := msgpack.Unmarshal(b, &evt)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ctl.reload(evt.GpioPin, evt.Leds, evt.Brightness, evt.StripType)
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
	// TODO: properly handle out of bounds
	//if ctl.leds != len(evt.Data) / 4 {
	//	ctl.leds = len(evt.Data) / 4
	//}

	for i := 0; i < len(evt.Data); i += 4 {
		// Parse color data for current LED
		r := evt.Data[i]
		g := evt.Data[i+1]
		b := evt.Data[i+2]
		// Set the current LED's color
		// Not need to check for error
		err := ctl.ws.SetLedColor(i/4, r, g, b)
		if err != nil {
			fmt.Println(err)
		}
	}

	err = ctl.ws.Render()
	if err != nil {
		fmt.Println(err)
	}
}
