// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	"go.wdy.de/nago/presentation/core"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/navsplitview"
)

// The conversions of batch c6, see native.go.

// c6Factory turns a function of the program into the factory of nago, which creates the views of a split view.
func c6Factory(fn func(id NavsplitviewViewID) View) navsplitview.Factory {
	return navsplitview.NavFn(func(id navsplitview.ViewID) core.View {
		return viewOf(fn(NavsplitviewViewID(id)))
	})
}

// c6Decored returns the view of nago of a view, which must be a view with padding, frame and border, or nil.
func c6Decored(v View) nui.DecoredView {
	d, _ := viewOf(v).(nui.DecoredView)
	return d
}
