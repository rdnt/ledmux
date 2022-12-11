package event

type UpdateEvent struct {
	Event    Type                 `json:"event"`
	Segments []UpdateEventSegment `json:"segments"`
}

type UpdateEventSegment struct {
	Id         int    `json:"id"`
	Leds       int    `json:"leds"`
	StripType  string `json:"stripType"`
	GpioPin    int    `json:"gpioPin"`
	Brightness int    `json:"brightness"`
}

func (e UpdateEvent) Type() Type {
	return Update
}
