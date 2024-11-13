package xcolor

import (
	"math"
)

// LAB represents the according CIELAB encoded Color value with alpha. See also https://colorizer.org/
// Currently the internal value range is L=[0;100], A=[-128;128] and B=[-128;128].
type LAB Vec4f

// WithLightness returns a new Vec4f. L is between 0 and 1.
func (c LAB) WithLightness(l float32) LAB {
	c[0] = l * 100
	return c
}

func (c LAB) RGBA() RGBA {
	r, g, b := lab2RGB(float64(c[0]), float64(c[1]), float64(c[2]))
	return RGBA8{r, g, b, fTU8(c[3])}.RGBA()
}

func rgb2LAB(R, G, B uint8) (l, a, b float64) {
	var RGB [3]float64
	var XYZ [3]float64

	RGB[0] = float64(R) * 0.003922
	RGB[1] = float64(G) * 0.003922
	RGB[2] = float64(B) * 0.003922

	RGB[0] = ifElse(RGB[0] > 0.04045, math.Pow((RGB[0]+0.055)/1.055, 2.4), RGB[0]/12.92)
	RGB[1] = ifElse(RGB[1] > 0.04045, math.Pow((RGB[1]+0.055)/1.055, 2.4), RGB[1]/12.92)
	RGB[2] = ifElse(RGB[2] > 0.04045, math.Pow((RGB[2]+0.055)/1.055, 2.4), RGB[2]/12.92)

	XYZ[0] = 0.412424*RGB[0] + 0.357579*RGB[1] + 0.180464*RGB[2]
	XYZ[1] = 0.212656*RGB[0] + 0.715158*RGB[1] + 0.0721856*RGB[2]
	XYZ[2] = 0.0193324*RGB[0] + 0.119193*RGB[1] + 0.950444*RGB[2]

	l = 116*ifElse(XYZ[1]/1.000000 > 0.008856, math.Pow(XYZ[1]/1.000000, 0.333333), 7.787*XYZ[1]/1.000000+0.137931) - 16
	a = 500 * (ifElse(XYZ[0]/0.950467 > 0.008856, math.Pow(XYZ[0]/0.950467, 0.333333), 7.787*XYZ[0]/0.950467+0.137931) - ifElse(XYZ[1]/1.000000 > 0.008856, math.Pow(XYZ[1]/1.000000, 0.333333), 7.787*XYZ[1]/1.000000+0.137931))
	b = 200 * (ifElse(XYZ[1]/1.000000 > 0.008856, math.Pow(XYZ[1]/1.000000, 0.333333), 7.787*XYZ[1]/1.000000+0.137931) - ifElse(XYZ[2]/1.088969 > 0.008856, math.Pow(XYZ[2]/1.088969, 0.333333), 7.787*XYZ[2]/1.088969+0.137931))

	return
}

func lab2RGB(L, A, B float64) (uint8, uint8, uint8) {
	var XYZ [3]float64
	var RGB [3]float64

	XYZ[1] = (L + 16) / 116
	XYZ[0] = A/500 + XYZ[1]
	XYZ[2] = XYZ[1] - B/200

	XYZ[1] = ifElse(XYZ[1]*XYZ[1]*XYZ[1] > 0.008856, XYZ[1]*XYZ[1]*XYZ[1], (XYZ[1]-(16/116))/7.787)
	XYZ[0] = ifElse(XYZ[0]*XYZ[0]*XYZ[0] > 0.008856, XYZ[0]*XYZ[0]*XYZ[0], (XYZ[0]-(16/116))/7.787)
	XYZ[2] = ifElse(XYZ[2]*XYZ[2]*XYZ[2] > 0.008856, XYZ[2]*XYZ[2]*XYZ[2], (XYZ[2]-(16/116))/7.787)

	RGB[0] = 0.950467*XYZ[0]*3.2406 + 1.000000*XYZ[1]*-1.5372 + 1.088969*XYZ[2]*-0.4986
	RGB[1] = 0.950467*XYZ[0]*-0.9689 + 1.000000*XYZ[1]*1.8758 + 1.088969*XYZ[2]*0.0415
	RGB[2] = 0.950467*XYZ[0]*0.0557 + 1.000000*XYZ[1]*-0.2040 + 1.088969*XYZ[2]*1.0570

	r := uint8(255 * ifElse(RGB[0] > 0.0031308, 1.055*math.Pow(RGB[0], 1/2.4)-0.055, RGB[0]*12.92))
	g := uint8(255 * ifElse(RGB[1] > 0.0031308, 1.055*math.Pow(RGB[1], 1/2.4)-0.055, RGB[1]*12.92))
	b := uint8(255 * ifElse(RGB[2] > 0.0031308, 1.055*math.Pow(RGB[2], 1/2.4)-0.055, RGB[2]*12.92))

	return r, g, b
}

func ifElse(condition bool, a, b float64) float64 {
	if condition {
		return a
	}
	return b
}
