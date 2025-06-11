// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"go.wdy.de/nago/presentation/core"
)

type DecoredView interface {
	core.View
	Padding(padding Padding) DecoredView
	WithFrame(fn func(Frame) Frame) DecoredView
	Frame(frame Frame) DecoredView
	Border(border Border) DecoredView
	Visible(visible bool) DecoredView
	// AccessibilityLabel is used to help screen readers, see guidelines, when to use it.
	// https://www.w3.org/WAI/tutorials/images/decision-tree/
	AccessibilityLabel(label string) DecoredView
}

// A Composable is a lazy factory of a view which is evaluated lately in the render cycle.
// This can be used, to more efficiently post-pone resource allocations in conditional renderings.
type Composable func() core.View

type Lazy = Composable

func (c Composable) Render(context core.RenderContext) core.RenderNode {
	if c == nil {
		return nil
	}

	v := c()
	if v == nil {
		return nil
	}

	return c().Render(context)
}
