package event

type Type string

const (
	Update      Type = "update"
	SetLeds     Type = "set-leds"
	SetColor    Type = "set-color"
	SetEffect   Type = "set-effect"
	SetGradient Type = "set-gradient"
	TurnOn      Type = "turn-on"
	TurnOff     Type = "turn-off"
)

type Event struct {
	Type Type `json:"event"`
}
