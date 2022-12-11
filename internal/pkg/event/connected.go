package event

type ConnectedEvent struct {
	Event      Type                    `json:"event"`
	Brightness int                     `json:"brightness"`
	GpioPin    int                     `json:"gpioPin"`
	StripType  string                  `json:"stripType"`
	Segments   []ConnectedEventSegment `json:"segments"`
}

type ConnectedEventSegment struct {
	Id   int `json:"id"`
	Leds int `json:"leds"`
}

func (e ConnectedEvent) Type() Type {
	return Connected
}

//func NewConnectedEvent(segments []TurnOnEventSegment) TurnOnEvent {
//	return TurnOnEvent{
//		Event: Event{
//			Event: TurnOn,
//		},
//		Segments: segments,
//	}
//}
