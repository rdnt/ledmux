package strip

import (
	"errors"
)

var ErrInvalidStripType = errors.New("invalid strip type")

type Type string

const (
	RGBW Type = "rgbw"
	RBGW Type = "rbgw"
	GRBW Type = "grbw"
	GBRW Type = "gbrw"
	BRGW Type = "brgw"
	BGRW Type = "bgrw"
	RGB  Type = "rgb"
	RBG  Type = "rbg"
	GRB  Type = "grb"
	GBR  Type = "gbr"
	BRG  Type = "brg"
	BGR  Type = "bgr"
)

var types = map[string]Type{
	"rgbw": RGBW,
	"rbgw": RBGW,
	"grbw": GRBW,
	"gbrw": GBRW,
	"brgw": BRGW,
	"bgrw": BGRW,
	"rgb":  RGB,
	"rbg":  RBG,
	"grb":  GRB,
	"gbr":  GBR,
	"brg":  BRG,
	"bgr":  BGR,
}

func Parse(typ string) (Type, error) {
	t, ok := types[typ]
	if !ok {
		return "", ErrInvalidStripType
	}

	return t, nil
}
