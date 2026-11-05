// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package handlevalues

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// WithFrame reads the frame of a TText through the method value WithFrame, which takes a function.
func WithFrame(dp float64) string {
	w := &widths{}
	g := ui.Text("a").Frame(frame(dp)).WithFrame
	_ = g(w.record())
	return w.s
}
