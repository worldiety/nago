package xcolor

import (
	"fmt"
	"strconv"
	"strings"
)

type RGBA8 Vec4i8

func (c RGBA8) RGBA() RGBA {
	return RGBA{norm(c[0]), norm(c[1]), norm(c[2]), norm(c[3])}
}

func norm(v uint8) float32 {
	return max(0, min(1, float32(v)/255))
}

func Hex(c Color) string {
	c8 := c.RGBA().RGBA8()
	return fmt.Sprintf("#%02X%02X%02X%02X", c8[0], c8[1], c8[2], c8[3])
}

func MustParseHex(hex string) RGBA8 {
	c, err := ParseHex(hex)
	if err != nil {
		panic(err)
	}

	return c
}

func ParseHex(hex string) (RGBA8, error) {
	if !strings.HasPrefix(hex, "#") {
		return RGBA8{}, fmt.Errorf("unsupported hex color notation: %s", hex)
	}

	hex = hex[1:]
	switch len(hex) {
	case 8:
		//ok
	case 6:
		hex += "FF"
	default:
		return RGBA8{}, fmt.Errorf("unsupported hex color notation: %s", hex)
	}

	rgba, _ := strconv.ParseUint(hex, 16, 32)
	r := uint8((rgba >> 24) & 0xFF)
	g := uint8((rgba >> 16) & 0xFF)
	b := uint8((rgba >> 8) & 0xFF)
	a := uint8(rgba & 0xFF)

	return RGBA8{r, g, b, a}, nil
}
