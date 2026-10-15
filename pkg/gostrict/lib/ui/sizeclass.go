// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
)

// WindowSizeClass is a size class of a window, which is a media breakpoint of the screen. Each width belongs to exactly
// one size class, see SizeClass.
type WindowSizeClass uint32

const (
	// SizeClassSmall are windows below 768dp width.
	SizeClassSmall WindowSizeClass = 0
	// SizeClassMedium are windows from 768dp up to below 1024dp width.
	SizeClassMedium WindowSizeClass = 1
	// SizeClassLarge are windows from 1024dp up to below 1280dp width.
	SizeClassLarge WindowSizeClass = 2
	// SizeClassXL are windows from 1280dp up to below 1536dp width.
	SizeClassXL WindowSizeClass = 3
	// SizeClass2XL are windows from 1536dp width.
	SizeClass2XL WindowSizeClass = 4
)

// SizeClass returns a view for the size class, which ViewThatMatches shows, if the window matches the size class best.
// The function view is called, when the window renders.
//
//gostrict:host
func SizeClass(class WindowSizeClass, view func() View) ViewWithSizeClass {
	return ViewWithSizeClass{ref: wrap(nui.SizeClass(core.WindowSizeClass(class), b5ViewFunc(view)))}
}
