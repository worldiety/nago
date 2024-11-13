package xcolor

type Vec4f [4]float32

type Vec4i8 [4]uint8

func fTU8(v float32) uint8 {
	x := max(0, min(1, v))
	return uint8(x * 255)
}
