// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package fuel bounds a loop, whose end can not be proven otherwise, e.g. the iterations of a numeric method,
// for the terminating profile of the gostrict subset, see the directive //gostrict:terminating:
//
//	f := fuel.New(1000)
//	for f.Burn() && !converged(x) {
//		x = step(x)
//	}
//
// The loop ends at the latest, when the fuel is used up.
package fuel

// Fuel counts down the remaining iterations.
type Fuel struct {
	left int32
}

// New returns fuel for n iterations.
func New(n int32) *Fuel {
	return &Fuel{left: n}
}

// Burn uses one unit and reports, whether there was one left.
func (f *Fuel) Burn() bool {
	if f.left <= 0 {
		return false
	}
	f.left--
	return true
}

// Left returns the remaining units.
func (f *Fuel) Left() int32 {
	return f.left
}
