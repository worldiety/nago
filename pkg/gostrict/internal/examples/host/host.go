// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package host is a header package: its functions run natively in Go and through their bindings in JavaScript.
//
//gostrict:header
package host

import (
	"errors"
	"strconv"
)

//gostrict:js (a, b) => a * b + 1
func Mix(a, b float64) float64 {
	return a*b + 1
}

//gostrict:js (x, lo, hi) => Math.min(hi, Math.max(lo, x))
func Clamp(x, lo, hi float64) float64 {
	return min(hi, max(lo, x))
}

//gostrict:js (n) => "#" + n
func Label(n int32) string {
	return "#" + strconv.Itoa(int(n))
}

//gostrict:js (a, b) => b === 0 ? [0, $errors_New("division by zero")] : [a / b, null]
func Ratio(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}
