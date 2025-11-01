// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"slices"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// TForm is a crud component(CRUD Form).
// Form creates a form view based on the field bindings. The given value is mapped automatically, based on the binding.
// The implementation pushes automatically.
type TForm[T any] struct {
	bnd   *Binding[T]
	state *core.State[T]

	padding            ui.Padding
	gap                ui.Length
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

// deprecated: use [entities.NewUseCases]
func Form[T any](bnd *Binding[T], state *core.State[T]) TForm[T] {
	return TForm[T]{
		bnd:   bnd,
		state: state,
		gap:   ui.L16,
		frame: ui.Frame{}.FullWidth(),
	}
}

func (t TForm[T]) Render(ctx core.RenderContext) core.RenderNode {
	return ui.VStack(
		slices.Collect(func(yield func(view core.View) bool) {
			for _, field := range t.bnd.fields {
				if field.RenderFormElement != nil {
					yield(ui.Composable(func() core.View {
						return field.RenderFormElement(field, t.state)
					}))
				}
			}

		})...,
	).Gap(t.gap).
		Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TForm[T]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TForm[T]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TForm[T]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TForm[T]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TForm[T]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TForm[T]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
