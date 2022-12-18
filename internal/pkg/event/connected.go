package event

type ConnectedEvent struct {
	Event      Type                    `json:"event"`
	Brightness int                     `json:"brightness"`
	GpioPin    int                     `json:"gpioPin"`
	StripType  string                  `json:"stripType"`
	Segments   []ConnectedEventSegment `json:"segments"`
	// TODO: add state
}

type ConnectedEventSegment struct {
	Id   int `json:"id"`
	Leds int `json:"leds"`
}

func (e ConnectedEvent) Type() Type {
	return Connected
}
