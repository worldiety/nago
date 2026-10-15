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

// TThemeSwitcher is a menu to switch the color scheme, e.g. between light, dark and system, see ThemeSwitcher.
type TThemeSwitcher struct{ ref js.Ref }

func (TThemeSwitcher) isView() {}

// ThemeSwitcher returns a menu to switch the color scheme, which is anchored to the view anchor.
//
//gostrict:host
func ThemeSwitcher(anchor View) TThemeSwitcher {
	return TThemeSwitcher{ref: wrap(nui.ThemeSwitcher(viewOf(anchor)))}
}

// Frame sets the size and the position of the menu.
//
//gostrict:host
func (c TThemeSwitcher) Frame(frame Frame) TThemeSwitcher {
	return TThemeSwitcher{ref: wrap(get[nui.TThemeSwitcher](c.ref).Frame(frame.nago()))}
}
