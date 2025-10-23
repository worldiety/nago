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
	"strconv"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/localization/rstring"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	icons "go.wdy.de/nago/presentation/icons/flowbite/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/pager"
	"golang.org/x/text/language"
)

var (
	StrDialogDeleteTitle = i18n.MustString("nago.dataview.dialog.delete_title", i18n.Values{language.English: "Deletes Elements", language.German: "Elemente löschen"})
	StrDialogDeleteDescX = i18n.MustQuantityString(
		"nago.dataview.dialog.delete_desc",
		i18n.QValues{
			language.English: i18n.Quantities{One: "Delete the selected element?", Other: "Delete {amount} elements?"},
			language.German:  i18n.Quantities{One: "Soll das ausgewählte Element unwiderruflich gelöscht werden?", Other: "Sollen {amount} Elemente unwiderruflich gelöscht werden?"},
		},
	)
)

type ConfirmDialog[ID any] struct {
	Title   string
	Message string
	action  func(selected []ID) error // func of SelectOption.Action
}

type SelectOption[ID any] struct {
	Icon          core.SVG
	Name          string
	Action        func(selected []ID) error
	Visible       func(selected []ID) bool
	ConfirmDialog func(selected []ID) ConfirmDialog[ID]
}

// NewSelectOptionDelete is a factory to create a generic delete option with a standard confirm dialog.
func NewSelectOptionDelete[ID any](wnd core.Window, action func(selected []ID) error) SelectOption[ID] {
	return SelectOption[ID]{
		Icon:   icons.TrashBin,
		Name:   StrDialogDeleteTitle.Get(wnd),
		Action: action,
		ConfirmDialog: func(selected []ID) ConfirmDialog[ID] {
			return ConfirmDialog[ID]{
				Title:   rstring.ActionDelete.Get(wnd),
				Message: StrDialogDeleteDescX.Get(wnd, float64(len(selected)), i18n.Int("amount", len(selected))),
			}
		},
	}
}

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
	data            Data[E, ID]
	action          func(e E)
	actionNew       func()
	selectOptions   []SelectOption[ID]
	modelOptions    pager.ModelOptions
	model           option.Opt[pager.Model[E, ID]]
	hideSelection   bool
	addChevronRight bool
}

type Idx string

func (i Idx) Int() (int, bool) {
	v, err := strconv.Atoi(string(i))
	return v, err == nil
}

type Element[T any] struct {
	ID    Idx
	Index int
	Value T
}

func (e Element[T]) Identity() Idx {
	return e.ID
}

// FromSlice takes the given slice of T and wraps it into an index Element box, so that the identity
// is based on the index within the slice. This works for any kind of T.
func FromSlice[T any](wnd core.Window, slice []T, fields []Field[Element[T]]) TDataView[Element[T], Idx] {
	idents := make([]Idx, 0, len(slice))
	for i := range slice {
		idents = append(idents, Idx(strconv.Itoa(i)))
	}

	return FromData[Element[T], Idx](wnd, Data[Element[T], Idx]{
		FindAll: xslices.ValuesWithError(idents, nil),
		FindByID: func(id Idx) (option.Opt[Element[T]], error) {
			i, err := strconv.Atoi(string(id))
			if err != nil {
				return option.Opt[Element[T]]{}, err
			}

			return option.Some(Element[T]{
				ID:    id,
				Index: i,
				Value: slice[i],
			}), nil
		},
		Fields: fields,
	})
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
// This automatically sets [TDataView.NextActionIndicator] to true.
func (t TDataView[E, ID]) Action(fn func(e E)) TDataView[E, ID] {
	t.action = fn
	t.addChevronRight = true
	return t
}

// NextActionIndicator sets a flag if another column should be appended to indicate
// that an entry has an attached next-action.
func (t TDataView[E, ID]) NextActionIndicator(b bool) TDataView[E, ID] {
	t.addChevronRight = b
	return t
}

// ActionNew inserts a conventional default button to create a new element for this data view. It has a default
// position, icon, text and button style.
func (t TDataView[E, ID]) ActionNew(fn func()) TDataView[E, ID] {
	t.actionNew = fn
	return t
}

// Selection enabled the flag to show or hide data selection. Default is show selection.
// See also [TDataView.SelectOptions].
func (t TDataView[E, ID]) Selection(showSelection bool) TDataView[E, ID] {
	t.hideSelection = !showSelection
	return t
}

// SelectOptions adds a default options button, which is enabled if at least a single item is selected.
// See also [NewSelectOptionDelete].
func (t TDataView[E, ID]) SelectOptions(options ...SelectOption[ID]) TDataView[E, ID] {
	t.selectOptions = options
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

	if t.addChevronRight {
		cols = append(cols, ui.TableColumn(ui.Text("")))
	}

	return ui.VStack(
		t.actionBar(ctx.Window(), model),
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

				if t.addChevronRight {
					cells = append(cells, ui.TableCell(ui.HStack(ui.ImageIcon(icons.ChevronRight))).Alignment(ui.Trailing))
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

func (t TDataView[E, ID]) actionBar(wnd core.Window, model pager.Model[E, ID]) core.View {
	if t.actionNew == nil {
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

		}, ui.HStack(ui.ImageIcon(selectOption.Icon), ui.Text(selectOption.Name)).Gap(ui.L8)))
	}

	return ui.HStack(
		t.confirmDialog(wnd, confirmPresented, dlgSpec, selected),
		ui.IfFunc(len(t.selectOptions) > 0, func() core.View {
			return ui.Menu(
				ui.SecondaryButton(nil).Enabled(len(selected) > 0).Title(rstring.LabelOptions.Get(wnd)).PreIcon(icons.Grid),
				ui.MenuGroup(items...),
			)
		}),
		ui.If(len(t.selectOptions) > 0, ui.VLineWithColor(ui.ColorInputBorder).Frame(ui.Frame{Height: ui.L40})),
		ui.If(t.actionNew != nil, ui.PrimaryButton(t.actionNew).PreIcon(icons.Plus).Title(rstring.ActionNew.Get(wnd))),
	).
		FullWidth().
		Gap(ui.L8).
		Alignment(ui.Trailing).
		Padding(ui.Padding{Bottom: ui.L16})
}

func (t TDataView[E, ID]) confirmDialog(wnd core.Window, presented *core.State[bool], spec *core.State[ConfirmDialog[ID]], selected []ID) core.View {
	if !presented.Get() {
		return nil
	}

	return alert.Dialog(spec.Get().Title, ui.Text(spec.Get().Message), presented, alert.Closeable(), alert.Cancel(nil), alert.Confirm(func() (close bool) {
		if err := spec.Get().action(selected); err != nil {
			alert.ShowBannerError(wnd, err)
			return false
		}

		return true
	}))
}
