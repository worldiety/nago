package ui

import (
	"fmt"
	"testing"
)

func Test_hsbToRGB(t *testing.T) {
	hsl := mustParseHSL("#1b8c30")
	expectedHSL := hslColor{
		H: 131,
		S: 0.68,
		L: 0.33,
	}

	if !nearlyEqualHSL(hsl, expectedHSL) {
		t.Errorf("expected: %v, got: %v", expectedHSL, hsl)
	}

	fmt.Println(hsl)

	if hsl.RGBHex() != "#1b8c30" {
		t.Errorf("expected: %v, got: %v", "1b8c30", hsl.RGBHex())
	}

	fmt.Println(Color("#1b8c30").WithBrightness(30).WithTransparency(10))
	fmt.Println(Color("#A6A5C2").WithBrightness(90))
	fmt.Println(Color("#1B8C30").WithBrightness(90))
}

func Test_hako(t *testing.T) {
	_, main := Color("#221A3F").RGBA()
	_, gray := Color("#999999").RGBA()
	fmt.Println(Color("#999999").Luminosity())

	fmt.Println(BlendLuminosity(main, gray))
}
