// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

// TCard is a crud component(CRUD Card).
// Card creates a card view based on the field bindings. The given value is mapped automatically, based on the binding.
// A Card is usually readonly.
type TCard[T any] struct {
	bnd   *Binding[T]
	state *core.State[T]

	padding            ui.Padding
	gap                ui.Length
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func Card[T any](bnd *Binding[T], state *core.State[T]) TCard[T] {
	return TCard[T]{
		bnd:     bnd,
		state:   state,
		padding: ui.Padding{}.All(ui.L20),
		gap:     ui.L16,
		border:  ui.Border{}.Radius(ui.L16),
	}
}

func (t TCard[T]) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range t.bnd.fields {
				if field.RenderCardElement != nil {
					yield(field.RenderCardElement(field, t.state).Frame(ui.Frame{}.FullWidth()))
				}
			}

		})...,
	).Gap(t.gap).
		BackgroundColor(ui.ColorCardBody).
		Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TCard[T]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TCard[T]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TCard[T]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TCard[T]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TCard[T]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TCard[T]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
