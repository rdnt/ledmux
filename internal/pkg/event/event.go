package event

type Type string

const (
	Update      Type = "update"
	Render      Type = "render"
	SetColor    Type = "set-color"
	SetEffect   Type = "set-effect"
	SetGradient Type = "set-gradient"
	TurnOn      Type = "turn-on"
	TurnOff     Type = "turn-off"
)

type Event struct {
	Type Type `json:"event"`
}
