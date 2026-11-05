// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package viewhost passes views of lib/ui to the host as a ui.View, while no exported function has the result
// ui.View: as an argument, in the list of a variadic parameter, nested and as the result of a function, which the
// host calls later. Each other file is a program of its own together with this one, so that the views of its types
// cross without the others. Each function returns the widths of the frames, which the host gives back to WithFrame,
// so that its result passes through the host.
package viewhost

import "go.wdy.de/nago/pkg/gostrict/lib/ui"

// widths collects the widths of the frames.
type widths struct{ s string }

// record returns the function for WithFrame, which appends the width of the frame.
func (w *widths) record() func(ui.Frame) ui.Frame {
	return func(fr ui.Frame) ui.Frame {
		w.s += string(fr.Width) + ";"
		return fr
	}
}

// frame returns a frame of the width dp.
func frame(dp float64) ui.Frame {
	return ui.Frame{Width: ui.Absolute(dp)}
}
