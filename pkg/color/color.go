package color

import (
	"encoding/hex"
	"errors"
	"fmt"
	"image/color"
	"strings"
)

func ToString(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x%02x", r>>8, g>>8, b>>8, a>>8)
}

func FromString(s string) (c color.Color, err error) {
	clr := make([]byte, 4)

	n, err := hex.Decode(clr, []byte(strings.TrimLeft(s, "#")))
	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, errors.New("could not parse rgba color")
	}

	return color.RGBA{
		R: clr[0],
		G: clr[1],
		B: clr[2],
		A: clr[3],
	}, nil
}
