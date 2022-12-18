package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEvent(t *testing.T) {
	b := []byte(`{"event":"set-leds"}`)

	events, err := Parse(b)
	assert.Nil(t, err)
	assert.Len(t, events, 1)
}

func TestParseEventArray(t *testing.T) {
	b := []byte(`[{"event":"turn-on"},{"event":"set-leds"}]`)

	events, err := Parse(b)
	assert.Nil(t, err)
	assert.Len(t, events, 2)
}
