//go:build (!linux && !darwin) || !cgo

package ws281x

import (
	"fmt"
	"image/color"
	"sync"

	gcolor "github.com/gookit/color"
)

// Engine struct placeholder
type Engine struct {
	wg        *sync.WaitGroup
	stop      chan struct{}
	rendering bool
	ledsCount int
	leds      []color.Color
}

// Init placeholder function -- initializes all waitgroup logic on windows
func Init(_ int, ledsCount int, _ int, _ string) (*Engine, error) {
	// Create the effects waitgroup
	wg := sync.WaitGroup{}
	// Add main routine's delta to the waitgroup
	wg.Add(1)
	// Initialize stop channel that will stop any running effect goroutines
	stop := make(chan struct{})
	// Return a reference to the engine instance

	colors := make([]color.Color, ledsCount)
	for i := 0; i < ledsCount; i++ {
		colors[i] = color.RGBA{}
	}

	return &Engine{
		wg:        &wg,
		stop:      stop,
		ledsCount: ledsCount,
		leds:      colors,
	}, nil
}

// Cancel returns the stop channel
func (e *Engine) Cancel() chan struct{} {
	return e.stop
}

// Stop stops all running effects and prepares for new commands
func (e *Engine) Stop() {
	// Notify all running goroutines that they should cancel
	close(e.stop)
	// Decrement main routine's delta
	e.wg.Done()
	// Wait for goroutines to finish their work
	e.wg.Wait()
	// Turn off all leds
	e.Clear()
	// Add main routine's delta to waitgroup again
	e.wg.Add(1)
	// Re-initialize stop channel
	e.stop = make(chan struct{})
}

// Fini placeholder
func (*Engine) Fini() {}

// Clear placeholder
func (*Engine) Clear() error {
	return nil
}

// Render placeholder
func (e *Engine) Render() error {
	//fmt.Println()

	out := ""
	for _, c := range e.leds {
		r, g, b, _ := c.RGBA()
		out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	}
	fmt.Println(out)

	//g, err := gradient.New(e.leds...)
	//if err != nil {
	//	return err
	//}
	//
	//out := "\n"
	//for i := 0.0; i <= 1.0; i += 0.014 {
	//	c := g.GetInterpolatedColor(i)
	//	r, g, b, _ := c.RGBA()
	//	out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	//}
	//
	//fmt.Print(out)

	return nil
}

// SetLedColor placeholder
func (e *Engine) SetLedColor(id int, r uint8, g uint8, b uint8, a uint8) error {
	e.leds[id] = color.RGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}
	return nil
}
