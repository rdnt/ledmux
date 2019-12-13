package packet

// Update asd
type Update struct {
	Action     uint8
	LedsCount  uint16
	GPIOPin    uint8
	Brightness uint8
}

// Size returns the size in bytes an encoded update packet consumes
func (Update) Size() int {
	return 5
}

// Ambilight asd
type Ambilight struct {
	Action uint8
	Data   [1024 * 3]uint8
}

// Size returns the size in bytes an encoded update packet consumes
func (Ambilight) Size() int {
	return 1 + 1024*3
}

// Rainbow asd
type Rainbow struct {
	Action uint8
}

// Rainbow asd
type Clear struct {
	Action uint8
}
