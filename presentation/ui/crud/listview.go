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
	"go.wdy.de/nago/presentation/ui/list"
	"slices"
)

// TList is a crud component(CRUD List).
type TList[Entity data.Aggregate[ID], ID data.IDType] struct {
	opts TOptions[Entity, ID]

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func List[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) TList[Entity, ID] {
	return TList[Entity, ID]{
		opts:  opts,
		frame: ui.Frame{}.FullWidth(),
	}
}

func (t TList[Entity, ID]) Render(ctx core.RenderContext) core.RenderNode {
	ds := t.opts.datasource()
	bnd := t.opts.bnd

	var count int
	return ui.VStack(
		list.List(ui.Each(slices.Values(ds.List()), func(entity Entity) core.View {
			count++
			if bnd.renderListEntry == nil {
				return list.Entry().Headline(fmt.Sprint(entity))
			}

			return bnd.renderListEntry(entity)
		})...).Caption(ui.Text("Alle Einträge")).
			Footer(ui.Text(fmt.Sprintf("%d von %d Einträgen", count, ds.totalCount))).
			Frame(ui.Frame{}.FullWidth()),
	).Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TList[Entity, ID]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TList[Entity, ID]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TList[Entity, ID]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TList[Entity, ID]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TList[Entity, ID]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TList[Entity, ID]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
