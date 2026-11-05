// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package handlevalues uses the methods of handles of lib/ui as values, e.g. f := ui.Text("a").Frame. The checker
// either accepts such a program, then it translates and runs natively, in goja and in the vm alike, or rejects it
// with a finding of its own. Each other file is a program of its own together with this one. The other files are
// excluded from the subset by their build constraint, so that the package as a whole is checked without them,
// whichever way the checker decides; the tests load each with the constraint removed. Each function returns the
// widths of the frames, which the host gives back to WithFrame, so that its result passes through the host.
package handlevalues

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
