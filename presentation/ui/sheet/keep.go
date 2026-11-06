// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sheet

import "go.wdy.de/nago/presentation/core"

// KeepStates keeps the states of the sheet with the id in a render, which does not render the sheet, e.g. while an
// error tile stands in its place, so that the sheet finds them again afterward, with the same pointers, e.g. the window
// the user looks at and its layout. A state keeps no observer, so a value, which a former tree sends meanwhile, calls
// no callback.
func KeepStates(wnd core.Window, id string) {
	prefix := "nago.sheet." + id
	core.StateOf[string](wnd, prefix+".viewport")
	core.StateOf[int64](wnd, prefix+".editseq")
	core.StateOf[map[string]string](wnd, prefix+".errors")
	core.StateOf[string](wnd, prefix+".edit")
	core.StateOf[Layout](wnd, prefix+".layout")
	core.StateOf[string](wnd, prefix+".sort")
	core.StateOf[string](wnd, prefix+".filter")
	core.StateOf[string](wnd, prefix+".copy")
	core.StateOf[string](wnd, prefix+".layoutop")
	core.StateOf[string](wnd, prefix+".paste")
}
