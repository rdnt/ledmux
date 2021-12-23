package audioviz

import (
	"ledctl3/internal/client/interfaces"
)

type Visualizer struct {
	source interfaces.AudioSource
	leds   int
}

func New(opts ...Option) (*Visualizer, error) {
	p := &Visualizer{}

	for _, opt := range opts {
		err := opt(p)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
