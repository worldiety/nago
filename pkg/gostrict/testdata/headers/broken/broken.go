// Package broken binds Round to Math.round, which rounds half up, while math.Round rounds half away from zero.
//
//gostrict:header
package broken

import "math"

//gostrict:js (x) => Math.round(x)
func Round(x float64) float64 {
	return math.Round(x)
}
