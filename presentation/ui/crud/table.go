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
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"reflect"
	"slices"
)

// TTable is a crud component(CRUD Table).
type TTable[Entity data.Aggregate[ID], ID data.IDType] struct {
	opts TOptions[Entity, ID]

	padding            ui.Padding
	frame              ui.Frame
	border             ui.Border
	accessibilityLabel string
	invisible          bool
}

func Table[Entity data.Aggregate[ID], ID data.IDType](opts TOptions[Entity, ID]) TTable[Entity, ID] {
	return TTable[Entity, ID]{
		opts: opts,
	}
}

func (t TTable[Entity, ID]) Render(ctx core.RenderContext) core.RenderNode {
	ds := t.opts.datasource()
	bnd := t.opts.bnd

	rows := ui.Each(slices.Values(ds.List()), func(entity Entity) ui.TTableRow {
		entityState := core.StateOf[Entity](t.opts.wnd, fmt.Sprintf("crud.row.entity.%v", entity.Identity())).Init(func() Entity {
			return entity
		})

		if !reflect.DeepEqual(entityState.Get(), entity) {
			entityState.Set(entity)
		}

		var cells []ui.TTableCell
		for _, field := range bnd.tableFields() {
			cells = append(cells, field.RenderTableCell(field, entityState))
		}

		return ui.TableRow(cells...).BackgroundColor(ui.ColorCardBody).HoveredBackgroundColor(ui.ColorCardFooter)
	})

	if err := ds.Error(); err != nil {
		return ui.VStack(alert.BannerError(err)).Frame(ui.Frame{}.FullWidth()).Render(ctx)
	}

	//todo localize

	rows = append(rows, ui.TableRow(ui.TableCell(ui.Text(fmt.Sprintf("%d Einträge", len(rows)))).ColSpan(bnd.CountTableColumns())).BackgroundColor(ui.ColorCardFooter))

	return ui.VStack(

		ui.Table(ui.Each(slices.Values(bnd.tableFields()), func(field Field[Entity]) ui.TTableColumn {
			var sortIcon core.SVG
			if field.Comparator != nil {
				sortIcon = heroSolid.ArrowsUpDown
			}

			if t.opts.sortByFieldState.Get() != nil && t.opts.sortByFieldState.Get().Label == field.Label {
				if t.opts.sortDirState.Get() == asc {
					sortIcon = heroSolid.ArrowUp
				} else {
					sortIcon = heroSolid.ArrowDown
				}
			}

			return ui.TableColumn(ui.IfElse(field.Comparator == nil, ui.Text(field.Label), ui.TertiaryButton(func() {
				if f := t.opts.sortByFieldState.Get(); f != nil && f.Label == field.Label {
					t.opts.sortDirState.Set(!t.opts.sortDirState.Get())
				} else {
					t.opts.sortByFieldState.Set(&field)
				}

			}).PreIcon(sortIcon).Title(field.Label).Font(ui.Font{Size: ui.L16, Weight: ui.BodyFontWeight}))).
				BackgroundColor(ui.ColorCardTop).
				Padding(ui.Padding{Left: ui.L0, Right: ui.L24, Top: ui.L16, Bottom: ui.L16})
		})...,
		).Rows(rows...).
			HeaderDividerColor("#00000000").
			Frame(ui.Frame{}.FullWidth()),
	).Visible(!t.invisible).
		Frame(t.frame).
		Border(t.border).
		Padding(t.padding).
		AccessibilityLabel(t.accessibilityLabel).
		Render(ctx)
}

func (t TTable[Entity, ID]) Padding(padding ui.Padding) ui.DecoredView {
	t.padding = padding
	return t
}

func (t TTable[Entity, ID]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	t.frame = fn(t.frame)
	return t
}

func (t TTable[Entity, ID]) Frame(frame ui.Frame) ui.DecoredView {
	t.frame = frame
	return t
}

func (t TTable[Entity, ID]) Border(border ui.Border) ui.DecoredView {
	t.border = border
	return t
}

func (t TTable[Entity, ID]) Visible(visible bool) ui.DecoredView {
	t.invisible = !visible
	return t
}

func (t TTable[Entity, ID]) AccessibilityLabel(label string) ui.DecoredView {
	t.accessibilityLabel = label
	return t
}
