package event

type UpdateEvent struct {
	Event      Type                 `json:"event"`
	Leds       int                  `json:"leds"`
	StripType  string               `json:"stripType"`
	GpioPin    int                  `json:"gpioPin"`
	Brightness int                  `json:"brightness"`
	Segments   []UpdateEventSegment `json:"segments"`
	// TODO: add state
}

type UpdateEventSegment struct {
	Id   int `json:"id"`
	Leds int `json:"leds"`
}

func (e UpdateEvent) Type() Type {
	return Update
}
