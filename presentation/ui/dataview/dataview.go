// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dataview

import (
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

type Style int

const (
	Auto Style = iota
	Table
	Card
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
	ID FieldID // ID may is optional and may be empty. You can set this to apply other effects like default enabled sorting.

	Name string
	Map  func(obj E) core.View

	// Comparator may be nil to disable field sorting. If not-nil this will increase at least the temporary
	// amount of memory to O(n) to apply the comparator on the entire data set. If you can provide a natural
	// identifier order which would be fine for most use cases, you should definitely rely on it for performance.
	Comparator func(a, b E) int
}
type TDataView[E data.Aggregate[ID], ID ~string] struct {
	data            Data[E, ID]
	action          func(e E)
	newAction       core.View
	selectOptions   []SelectOption[ID]
	modelOptions    pager.ModelOptions
	model           option.Opt[pager.Model[E, ID]]
	hideSelection   bool
	showSearchbar   bool
	addChevronRight bool
	wnd             core.Window
	style           Style
	cardOptions     CardOptions
	tableOptions    TableOptions
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
	t := TDataView[E, ID]{
		data: data,
		wnd:  wnd,
	}

	initView(&t)
	return t
}

func FromModel[E data.Aggregate[ID], ID ~string](wnd core.Window, model pager.Model[E, ID], fields []Field[E]) TDataView[E, ID] {
	t := TDataView[E, ID]{
		model: option.Some(model),
		wnd:   wnd,
		data: Data[E, ID]{
			Fields: fields,
		},
	}

	initView(&t)
	return t
}

func initView[E data.Aggregate[ID], ID ~string](t *TDataView[E, ID]) {
	for i, field := range t.data.Fields {
		if field.ID == "" {
			t.data.Fields[i].ID = FieldID(strconv.Itoa(i))
		}
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

// NewAction inserts a conventional default button to create a new element for this data view. It has a default
// position, icon, text and button style. See also [TDataView.NewActionView].
func (t TDataView[E, ID]) NewAction(fn func()) TDataView[E, ID] {
	return t.NewActionView(ui.PrimaryButton(fn).PreIcon(icons.Plus).Title(rstring.ActionNew.Get(t.wnd)))
}

// NewActionView inserts the given view to idiomatic position for creating new elements. See also [TDataView.NewAction].
func (t TDataView[E, ID]) NewActionView(view core.View) TDataView[E, ID] {
	t.newAction = view
	return t
}

// Selection enabled the flag to show or hide data selection. Default is show selection.
// See also [TDataView.SelectOptions].
func (t TDataView[E, ID]) Selection(showSelection bool) TDataView[E, ID] {
	t.hideSelection = !showSelection
	return t
}

func (t TDataView[E, ID]) Style(style Style) TDataView[E, ID] {
	t.style = style
	return t
}

func (t TDataView[E, ID]) CardOptions(cardOptions CardOptions) TDataView[E, ID] {
	t.cardOptions = cardOptions
	return t
}

func (t TDataView[E, ID]) TableOptions(tableOptions TableOptions) TDataView[E, ID] {
	t.tableOptions = tableOptions
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

func (t TDataView[E, ID]) Search(visible bool) TDataView[E, ID] {
	t.showSearchbar = visible
	return t
}

func (t TDataView[E, ID]) Render(ctx core.RenderContext) core.RenderNode {
	var style Style
	if t.style == Auto {
		if len(t.data.Fields) < 3 {
			style = Table
		} else {
			if ctx.Window().Info().SizeClass <= core.SizeClassSmall {
				style = Card
			}
		}

	} else {
		style = t.style
	}

	switch style {
	case Card:
		return t.renderCards(ctx)
	default:
		return t.renderTable(ctx)
	}

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

func (t TDataView[E, ID]) wrappedAction(e E) func() {
	if t.action == nil {
		return nil
	}

	return func() {
		t.action(e)
	}
}
