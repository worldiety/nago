// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"slices"
)

// TCards is a crud component(CRUD Cards).
type TCards[Entity data.Aggregate[ID], ID data.IDType] struct {
	opts TOptions[Entity, ID]

	padding            ui.Padding
	gap                ui.Length
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func Cards[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) TCards[Entity, ID] {
	return TCards[Entity, ID]{
		opts:    opts,
		padding: ui.Padding{}.All(ui.L16),
		gap:     ui.L16,
	}
}

func (t TCards[Entity, ID]) Render(ctx core.RenderContext) core.RenderNode {
	ds := t.opts.datasource()
	bnd := t.opts.bnd

	return ui.VStack(
		ui.Each(slices.Values(ds.List()), func(entity Entity) core.View {
			entityState := core.StateOf[Entity](t.opts.wnd, fmt.Sprintf("crud.card.entity.%v", entity.Identity())).Init(func() Entity {
				return entity
			})
			return Card[Entity](bnd, entityState).Frame(ui.Frame{}.FullWidth())
		})...,
	).Gap(t.gap).
		Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TCards[Entity, ID]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TCards[Entity, ID]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TCards[Entity, ID]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TCards[Entity, ID]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TCards[Entity, ID]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TCards[Entity, ID]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
