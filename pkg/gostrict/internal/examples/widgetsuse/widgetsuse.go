// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package widgetsuse builds views of the header package widgets like a view of lib/ui.
package widgetsuse

import "go.wdy.de/nago/pkg/gostrict/internal/examples/widgets"

// State counts the clicks.
type State struct {
	Clicks int32
}

// Run builds a tree of views and describes it.
func Run(n int32) string {
	s := &State{}
	title := widgets.Label("title").Padding(widgets.Insets{}.All(2))
	box := widgets.Box(
		title,
		widgets.Label("wide").Padding(widgets.Wide).OnClick(func() int32 {
			s.Clicks += n
			return s.Clicks
		}),
		widgets.Box(widgets.Label("inner")).Align(widgets.Center),
	).Align(widgets.End)
	return widgets.Describe(box)
}
