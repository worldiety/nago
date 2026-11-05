// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Nested passes views to the host, which contain views, which crossed before.
func Nested(dp float64) string {
	w := &widths{}
	_ = ui.HStack(
		ui.ScrollView(ui.List(ui.Card("c"))),
		ui.Grid(ui.GridCell(ui.AlertBanner("t", "m"))),
	).Frame(frame(dp)).WithFrame(w.record())
	return w.s
}
