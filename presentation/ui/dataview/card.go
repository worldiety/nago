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
	"go.wdy.de/nago/presentation/ui/accordion"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/pager"
)

type FormatHint int

const (
	HintInline FormatHint = iota + 1
	HintInvisible
)

type FieldID string

type CardOptions struct {
	Title FieldID // TitleField identifier (or zero-based index). If empty, the first field is used.
	Hints map[FieldID]FormatHint
}

func (t TDataView[E, ID]) renderCards(ctx core.RenderContext) core.RenderNode {
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

	return ui.VStack(
		t.cardActionBar(wnd, model),
	).
		Append(
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

				opts := t.cardOptions
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

				// card
				return ui.VStack(
					// title
					ui.HStack(
						ui.IfFunc(!t.hideSelection, func() core.View {
							if v := u.Identity(); v == "" {
								panic(fmt.Errorf("an item must not have a zero identity"))
							}

							myState := model.Selections[u.Identity()]
							if myState == nil {
								panic(fmt.Errorf("no selection state for %q", u.Identity()))
							}

							return ui.Checkbox(myState.Get()).InputChecked(myState)
						}),

						ui.Text(titleField.Name).Font(ui.TitleMedium),
						ui.Spacer(),
						titleField.Map(u),
					).FullWidth().BackgroundColor(ui.ColorCardTop).Padding(ui.Padding{}.All(ui.L16)),
					ui.Space(ui.L4),
				).
					// actual fields
					Append(
						ui.HStack(
							// field values
							ui.VStack(
								ui.ForEach(t.data.Fields, func(f Field[E]) core.View {

									if f.ID == titleField.ID {
										return nil
									}

									if opts.Hints[f.ID] == HintInvisible {
										return nil
									}

									if opts.Hints[f.ID] == HintInline {
										return ui.HStack(
											ui.Text(f.Name).Font(ui.TitleMedium),
											ui.Spacer(),
											f.Map(u),
										).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L16))
									}

									return ui.VStack(
										ui.Text(f.Name).Font(ui.TitleMedium),
										f.Map(u),
									).Alignment(ui.Leading).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L16))
								})...,
							).FullWidth(),

							// to right we have an optional chevron
							ui.If(t.addChevronRight, ui.VLine().Frame(ui.Frame{}).Padding(ui.Padding{Left: ui.L16})),
							ui.IfFunc(t.addChevronRight, func() core.View {
								return ui.VStack(ui.ImageIcon(icons.ChevronRight).Frame(ui.Frame{MaxWidth: ui.L32, MinWidth: ui.L32}))
							}),
						).Alignment(ui.Stretch).FullWidth().Action(t.wrappedAction(u)),
					).
					Append(ui.Space(ui.L4)).
					FullWidth().
					Gap(ui.L8).
					Border(ui.Border{}.Radius(ui.L16).Width(ui.L1).Color(ui.ColorCardTop))

				/*
					return ui.TableRow(cells...).Action(func() {
						if t.action != nil {
							t.action(u)
						}
					}).HoveredBackgroundColor(ui.ColorCardFooter)*/

			})...).
		// page footer
		Append(
			ui.HStack(
				ui.Text(model.PageString()),
				ui.Spacer(),
				pager.Pager(model.PageIdx).Count(model.Page.PageCount).Visible(model.HasPages()),
			).FullWidth(),
		).
		FullWidth().
		Gap(ui.L16).
		Render(ctx)
}

func (t TDataView[E, ID]) cardActionBar(wnd core.Window, model pager.Model[E, ID]) core.View {
	if t.newAction == nil {
		return nil
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

	filterToggle := core.StateOf[bool](wnd, model.SelectSubset.ID()+"-filterToggle")

	sortByField := core.StateOf[FieldID](wnd, t.modelOptions.StatePrefix+"-sortByField")
	sortReverse := core.StateOf[bool](wnd, t.modelOptions.StatePrefix+"-sortReverse")

	return ui.VStack(

		ui.HStack(
			t.confirmDialog(wnd, confirmPresented, dlgSpec, selected),

			ui.IfFunc(len(t.selectOptions) > 0 && !t.hideSelection, func() core.View {
				return ui.Menu(
					ui.SecondaryButton(nil).Enabled(len(selected) > 0).PreIcon(icons.Grid),
					ui.MenuGroup(items...),
				)
			}),
			ui.If(len(t.selectOptions) > 0 && !t.hideSelection || t.showSearchbar, ui.VLineWithColor(ui.ColorInputBorder).Frame(ui.Frame{Height: ui.L40})),
			t.newAction,
		).
			FullWidth().
			Gap(ui.L8).
			Alignment(ui.Trailing).
			Padding(ui.Padding{Bottom: ui.L16}),

		ui.HLine().Padding(ui.Padding{}.Vertical(ui.L8)),

		accordion.Accordion(
			ui.HStack(
				ui.ImageIcon(icons.Filter),
				ui.Text(rstring.LabelFilter.Get(wnd)),
			),
			ui.VStack(
				ui.If(t.showSearchbar, ui.TextField("", model.Query.Get()).InputValue(model.Query).Style(ui.TextFieldReduced).Leading(ui.ImageIcon(icons.Search)).FullWidth()),
				ui.HStack(
					ui.TertiaryButton(func() {
						model.UnselectAll()
					}).Title(StrXSelected.Get(wnd, i18n.Int("num", model.SelectionCount))).PostIcon(icons.Close).Visible(model.SelectionCount > 0),
				).Alignment(ui.Leading).FullWidth(),

				ui.HStack(
					ui.ForEach(t.data.Fields, func(f Field[E]) core.View {
						if f.Comparator == nil {
							return nil
						}

						ico := icons.ArrowUpDown
						if f.ID == sortByField.Get() {
							if sortReverse.Get() {
								ico = icons.ArrowDown
							} else {
								ico = icons.ArrowUp
							}
						}

						return ui.SecondaryButton(func() {
							if sortByField.Get() != f.ID {
								sortByField.Set(f.ID)
								sortReverse.Set(false)
							} else {
								sortReverse.Set(!sortReverse.Get())
							}
						}).Title(f.Name).PreIcon(ico)
					})...,
				).Wrap(true).Gap(ui.L8).FullWidth(),
			).FullWidth().Gap(ui.L8),

			filterToggle,
		).FullWidth(),

		ui.HLine().Padding(ui.Padding{}.Vertical(ui.L8)),
	).FullWidth().Gap(ui.L8)
}
