// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataview

import (
	"fmt"

	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrXSelected = i18n.MustVarString("nago.dataview.x_selected", i18n.Values{language.English: "{num} selected", language.German: "{num} ausgewählt"})
)

type TableOptions struct {
	ColumnWidths map[FieldID]ui.Length // Field ID -> column width
}

func (t TDataView[E, ID]) renderTable(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()
	data := t.data

	sortByField := core.StateOf[FieldID](wnd, t.modelOptions.StatePrefix+"-sortByField")
	sortReverse := core.StateOf[bool](wnd, t.modelOptions.StatePrefix+"-sortReverse")

	findAllIDs := data.FindAll
	if sortByField.Get() != "" {
		var sortField Field[E]
		for _, field := range t.data.Fields {

			if field.ID == sortByField.Get() {
				sortField = field
				break
			}
		}

		if sortField.Comparator != nil {
			findAllIDs = pager.Sort(data.FindByID, data.FindAll, func(a, b E) int {
				if sortReverse.Get() {
					return -sortField.Comparator(a, b)
				}

				return sortField.Comparator(a, b)
			}, pager.SortOptions[E]{})
		}

	}

	var model pager.Model[E, ID]
	if t.model.IsSome() {
		model = t.model.Unwrap()
	} else {
		m, err := pager.NewModel(
			wnd,
			data.FindByID,
			findAllIDs,
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
		if field.Visible != nil {
			if !field.Visible(FieldContext{Style: Table, Window: wnd}) {
				continue
			}
		}

		if field.Comparator == nil {
			cols = append(cols, ui.TableColumn(ui.Text(field.Name)).Width(t.tableOptions.ColumnWidths[field.ID]))
			continue
		}

		ico := icons.ArrowUpDown
		if field.ID == sortByField.Get() {
			if sortReverse.Get() {
				ico = icons.ArrowDown
			} else {
				ico = icons.ArrowUp
			}
		}

		cols = append(cols, ui.TableColumn(ui.HStack(
			ui.TertiaryButton(func() {
				if sortByField.Get() != field.ID {
					sortByField.Set(field.ID)
					sortReverse.Set(false)
				} else {
					sortReverse.Set(!sortReverse.Get())
				}
			}).PreIcon(ico),
			ui.Text(field.Name),
		)).Width(t.tableOptions.ColumnWidths[field.ID]))
	}

	if t.addChevronRight {
		cols = append(cols, ui.TableColumn(ui.Text("")).Width(ui.L64))
	}

	return ui.VStack(
		t.tableActionBar(ctx.Window(), model),
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
					if field.Visible != nil {
						if !field.Visible(FieldContext{Style: Table, Window: wnd}) {
							continue
						}
					}
					cells = append(cells, ui.TableCell(field.Map(u)))
				}

				if t.addChevronRight {
					cells = append(cells, ui.TableCell(ui.HStack(ui.ImageIcon(icons.ChevronRight))).Alignment(ui.Trailing))
				}

				return ui.TableRow(cells...).Action(t.wrappedAction(u)).HoveredBackgroundColor(ui.ColorCardFooter)
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

func (t TDataView[E, ID]) tableActionBar(wnd core.Window, model pager.Model[E, ID]) core.View {
	if wnd.Info().SizeClass <= core.SizeClassSmall {
		return t.cardActionBar(wnd, model)
	}

	selected := model.Selected()
	dlgSpec := core.StateOf[ConfirmDialog[ID]](wnd, model.SelectSubset.ID()+"-dlgConfirm")
	confirmPresented := core.DerivedState[bool](dlgSpec, "confirm-presented")

	var items []ui.TMenuItem
	for _, selectOption := range t.selectOptions {
		if selectOption.Visible != nil && !selectOption.Visible(selected) {
			continue
		}

		items = append(items, ui.MenuItem(func() {
			if selectOption.ConfirmDialog == nil {
				if err := selectOption.Action(selected); err != nil {
					alert.ShowBannerError(wnd, err)
					return
				}

				return
			}

			spec := selectOption.ConfirmDialog(selected)
			spec.action = selectOption.Action
			dlgSpec.Set(spec)
			confirmPresented.Set(true)

		}, ui.HStack(ui.ImageIcon(selectOption.Icon), ui.Text(selectOption.Name).TextAlignment(ui.TextAlignStart)).Gap(ui.L8).Alignment(ui.Leading).FullWidth()))
	}

	return ui.HStack(
		t.confirmDialog(wnd, confirmPresented, dlgSpec, selected),
		ui.If(t.showSearchbar, ui.TextField("", model.Query.Get()).InputValue(model.Query).Style(ui.TextFieldReduced).Leading(ui.ImageIcon(icons.Search))),

		ui.TertiaryButton(func() {
			model.UnselectAll()
		}).Title(StrXSelected.Get(wnd, i18n.Int("num", model.SelectionCount))).PostIcon(icons.Close).Visible(model.SelectionCount > 0),

		ui.IfFunc(len(t.selectOptions) > 0 && !t.hideSelection, func() core.View {
			return ui.Menu(
				ui.SecondaryButton(nil).Enabled(len(selected) > 0).Title(rstring.LabelOptions.Get(wnd)).PreIcon(icons.Grid),
				ui.MenuGroup(items...),
			)
		}),
		ui.If((len(t.selectOptions) > 0 && !t.hideSelection || t.showSearchbar) && t.newAction != nil, ui.VLineWithColor(ui.ColorInputBorder).Frame(ui.Frame{Height: ui.L40})),
		t.newAction,
	).
		FullWidth().
		Gap(ui.L8).
		Alignment(ui.Trailing).
		Padding(ui.Padding{Bottom: ui.L16})
}
