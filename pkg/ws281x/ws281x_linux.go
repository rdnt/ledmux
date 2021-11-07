// +build linux,cgo darwin,cgo

package ws281x

import (
	"errors"
	"sync"

	ws281x "github.com/rpi-ws281x/rpi-ws281x-go"
)

// Engine represents a wrapper around the ws281x library.
// Holds the state of the leds, the state of the effects running, the leds count
//  and a reference to the underlying ws281x library instance
type Engine struct {
	LedsCount int
	engine    *ws281x.WS2811
	leds      []uint32
	wg        *sync.WaitGroup
	stop      chan struct{}
	rendering bool
}

// Init initializes a new instance of the ws281x library
func Init(pin int, ledCount int, brightness int, strip string) (*Engine, error) {
	// Initialize ws281x engine
	opt := ws281x.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount
	opt.Channels[0].GpioPin = pin
	st := 0
	switch strip {
	case "RGBW":
		st = ws281x.SK6812StripRGBW
	case "RBGW":
		st = ws281x.SK6812StripRBGW
	case "GRBW":
		st = ws281x.SK6812StripGRBW
	case "GBRW":
		st = ws281x.SK6812StrioGBRW
	case "BRGW":
		st = ws281x.SK6812StrioBRGW
	case "BGRW":
		st = ws281x.SK6812StripBGRW
	case "RGB":
		st = ws281x.WS2811StripRGB
	case "RBG":
		st = ws281x.WS2811StripRBG
	case "GRB":
		st = ws281x.WS2811StripGRB
	case "GBR":
		st = ws281x.WS2811StripGBR
	case "BRG":
		st = ws281x.WS2811StripBRG
	case "BGR":
		st = ws281x.WS2811StripBGR
	default:
		st = ws281x.WS2811StripGRB
	}
	opt.Channels[0].StripeType = st
	ws, err := ws281x.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}
	err = ws.Init()
	if err != nil {
		return nil, err
	}
	// Create the effects waitgroup
	wg := sync.WaitGroup{}
	// Add main routine's delta to the waitgroup
	wg.Add(1)
	// Initialize stop channel that will stop any running effect goroutines
	stop := make(chan struct{})
	return &Engine{
		LedsCount: ledCount,
		engine:    ws,
		leds:      make([]uint32, ledCount),
		wg:        &wg,
		stop:      stop,
		rendering: false,
	}, nil
}

// Add adds a delta of 1 to the waitgroup counter
func (ws *Engine) Add() bool {
	if ws.rendering {
		return false
	}
	ws.rendering = true
	ws.wg.Add(1)
	return true
}

// Done decrements the waitgroup counter by one
func (ws *Engine) Done() {
	ws.wg.Done()
	ws.rendering = false
}

// Cancel returns the stop channel
func (ws *Engine) Cancel() chan struct{} {
	return ws.stop
}

// Stop stops all running effects and prepares for new commands
func (ws *Engine) Stop() {
	// Notify all running goroutines that they should cancel
	close(ws.stop)
	// Decrement main routine's delta
	ws.wg.Done()
	// Wait for goroutines to finish their work
	ws.wg.Wait()
	// Turn off all leds
	ws.Clear()
	// Add main routine's delta to waitgroup again
	ws.wg.Add(1)
	// Re-initialize stop channel
	ws.stop = make(chan struct{})
}

// Fini does cleanup operations
func (ws *Engine) Fini() {
	ws.engine.Fini()
}

// Clear resets all the leds (turns them off by setting their color to black)
func (ws *Engine) Clear() error {
	ws.leds = make([]uint32, ws.LedsCount)
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
