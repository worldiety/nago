// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import "go.wdy.de/nago/presentation/proto"

// DP is Density-independent pixels: an abstract unit that is based on the physical density of the screen.
// These units are relative to a 160 dpi (dots per inch) screen, on which 1 dp is roughly equal to 1 px.
// When running on a higher density screen, the number of pixels used to draw 1 dp is scaled up by a factor
// appropriate for the screen's dpi.
//
// Likewise, when on a lower-density screen, the number of pixels used for 1 dp is scaled down.
// The ratio of dps to pixels changes with the screen density, but not necessarily in direct proportion.
// Using dp units instead of px units is a solution to making the view dimensions in your layout
// resize properly for different screen densities. It provides consistency for the real-world sizes of
// your UI elements across different devices.
// Source: https://developer.android.com/guide/topics/resources/more-resources.html#Dimension
type DP float64

// Density describes the scale factor of physical pixels to screen pixels normalized to a 160dpi screen.
// This is identical to the Android specification. On a 160dpi screen, this factor is 1. Note, that
// this may also be used to optimize accessibility which makes everything equally larger. There is also the
// concept of SP, but that is usually implemented at the frontend interpreter anyway.
type Density float64

// Weight is between 0-1 and can be understood as 1 = 100%, however implementations must normalize the total
// of all weights and recalculate the effective percentage.
type Weight float64

type WindowInfo struct {
	Width       DP
	Height      DP
	Density     Density
	SizeClass   WindowSizeClass
	ColorScheme ColorScheme
}

// WindowSizeClass represents media break points of the screen which an ora application is shown.
// The definition of a size class is disjunct and for all possible sizes, exact one size class will match.
// See also https://developer.android.com/develop/ui/views/layout/window-size-classes and
// https://tailwindcss.com/docs/responsive-design.
type WindowSizeClass uint

func (w WindowSizeClass) Ordinal() int {
	switch w {
	case SizeClassSmall:
		return 1
	case SizeClassMedium:
		return 2
	case SizeClassLarge:
		return 3
	case SizeClassXL:
		return 4
	case SizeClass2XL:
		return 5
	default:
		return 0
	}
}

func (w WindowSizeClass) Valid() bool {
	return w.Ordinal() != 0
}

const (
	// SizeClassSmall are devices below 640 dp screen width.
	SizeClassSmall WindowSizeClass = WindowSizeClass(proto.SizeClassSmall)
	// SizeClassMedium are devices below 768dp screen width.
	SizeClassMedium WindowSizeClass = WindowSizeClass(proto.SizeClassMedium)
	// SizeClassLarge are devices below 1024dp screen width.
	SizeClassLarge WindowSizeClass = WindowSizeClass(proto.SizeClassLarge)
	// SizeClassXL are devices below 1280dp screen width.
	SizeClassXL WindowSizeClass = WindowSizeClass(proto.SizeClassXL)
	// SizeClass2XL are devices below 1536dp screen width.
	SizeClass2XL WindowSizeClass = WindowSizeClass(proto.SizeClass2XL)
)
