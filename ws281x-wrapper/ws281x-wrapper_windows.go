// +build !linux,!darwin !cgo

package ws281x

// Engine struct placeholder
type Engine struct{}

// Init placeholder
func Init(int, int, int) (*Engine, error) {
	return nil, nil
}

// Fini placeholder
func (*Engine) Fini() {}

// SetLedsSync placeholder
func (*Engine) SetLedsSync(int, []uint32) error {
	return nil
}

// Clear resets all the leds (turns them off by setting their color to black)
func (*Engine) Clear() error {
	return nil
}

// Render renders the colors saved on the leds array onto the led strip
func (*Engine) Render() error {
	return nil
}

// SetLedColor changes the color of the led in the specified index
func (*Engine) SetLedColor(int, uint8, uint8, uint8) error {
	return nil
}
