// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Later passes functions to the host, which the host calls later and which return a view as a ui.View.
func Later(dp float64) string {
	w := &widths{}
	_ = ui.Tabs(ui.TabsPage("a", func() ui.View {
		return ui.Hero("h")
	}))
	_ = ui.SizeClass(ui.SizeClassSmall, func() ui.View {
		return ui.CodeEditor("x")
	})
	_ = ui.Text("t").Frame(frame(dp)).WithFrame(w.record())
	return w.s
}
