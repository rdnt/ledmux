package audio

import (
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
