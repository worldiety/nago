// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package handlevalues

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

func passImage(f func(ui.Frame) ui.TImage, dp float64) ui.TImage {
	return f(frame(dp))
}

func passDivider(f func(ui.Frame) ui.TDivider, dp float64) ui.TDivider {
	return f(frame(dp))
}

// PassFrame passes the method values Frame of a TImage and of a TDivider to functions, which call them.
func PassFrame(dp float64) string {
	w := &widths{}
	_ = passImage(ui.ImageIcon("<svg/>").Frame, dp).WithFrame(w.record())
	_ = passDivider(ui.HLine().Frame, 2*dp)
	return w.s
}
