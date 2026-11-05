// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Own builds a view, which crosses to the host only as the receiver of its methods, never as a ui.View.
func Own(dp float64) string {
	w := &widths{}
	_ = ui.Text("a").Frame(frame(dp)).WithFrame(w.record())
	return w.s
}
