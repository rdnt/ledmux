package event

type Type string

const (
	Connected   Type = "connected"
	Update      Type = "update"
	SetLeds     Type = "set-leds"
	SetColor    Type = "set-color"
	SetEffect   Type = "set-effect"
	SetGradient Type = "set-gradient"
	TurnOn      Type = "turn-on"
	TurnOff     Type = "turn-off"
)

type Event interface {
	Type() Type
}
