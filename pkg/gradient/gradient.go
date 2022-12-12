package gradient

import (
	"errors"
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

// Gradient represents a color gradient
type Gradient []Step

// Step is a point in the gradient that has a specific color
type Step struct {
	Color    color.Color
	Position float64
}

// GetInterpolatedColor returns the color between the two steps that are
// around t.
func (gt Gradient) GetInterpolatedColor(t float64) color.Color {
	for i := 0; i < len(gt)-1; i++ {
		c1 := gt[i]
		c2 := gt[i+1]
		c1c, _ := colorful.MakeColor(c1.Color)
		c2c, _ := colorful.MakeColor(c2.Color)
		if c1.Position <= t && t <= c2.Position {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Position) / (c2.Position - c1.Position)

			return c1c.BlendLuv(c2c, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient step.
	return gt[len(gt)-1].Color
}

// New parses the colors into a linearly-interpolated gradient
func New(colors ...color.Color) (Gradient, error) {
	if len(colors) < 2 {
		return Gradient{}, errors.New("minimum two colors required")
	}

	g := Gradient{}

	for i, clr := range colors {
		g = append(g, Step{
			Color:    clr,
			Position: float64(i) / float64(len(colors)-1),
		})
	}

	return g, nil
}
