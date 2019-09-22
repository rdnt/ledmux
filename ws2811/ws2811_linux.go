// +build linux,cgo darwin,cgo

package ws2811

import (
	"errors"
	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

// Engine asdad
type Engine struct {
	engine   *ws2811.WS2811
	leds     []uint32
	ledCount int
}

// Init asdad
func Init(pin int, ledCount int, brightness int) (*Engine, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount
	opt.Channels[0].GpioPin = pin
	ws, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}
	err = ws.Init()
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		leds:     make([]uint32, ledCount),
		engine:   ws,
		ledCount: ledCount,
	}
	return engine, nil
}

// Fini aosdasd
func (ws *Engine) Fini() {
	ws.engine.Fini()
}

// SetLedsSync adad
func (ws *Engine) SetLedsSync(channel int, leds []uint32) error {
	return ws.engine.SetLedsSync(channel, leds)
}

// Clear resets all the leds (turns them off by setting their color to black)
func (ws *Engine) Clear() error {
	ws.leds = make([]uint32, ws.ledCount)
	return ws.Render()
}

// Render renders the colors saved on the leds array onto the led strip
func (ws *Engine) Render() error {
	ws.engine.SetLedsSync(0, ws.leds)
	return ws.engine.Render()
}

// SetLedColor changes the color of the led in the specified index
func (ws *Engine) SetLedColor(index int, r uint8, g uint8, b uint8) error {
	if index >= len(ws.leds) || index < 0 {
		return errors.New("Invalid led index")
	}
	// WRGB
	color := uint32(0xff)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b)
	ws.leds[index] = color
	return nil
}
