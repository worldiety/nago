package ui

import (
	"fmt"
	"math"
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
	c := Color("#221A3F")
	l := c.Luminosity()
	fmt.Println(Color("#221A3F").AddBrightness(l))
}

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1
}

func nearlyEqualHSL(a, b hslColor) bool {
	if !nearlyEqual(a.S, b.S) {
		return false
	}

	if !nearlyEqual(a.L, b.L) {
		return false
	}

	if !nearlyEqual(a.H, b.H) {
		return false
	}

	return true
}
