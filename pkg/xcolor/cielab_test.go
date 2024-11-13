package xcolor

import (
	"fmt"
	"math"
	"testing"
)

func TestNewCIELAB(t *testing.T) {

	rgb := RGBA8{
		34,
		26,
		63,
		255,
	}
	fmt.Println(rgb.RGBA().LAB())
	fmt.Println(Hex(rgb.RGBA().LAB()))
	fmt.Println(Hex(rgb.RGBA().LAB().WithLightness(0.6)))
	fmt.Println(rgb.RGBA().YUV())
	fmt.Println(Hex(rgb.RGBA().YUV()))
	fmt.Println()
	fmt.Println(rgb.RGBA().YPbPr())
	fmt.Println(Hex(rgb.RGBA().YPbPr().RGBA()))
	fmt.Println(Hex(rgb.RGBA().YPbPr().WithLuma(0.6)))
}

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1
}
