package event

type UpdateEvent struct {
	Event
	Segments []UpdateEventSegment `json:"segments"`
}

type UpdateEventSegment struct {
	Id         int    `json:"id"`
	Leds       int    `json:"leds"`
	StripType  string `json:"stripType"`
	GpioPin    int    `json:"gpioPin"`
	Brightness int    `json:"brightness"`
}

func NewUpdateEvent(segments []UpdateEventSegment) UpdateEvent {
	return UpdateEvent{
		Event: Event{
			Type: Update,
		},
		Segments: segments,
	}
}
