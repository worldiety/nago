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
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
)

// ViewWithSizeClass associates a view with a window size class, see SizeClass. It is a building block of
// ViewThatMatches, which allows different views for different screen sizes.
type ViewWithSizeClass struct{ ref js.Ref }

// TViewThatMatches is a utility component, which shows the view of the size class, which matches the window best, see
// ViewThatMatches.
type TViewThatMatches struct{ ref js.Ref }

func (TViewThatMatches) isView() {}

// ViewThatMatches returns a component, which shows the view of the size class of the window, which matches best. The
// matches are the views of the size classes, see SizeClass.
//
//gostrict:host
func ViewThatMatches(wnd Window, matches ...ViewWithSizeClass) TViewThatMatches {
	return TViewThatMatches{ref: wrap(nui.ViewThatMatches(wnd.nago(), b5ViewWithSizeClasses(matches)...))}
}
