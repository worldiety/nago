// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataview

import (
	"fmt"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/pager"
)

type Data[E data.Aggregate[ID], ID ~string] struct {
	FindAll  iter.Seq2[ID, error]
	FindByID data.ByIDFinder[E, ID]
	Fields   []Field[E]
}

type Field[E any] struct {
	Name string
	Map  func(obj E) core.View
}
type TDataView[E data.Aggregate[ID], ID ~string] struct {
	data         Data[E, ID]
	action       func(e E)
	modelOptions pager.ModelOptions
	model        option.Opt[pager.Model[E, ID]]
}

func FromData[E data.Aggregate[ID], ID ~string](wnd core.Window, data Data[E, ID]) TDataView[E, ID] {
	return TDataView[E, ID]{
		data: data,
	}
}

func FromModel[E data.Aggregate[ID], ID ~string](wnd core.Window, model pager.Model[E, ID], fields []Field[E]) TDataView[E, ID] {
	return TDataView[E, ID]{
		model: option.Some(model),
		data: Data[E, ID]{
			Fields: fields,
		},
	}
}

func (t TDataView[E, ID]) Action(fn func(e E)) TDataView[E, ID] {
	t.action = fn
	return t
}

func (t TDataView[E, ID]) ModelOptions(opts pager.ModelOptions) TDataView[E, ID] {
	t.modelOptions = opts
	return t
}

func (t TDataView[E, ID]) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	data := t.data

	var model pager.Model[E, ID]
	if t.model.IsSome() {
		model = t.model.Unwrap()
	} else {
		m, err := pager.NewModel(
			wnd,
			data.FindByID,
			data.FindAll,
			t.modelOptions,
		)

		if err != nil {
			return alert.BannerError(err).Render(ctx)
		}

		model = m
	}

	var cols []ui.TTableColumn
	cols = append(cols, ui.TableColumn(ui.Checkbox(model.SelectSubset.Get()).InputChecked(model.SelectSubset)).Width(ui.L64))
	for _, field := range data.Fields {
		cols = append(cols, ui.TableColumn(ui.Text(field.Name)))
	}

	return ui.Table(cols...).Rows(
		ui.ForEach(model.Page.Items, func(u E) ui.TTableRow {
			if v := u.Identity(); v == "" {
				panic(fmt.Errorf("an item must not have a zero identity"))
			}

			myState := model.Selections[u.Identity()]
			if myState == nil {
				panic(fmt.Errorf("no selection state for %q", u.Identity()))
			}

			var cells []ui.TTableCell
			cells = append(cells, ui.TableCell(ui.Checkbox(myState.Get()).InputChecked(myState)))
			for _, field := range data.Fields {
				cells = append(cells, ui.TableCell(field.Map(u)))
			}

			return ui.TableRow(cells...).Action(func() {
				if t.action != nil {
					t.action(u)
				}
			}).HoveredBackgroundColor(ui.ColorCardFooter)
		})...,
	).Rows(
		ui.TableRow(
			ui.TableCell(
				ui.HStack(
					ui.Text(model.PageString()),
					ui.Spacer(),
					pager.Pager(model.PageIdx).Count(model.Page.PageCount).Visible(model.HasPages()),
				).FullWidth(),
			).ColSpan(6),
		).BackgroundColor(ui.ColorCardFooter),
	).
		Frame(ui.Frame{}.FullWidth()).Render(ctx)
}
