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
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
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
	data          Data[E, ID]
	action        func(e E)
	actionNew     func()
	modelOptions  pager.ModelOptions
	model         option.Opt[pager.Model[E, ID]]
	hideSelection bool
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

// Action sets a global listener which is triggered, if an entry has been clicked.
func (t TDataView[E, ID]) Action(fn func(e E)) TDataView[E, ID] {
	t.action = fn
	return t
}

// ActionNew inserts a conventional default button to create a new element for this data view. It has a default
// position, icon, text and button style.
func (t TDataView[E, ID]) ActionNew(fn func()) TDataView[E, ID] {
	t.actionNew = fn
	return t
}

// Selection enabled the flag to show or hide data selection. Default is show selection.
func (t TDataView[E, ID]) Selection(showSelection bool) TDataView[E, ID] {
	t.hideSelection = !showSelection
	return t
}

// ModelOptions sets the internal model options used to render directly.
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
	if !t.hideSelection {
		cols = append(cols, ui.TableColumn(ui.Checkbox(model.SelectSubset.Get()).InputChecked(model.SelectSubset)).Width(ui.L64))
	}

	for _, field := range data.Fields {
		cols = append(cols, ui.TableColumn(ui.Text(field.Name)))
	}

	return ui.VStack(
		t.actionBar(ctx.Window()),
		ui.Table(cols...).Rows(
			ui.ForEach(model.Page.Items, func(u E) ui.TTableRow {
				if v := u.Identity(); v == "" {
					panic(fmt.Errorf("an item must not have a zero identity"))
				}

				myState := model.Selections[u.Identity()]
				if myState == nil {
					panic(fmt.Errorf("no selection state for %q", u.Identity()))
				}

				var cells []ui.TTableCell
				if !t.hideSelection {
					cells = append(cells, ui.TableCell(ui.Checkbox(myState.Get()).InputChecked(myState)))
				}
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
				).ColSpan(len(cols)),
			).BackgroundColor(ui.ColorCardFooter),
		).Frame(ui.Frame{}.FullWidth()),
	).FullWidth().Render(ctx)
}

func (t TDataView[E, ID]) actionBar(wnd core.Window) core.View {
	if t.actionNew == nil {
		return nil
	}

	return ui.HStack(
		ui.If(t.actionNew != nil, ui.PrimaryButton(t.actionNew).PreIcon(icons.Plus).Title(rstring.ActionNew.Get(wnd))),
	).
		FullWidth().
		Alignment(ui.Trailing).
		Padding(ui.Padding{Bottom: ui.L16})
}
