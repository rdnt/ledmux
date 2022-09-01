package gradient

import (
	"errors"

	"github.com/lucasb-eyer/go-colorful"
)

// Gradient represents a color gradient
type Gradient []Keypoint

// Keypoint is a point in the gradient that has a specific color.
type Keypoint struct {
	Color    colorful.Color
	Position float64
}

// GetInterpolatedColor returns the color between the two keypoints that are
// around t.
func (gt Gradient) GetInterpolatedColor(t float64) colorful.Color {
	for i := 0; i < len(gt)-1; i++ {
		c1 := gt[i]
		c2 := gt[i+1]
		if c1.Position <= t && t <= c2.Position {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Position) / (c2.Position - c1.Position)
			return c1.Color.BlendLuv(c2.Color, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return gt[len(gt)-1].Color
}

// New parses the hex-encoded (#RRGGBB) colors into a gradient.
func New(colors ...string) (Gradient, error) {
	if len(colors) < 2 {
		return Gradient{}, errors.New("minimum two colors required")
	}

	g := Gradient{}

	for i, hex := range colors {
		clr, err := colorful.Hex(hex)
		if err != nil {
			return Gradient{}, err
		}

		g = append(g, Keypoint{
			Color:    clr,
			Position: float64(i) / float64(len(colors)-1),
		})
	}

	return g, nil
}
