// +build !linux,!darwin !cgo

package ws281x

import (
	"sync"
)

// Engine struct placeholder
type Engine struct {
	LedsCount int
	wg        *sync.WaitGroup
	stop      chan struct{}
	rendering bool
}

// Init placeholder function -- initializes all waitgroup logic on windows
func Init(_ int, _ int, _ int, _ string) (*Engine, error) {
	// Create the effects waitgroup
	wg := sync.WaitGroup{}
	// Add main routine's delta to the waitgroup
	wg.Add(1)
	// Initialize stop channel that will stop any running effect goroutines
	stop := make(chan struct{})
	// Return a reference to the engine instance
	return &Engine{
		wg:        &wg,
		stop:      stop,
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

// Fini placeholder
func (*Engine) Fini() {}

// Clear placeholder
func (*Engine) Clear() error {
	return nil
}

// Render placeholder
func (*Engine) Render() error {
	return nil
}

// SetLedColor placeholder
func (*Engine) SetLedColor(id int, r uint8, g uint8, b uint8) error {
	//fmt.Println("set led", id, r, g, b)
	return nil
}
