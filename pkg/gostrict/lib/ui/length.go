// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//gostrict:header
package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// Length is a length of CSS, e.g. "1rem", "42dp", "50%", "100dvh" or "calc(100dvh - 2rem)". The empty Length is
// the default of a view.
type Length string

const (
	Auto Length = ""
	// ViewportHeight is the height of the smallest viewport, e.g. to center a view vertically on the screen.
	ViewportHeight Length = "100dvh"
)

const (
	L0    Length = "0px"
	L1    Length = "1px"
	L2    Length = "0.125rem"
	L4    Length = "0.25rem"
	L8    Length = "0.5rem"
	L12   Length = "0.75rem"
	L14   Length = "0.875rem"
	L16   Length = "1rem"
	L20   Length = "1.25rem"
	L24   Length = "1.5rem"
	L32   Length = "2rem"
	L40   Length = "2.5rem"
	L44   Length = "2.75rem"
	L48   Length = "3rem"
	L60   Length = "3.75rem"
	L64   Length = "4rem"
	L80   Length = "5rem"
	L96   Length = "6rem"
	L120  Length = "7.5rem"
	L144  Length = "9rem"
	L160  Length = "10rem"
	L200  Length = "12.5rem"
	L256  Length = "16rem"
	L320  Length = "20rem"
	L400  Length = "25rem"
	L480  Length = "30rem"
	L560  Length = "35rem"
	L880  Length = "55rem"
	L1200 Length = "75rem"
	L1600 Length = "100rem"
)

// Full is the full length of the parent.
//
//gostrict:js "100%"
var Full Length = Length(nui.Full)

// Absolute returns a length in dp.
//
//gostrict:host
func Absolute(dp float64) Length {
	return Length(nui.Absolute(core.DP(dp)))
}

// Relative returns a length relative to the parent, e.g. 0.5 is 50%.
//
//gostrict:host
func Relative(weight float64) Length {
	return Length(nui.Relative(core.Weight(weight)))
}

// L returns a length in dp as rem, like the constants L16 and others.
//
//gostrict:host
func L(dp float64) Length {
	return Length(nui.L(dp))
}
