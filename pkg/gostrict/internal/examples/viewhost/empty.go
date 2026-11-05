// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// Empty calls a function of the host with a variadic parameter of the type ui.View without any view.
func Empty(dp float64) string {
	w := &widths{}
	_ = ui.VStack().Frame(frame(dp)).WithFrame(w.record())
	return w.s
}
