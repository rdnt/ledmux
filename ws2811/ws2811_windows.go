// +build !linux,!darwin !cgo

package ws2811

// Init placeholder
func Init(int, int, int) error {
	return nil
}

// Clear placeholder
func Clear() {}

// Fini placeholder
func Fini() {}

// SetLed placeholder
func SetLed(int, uint32) {}

// Render placeholder
func Render() error { return nil }
