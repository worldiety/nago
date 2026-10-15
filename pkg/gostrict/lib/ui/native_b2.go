// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package ui

import (
	nui "go.wdy.de/nago/presentation/ui"
)

// The conversions of batch b2, see native.go.

// b2GroupOf returns the group of nago, which the handle holds. The zero handle yields the empty group, which has no
// states, so that its methods do not panic.
func b2GroupOf(s RadioStateGroup) nui.RadioStateGroup {
	g := get[*nui.RadioStateGroup](s.ref)
	if g == nil {
		return nui.RadioStateGroup{}
	}
	return *g
}

// b2GroupFrom returns the handle, which holds the group of nago.
func b2GroupFrom(g nui.RadioStateGroup) RadioStateGroup {
	return RadioStateGroup{ref: wrap(&g)}
}
