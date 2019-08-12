// +build linux,cgo darwin,cgo

package ws2811

import ws "github.com/jgarff/rpi_ws281x/golang/ws2811"

// Init placeholder
func Init(gpioPin int, ledCount int, brightness int) error {
	return ws.Init(gpioPin, ledCount, brightness)
}

// Clear placeholder
func Clear() {
	ws.Clear()
}

// Fini placeholder
func Fini() {
	ws.Fini()
}

// SetLed placeholder
func SetLed(index int, value uint32) {
	ws.SetLed(index, value)
}

// Render placeholder
func Render() error {
	return ws.Render()
}
