package gradient

import (
	"errors"
	"fmt"

	"github.com/gookit/color"
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

// New parses the colors into a linearly-interpolated gradient
func New(colors ...colorful.Color) (Gradient, error) {
	if len(colors) < 2 {
		return Gradient{}, errors.New("minimum two colors required")
	}

	g := Gradient{}

	for i, clr := range colors {
		r, gg, b, _ := clr.RGBA()

		r = r >> 8
		gg = gg >> 8
		b = b >> 8

		color.RGB(uint8(r), uint8(gg), uint8(b), true).Print("    ")
		g = append(g, Keypoint{
			Color:    clr,
			Position: float64(i) / float64(len(colors)-1),
		})
	}
	fmt.Println(g)

	return g, nil
}
