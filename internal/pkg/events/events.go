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
	Type Type `json:"event"`
}

type AmbilightEvent struct {
	Event

	Segments []Segment `json:"segments"`
}

type Segment struct {
	Id  int    `json:"id"`
	Pix []byte `json:"pix"`
}

func NewAmbilightEvent(segs []Segment) AmbilightEvent {
	return AmbilightEvent{
		Event: Event{
			Type: Ambilight,
		},
		Segments: segs,
	}
}

type ReloadEvent struct {
	Event
	Leds       int             `json:"leds"`
	StripType  string          `json:"stripType"`
	GpioPin    int             `json:"gpioPin"`
	Brightness int             `json:"brightness"`
	Segments   []SegmentConfig `json:"segments"`
}

type SegmentConfig struct {
	Id   int `json:"id"`
	Leds int `json:"leds"`
}

func NewReloadEvent(leds int, stripType string, gpioPin, brightness int, segments []SegmentConfig) ReloadEvent {
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
	Color [4]byte `json:"color"`
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
	Payload []byte `json:"payload"`
}

func NewUpdateEvent(payload []byte) UpdateEvent {
	return UpdateEvent{
		Event: Event{
			Type: Update,
		},
		Payload: payload,
	}
}
