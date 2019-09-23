// +build !linux,!darwin !cgo

package ws281x_wrapper

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

// Clear placeholder
func (*Engine) Clear() error {
	return nil
}

// Render placeholder
func (*Engine) Render() error {
	return nil
}

// SetLedColor placeholder
func (*Engine) SetLedColor(int, uint8, uint8, uint8) error {
	return nil
}
