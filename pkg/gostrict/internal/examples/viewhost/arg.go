// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Arg passes a TText to the host as the argument of a parameter of the type ui.View.
func Arg(dp float64) string {
	w := &widths{}
	_ = ui.BoxAlign(ui.Center, ui.Text("a")).Frame(frame(dp)).WithFrame(w.record())
	return w.s
}
