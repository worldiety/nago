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
	"github.com/worldiety/option"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/list"
	"go.wdy.de/nago/presentation/ui/pager"
)

type ListOptions[ID ~string] struct {
	Title          FieldID // TitleField identifier (or zero-based index). If empty, the first field is used.
	Description    FieldID // Description field identifier (or zero-based index). If empty, the second field is used.
	Hints          map[FieldID]FormatHint
	ColorBody      option.Opt[ui.Color]
	ColorCaption   option.Opt[ui.Color]
	ColorFooter    option.Opt[ui.Color]
	ColorHover     option.Opt[ui.Color]
	ColorHighlight option.Opt[ui.Color]
	Highlight      map[ID]bool // highlight items by ID
}

func (t TDataView[E, ID]) renderList(ctx core.RenderContext) core.RenderNode {
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

	if model.Window == nil {
		model.Window = wnd
	}

	return ui.VStack().
		Append(
			list.List(
				ui.ForEach(model.Page.Items, func(u E) core.View {
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

					if t.addChevronRight {
						cells = append(cells, ui.TableCell(ui.HStack(ui.ImageIcon(icons.ChevronRight))).Alignment(ui.Trailing))
					}

					opts := t.listOptions
					var titleField Field[E]
					for idx, field := range t.data.Fields {
						if opts.Title == "" {
							if idx == 0 {
								titleField = field
								break
							}
						} else if opts.Title == field.ID {
							titleField = field
							break
						}
					}

					var descField Field[E]
					for idx, field := range t.data.Fields {
						if opts.Description == "" {
							if idx == 1 {
								descField = field
								break
							}
						} else if opts.Description == field.ID {
							descField = field
							break
						}
					}

					if descField.Map == nil {
						descField.Map = func(obj E) core.View {
							return nil
						}
					}

					var checkboxView core.View
					if !t.hideSelection {
						if v := u.Identity(); v == "" {
							panic(fmt.Errorf("an item must not have a zero identity"))
						}

						myState := model.Selections[u.Identity()]
						if myState == nil {
							panic(fmt.Errorf("no selection state for %q", u.Identity()))
						}

						checkboxView = ui.Checkbox(myState.Get()).InputValue(myState)
					}

					var trailingView core.View
					if t.addChevronRight {
						trailingView = ui.ImageIcon(icons.ChevronRight).Frame(ui.Frame{MaxWidth: ui.L32, MinWidth: ui.L32})
					}

					return list.Entry().
						Leading(checkboxView).
						HeadlineView(titleField.Map(u)).
						SupportingView(descField.Map(u)).
						Trailing(trailingView)

				})...,
			).OnHighlighted(func(idx int) bool {
				item := model.Page.Items[idx]
				return t.listOptions.Highlight[item.Identity()]
			}).
				Caption(t.listActionBar(wnd, model)).
				Footer(
					// page footer
					ui.HStack(
						ui.Text(model.PageString()),
						ui.Spacer(),
						pager.Pager(model.PageIdx).Count(model.Page.PageCount).Visible(model.HasPages()),
					).FullWidth(),
				).
				With(func(c list.TList) list.TList {
					if t.listOptions.ColorBody.IsSome() {
						c = c.ColorBody(t.listOptions.ColorBody.Unwrap())
					}
					if t.listOptions.ColorCaption.IsSome() {
						c = c.ColorCaption(t.listOptions.ColorCaption.Unwrap())
					}
					if t.listOptions.ColorFooter.IsSome() {
						c = c.ColorFooter(t.listOptions.ColorFooter.Unwrap())
					}

					if t.listOptions.ColorHover.IsSome() {
						c = c.ColorHover(t.listOptions.ColorHover.Unwrap())
					}

					if t.listOptions.ColorHighlight.IsSome() {
						c = c.ColorHighlight(t.listOptions.ColorHighlight.Unwrap())
					}

					if t.action != nil {
						c = c.OnEntryClicked(func(idx int) {
							t.action(model.Page.Items[idx])
						})
					}

					return c
				}).FullWidth(),
		).
		FullWidth().
		Gap(ui.L16).
		Render(ctx)
}

func (t TDataView[E, ID]) listActionBar(wnd core.Window, model pager.Model[E, ID]) core.View {
	selected := model.Selected()
	dlgSpec := core.StateOf[ConfirmDialog[ID]](wnd, model.SelectSubset.ID()+"-dlgConfirm")
	confirmPresented := core.DerivedState[bool](dlgSpec, "confirm-presented")

	var groups []ui.TMenuGroup
	if t.createMenuGroup != nil {
		groups = append(groups, *t.createMenuGroup)
	}

	var items []ui.TMenuItem

	if len(selected) > 0 && !t.hideSelection {
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

	}

	if len(items) > 0 {
		groups = append(groups, ui.MenuGroup(items...))
	}

	var selItems []ui.TMenuItem
	if !t.hideSelection {
		if model.SelectionCount != model.Page.Total {
			selItems = append(selItems, ui.MenuItem(func() {
				model.SelectAll()
			}, ui.Text(StrXSelectAll.Get(wnd))))
		}
	}

	if !t.hideSelection && len(selected) > 0 {
		selItems = append(selItems, ui.MenuItem(func() {
			model.UnselectAll()
		}, ui.Text(StrXDeselect.Get(wnd, i18n.Int("num", model.SelectionCount)))))
	}

	if len(selItems) > 0 {
		groups = append(groups, ui.MenuGroup(selItems...))
	}

	sortByField := core.StateOf[FieldID](wnd, t.modelOptions.StatePrefix+"-sortByField")
	sortReverse := core.StateOf[bool](wnd, t.modelOptions.StatePrefix+"-sortReverse")
	var sortItems []ui.TMenuItem

	for _, f := range t.data.Fields {
		if f.Comparator == nil {
			continue
		}

		ico := icons.ArrowUpDown
		if f.ID == sortByField.Get() {
			if sortReverse.Get() {
				ico = icons.ArrowDown
			} else {
				ico = icons.ArrowUp
			}
		}

		sortItems = append(sortItems, ui.MenuItem(func() {
			if sortByField.Get() != f.ID {
				sortByField.Set(f.ID)
				sortReverse.Set(false)
			} else {
				sortReverse.Set(!sortReverse.Get())
			}

		}, ui.HStack(ui.ImageIcon(ico), ui.Text(f.Name))))

	}

	if len(sortItems) > 0 {
		groups = append(groups, ui.MenuGroup(sortItems...))
	}

	return ui.HStack(
		t.confirmDialog(wnd, confirmPresented, dlgSpec, selected),
		ui.If(t.showSearchbar, ui.TextField("", model.Query.Get()).InputValue(model.Query).Style(ui.TextFieldReduced).Leading(ui.ImageIcon(icons.Search)).FullWidth()),
		ui.If(len(groups) > 0, ui.Menu(
			ui.TertiaryButton(nil).PreIcon(icons.DotsVertical),
			groups...,
		)),
	).FullWidth()

}
