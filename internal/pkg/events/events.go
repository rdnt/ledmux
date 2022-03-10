package events

type Type string

const (
	Update    Type = "update"
	Ambilight Type = "ambilight"
	Rainbow   Type = "rainbow"
	Static    Type = "static"
	Reload    Type = "reload"
)

type Event struct {
	Type Type `msgpack:"event"`
}

type AmbilightEvent struct {
	Event
	SegmentId int    `msgpack:"segmentId"`
	Data      []byte `msgpack:"data"`
}

func NewAmbilightEvent(segmentId int, b []byte) AmbilightEvent {
	return AmbilightEvent{
		Event: Event{
			Type: Ambilight,
		},
		SegmentId: segmentId,
		Data:      b,
	}
}

type ReloadEvent struct {
	Event
	Leds       int       `msgpack:"leds"`
	StripType  string    `msgpack:"stripType"`
	GpioPin    int       `msgpack:"gpioPin"`
	Brightness int       `msgpack:"brightness"`
	Segments   []Segment `msgpack:"segments"`
}

type Segment struct {
	Id   int `msgpack:"id"`
	Leds int `msgpack:"leds"`
}

func NewReloadEvent(leds int, stripType string, gpioPin, brightness int, segments []Segment) ReloadEvent {
	return ReloadEvent{
		Event: Event{
			Type: Reload,
		},
		Leds:       leds,
		StripType:  stripType,
		GpioPin:    gpioPin,
		Brightness: brightness,
		Segments:   segments,
	}
}

type RainbowEvent struct {
	Event
}

func NewRainbowEvent() RainbowEvent {
	return RainbowEvent{
		Event: Event{
			Type: Rainbow,
		},
	}
}

type StaticEvent struct {
	Event
	Color [4]byte `msgpack:"color"`
}

func NewStaticEvent(color [4]byte) StaticEvent {
	return StaticEvent{
		Event: Event{
			Type: Static,
		},
		Color: color,
	}
}

type UpdateEvent struct {
	Event
	Payload []byte `msgpack:"payload"`
}

func NewUpdateEvent(payload []byte) UpdateEvent {
	return UpdateEvent{
		Event: Event{
			Type: Update,
		},
		Payload: payload,
	}
}
