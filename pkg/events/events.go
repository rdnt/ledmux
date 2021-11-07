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
	GroupId int    `msgpack:"groupId"`
	Data    []byte `msgpack:"data"`
}

func NewAmbilightEvent(gid int, b []byte) AmbilightEvent {
	return AmbilightEvent{
		Event: Event{
			Type: Ambilight,
		},
		GroupId: gid,
		Data:    b,
	}
}

type ReloadEvent struct {
	Event
	Leds int `msgpack:"leds"`
	Brightness int `msgpack:"brightness"`
}

func NewReloadEvent(leds, brightness int) ReloadEvent {
	return ReloadEvent{
		Event: Event{
			Type: Reload,
		},
		Leds: leds,
		Brightness: brightness,
	}
}
