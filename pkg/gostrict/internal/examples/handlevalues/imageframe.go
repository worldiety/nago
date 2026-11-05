// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package handlevalues

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// ImageFrame sets the frame of a TImage through the method value Frame.
func ImageFrame(dp float64) string {
	w := &widths{}
	f := ui.ImageIcon("<svg/>").Frame
	_ = f(frame(dp)).WithFrame(w.record())
	return w.s
}
