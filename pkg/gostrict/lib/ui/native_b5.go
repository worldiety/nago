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
)

// The conversions of batch b5, see native.go.

// b5MenuEntries returns the menu entries of nago, which the handles hold.
func b5MenuEntries(entries []ScaffoldMenuEntry) []nui.ScaffoldMenuEntry {
	res := make([]nui.ScaffoldMenuEntry, 0, len(entries))
	for _, e := range entries {
		res = append(res, get[nui.ScaffoldMenuEntry](e.ref))
	}
	return res
}

// b5ViewWithSizeClasses returns the matches of nago, which the handles hold.
func b5ViewWithSizeClasses(matches []ViewWithSizeClass) []nui.ViewWithSizeClass {
	res := make([]nui.ViewWithSizeClass, 0, len(matches))
	for _, m := range matches {
		res = append(res, get[nui.ViewWithSizeClass](m.ref))
	}
	return res
}

// b5ViewFunc converts a function of the program, which returns a view, into a function of nago. A nil function stays
// nil, so that nago reports it.
func b5ViewFunc(fn func() View) func() core.View {
	if fn == nil {
		return nil
	}

	return func() core.View {
		return viewOf(fn())
	}
}
