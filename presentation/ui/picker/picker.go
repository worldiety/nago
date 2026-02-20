// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package picker

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// TPicker is a composite component (Picker).
// It displays a list of values and lets users choose one or multiple items,
// with optional "Select all" and quick-filtering support. Rendering of both the
// selected summary and the selectable rows is customizable via callbacks.
// The picker can bind to external selection state or manage its own, and it
// can be presented in a dialog with configurable options.
type TPicker[T any] struct {
	renderPicked         func(TPicker[T], []T) core.View // renders the "picked" summary view from current selections
	renderToSelect2      func(wnd core.Window, item T, state *core.State[bool]) core.View
	stringer             func(T) string      // returns a textual representation (used e.g. for quick search/filter)
	label                string              // primary label shown with the control
	supportingText       string              // secondary/helper text shown under the label
	errorText            string              // validation or error message
	values               []T                 // full set of candidate values to pick from
	frame                ui.Frame            // layout container / sizing and spacing context
	title                string              // title used when the picker is presented (e.g., in a dialog)
	selectAllCheckbox    *core.State[bool]   // nil if "Select all" is not applicable/enabled
	targetSelectedState  *core.State[[]T]    // external binding; nil if uncontrolled
	currentSelectedState *core.State[[]T]    // internal working selection (mirrors/buffers target state during edits)
	pickerPresented      *core.State[bool]   // controls whether the picker UI is currently shown
	multiselect          *core.State[bool]   // true for multi-select mode; single-select otherwise
	checkboxStates       []*core.State[bool] // per-item selection states (used in multi-select list)
	quickSearch          *core.State[string] // current quick-filter query; nil if filtering is disabled
	selectAllSupported   bool                // enables/disables "Select all" behavior
	quickFilterSupported bool                // enables/disables quick search/filter UI
	disabled             bool                // when true, interaction is disabled
	detailView           core.View           // optional detail pane shown alongside the list
	invisible            bool                // when true, the control is not rendered
	dlgOptions           []alert.Option      // dialog options used when presenting the picker
}

// Picker takes the given slice and state to represent the selection. Internally, it uses deep equals, to determine
// the unique set of selected elements and coordinate that with the UI state.
func Picker[T any](label string, values []T, selectedState *core.State[[]T]) TPicker[T] {
	c := TPicker[T]{
		selectAllCheckbox: core.DerivedState[bool](selectedState, "pck.all"),
		multiselect:       core.DerivedState[bool](selectedState, ".pck.ms"),
		// warning: do not reference this from constructor, because it may get overridden from outside
		pickerPresented:      core.DerivedState[bool](selectedState, ".pck.pre"),
		quickSearch:          core.DerivedState[string](selectedState, ".pck.qs"),
		label:                label,
		values:               values,
		targetSelectedState:  selectedState,
		currentSelectedState: core.DerivedState[[]T](selectedState, ".pck.tmp"),
		selectAllSupported:   true,
		quickFilterSupported: true,
		checkboxStates:       make([]*core.State[bool], 0, len(values)),
	}

	c.renderPicked = func(p TPicker[T], t []T) core.View {
		textColor := ui.M8
		if p.disabled {
			textColor = ui.ColorIconsMuted
		}

		switch len(t) {
		case 0:
			return ui.Text("nichts gewählt").Color(textColor)
		case 1:
			return ui.Text(cleanMarkdown(fmt.Sprintf("%v", t[0]))).Color(textColor)
		default:
			return ui.Text(fmt.Sprintf("%d gewählt", len(t))).Color(textColor)
		}

	}

	c = c.ItemRenderer(nil)

	c.stringer = func(t T) string {
		return fmt.Sprintf("%v", t)
	}

	for i := range c.values {
		c.checkboxStates = append(c.checkboxStates, core.DerivedState[bool](selectedState, fmt.Sprintf(".pck.cb-%d", i)).
			Observe(func(newValue bool) {
				if !c.multiselect.Get() {
					for i2, state := range c.checkboxStates {
						if i2 != i {
							state.Set(false)
						}
					}
				}

				count := c.syncCurrentSelectedState()

				c.selectAllCheckbox.Set(count == len(c.values))
			}))
	}

	c.selectAllCheckbox.Observe(func(newValue bool) {
		for _, state := range c.checkboxStates {
			state.Set(newValue)
		}

		if !newValue {
			c.currentSelectedState.Set(nil)
		} else {
			c.currentSelectedState.Set(append([]T{}, c.values...))
		}
	})

	// init the checkbox state exactly once
	c.selectAllCheckbox.Init(func() bool {
		c.syncCheckboxStates(c.targetSelectedState)
		return false
	})

	if c.currentSelectedState.Get() == nil {
		// try to optimize state init to lower roundtrip times. we cannot use init, because
		// of logic ordering
		count := c.syncCurrentSelectedState()
		c.selectAllCheckbox.Set(count == len(c.values))
	}

	return c
}

func (c TPicker[T]) syncCurrentSelectedState() (selectedCount int) {
	// better re-allocate, we don't know what the owner does with it and it may break value-semantics otherwise
	selected := make([]T, 0, len(c.values))
	count := 0
	for i2, state := range c.checkboxStates {
		if state.Get() {
			selected = append(selected, c.values[i2])
			count++
		}
	}
	c.currentSelectedState.Set(selected)

	return count
}

func (c TPicker[T]) DialogPresented() *core.State[bool] {
	return c.pickerPresented
}

func (c TPicker[T]) WithDialogPresented(state *core.State[bool]) TPicker[T] {
	c.pickerPresented = state
	return c
}

func (c TPicker[T]) syncCheckboxStates(state *core.State[[]T]) {
	for i, value := range c.values {
		found := false
		equalable, ok := any(value).(interface{ Equals(other any) bool })
		for _, t := range state.Get() {
			if ok {
				if equalable.Equals(t) {
					found = true
					break
				}

			} else {
				if reflect.DeepEqual(value, t) {
					found = true
					break
				}
			}

		}

		c.checkboxStates[i].Set(found)
	}

}

// DetailView is optional and placed between the picker section and the button footer.
func (c TPicker[T]) DetailView(detailView core.View) TPicker[T] {
	c.detailView = detailView
	return c
}

func (c TPicker[T]) Padding(padding ui.Padding) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker[T]) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

func (c TPicker[T]) FullWidth() TPicker[T] {
	c.frame = c.frame.FullWidth()
	return c
}

func (c TPicker[T]) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TPicker[T]) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker[T]) Visible(visible bool) ui.DecoredView {
	c.invisible = !visible
	return c
}

func (c TPicker[T]) AccessibilityLabel(label string) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker[T]) Title(title string) TPicker[T] {
	c.title = title
	return c
}

func (c TPicker[T]) SupportingText(text string) TPicker[T] {
	c.supportingText = text
	return c
}

func (c TPicker[T]) ErrorText(text string) TPicker[T] {
	c.errorText = text
	return c
}

// MultiSelect is by default false.
func (c TPicker[T]) MultiSelect(mv bool) TPicker[T] {
	c.multiselect.Set(mv)
	return c
}

// SelectAllSupported sets the select-all-support and if true and
// multiSelect is enabled, a checkbox to select all is shown. Default is true.
func (c TPicker[T]) SelectAllSupported(flag bool) TPicker[T] {
	c.selectAllSupported = flag
	return c
}

// QuickFilterSupported sets the quick-filter-support and if true and
// values contains more than 10 items, the quick filter is shown. Default is true.
func (c TPicker[T]) QuickFilterSupported(flag bool) TPicker[T] {
	c.quickFilterSupported = flag
	return c
}

// ItemPickedRenderer can be customized to return a non-text view for the given T. This is shown
// within the selected window for the currently selected items.
func (c TPicker[T]) ItemPickedRenderer(fn func([]T) core.View) TPicker[T] {
	c.renderPicked = func(t TPicker[T], ts []T) core.View {
		return fn(ts)
	}
	return c
}

// Deprecated: ItemRenderer can be customized to return a non-text view for the given T. This is
// shown within the picker popup. If fn is nil, the default fallback rendering will be applied.
func (c TPicker[T]) ItemRenderer(fn func(T) core.View) TPicker[T] {
	if fn == nil {
		return c.ItemRenderer2(nil)
	}

	return c.ItemRenderer2(func(wnd core.Window, item T, state *core.State[bool]) core.View {
		return fn(item)
	})
}

// ItemRenderer2 can be customized to return a non-text view for the given T. This is
// shown within the picker popup. If fn is nil, the default fallback rendering will be applied.
func (c TPicker[T]) ItemRenderer2(fn func(wnd core.Window, item T, state *core.State[bool]) core.View) TPicker[T] {
	c.renderToSelect2 = fn
	if fn == nil {
		c.renderToSelect2 = func(wnd core.Window, item T, state *core.State[bool]) core.View {
			text := cleanMarkdown(fmt.Sprintf("%v", item))
			resolve := strings.HasPrefix(text, "@")

			return ui.HStack(ui.Text(text).Resolve(resolve).LabelFor(state.ID())).Padding(ui.Padding{}.Vertical(ui.L16))
		}
	}

	return c
}

func (c TPicker[D]) Disabled(disabled bool) TPicker[D] {
	c.disabled = disabled
	return c
}

func (c TPicker[T]) pickerTable(wnd core.Window) (table core.View, quickFilter core.View) {
	filtered := make([]bool, len(c.values))
	hiddenEntries := 0
	if c.quickSearch.Get() != "" {

		// do not allocate strings, if we don't need the filter at all
		predicate := rquery.SimplePredicate[string](c.quickSearch.Get())

		for idx, value := range c.values {
			if !predicate(c.stringer(value)) {
				hiddenEntries++
				filtered[idx] = true
			}
		}

	}

	var quickSearchHelpText string
	if hiddenEntries > 0 {
		quickSearchHelpText = fmt.Sprintf("%d/%d angezeigt", len(filtered)-hiddenEntries, len(c.values))
	} else {
		quickSearchHelpText = fmt.Sprintf("%d Einträge durchsuchen", len(c.values))
	}

	var noEntries core.View
	if len(c.values) == 0 {
		noEntries = ui.HStack(ui.Text("Es sind noch keine Einträge vorhanden."))
	}
	quickFilterVisible := c.quickFilterSupported && len(c.values) > 10
	selectAllVisible := c.multiselect.Get() && c.selectAllSupported && quickFilterVisible

	quickFilter = ui.HStack(
		ui.Image().
			Embed(heroSolid.MagnifyingGlass).
			Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
		ui.TextField("", c.quickSearch.Get()).
			InputValue(c.quickSearch).
			Style(ui.TextFieldReduced).
			SupportingText(quickSearchHelpText).
			Frame(ui.Frame{}.FullWidth()),
	).Gap(ui.L4).
		Visible(quickFilterVisible).
		Frame(ui.Frame{MinHeight: ui.L64}.FullWidth()).Padding(ui.Padding{Left: ui.L8, Bottom: ui.L16})

	return ui.VStack(

		ui.With(ui.Table(
			ui.TableColumn(ui.Checkbox(c.selectAllCheckbox.Get()).InputChecked(c.selectAllCheckbox).Visible(selectAllVisible)).Width(ui.L20),
			ui.TableColumn(ui.Text("alles auswählen").Visible(selectAllVisible)).Width(ui.L320),
		).CellPadding(ui.Padding{}).
			RowDividerColor(ui.ColorCardBody).
			BackgroundColor(ui.ColorCardBody).Rows(slices.Collect(func(yield func(row ui.TTableRow) bool) {

			for i, value := range filtered {
				if value {
					continue
				}

				state := c.checkboxStates[i]

				if c.multiselect.Get() {
					yield(ui.TableRow(
						ui.TableCell(ui.Checkbox(state.Get()).InputChecked(state).ID(state.ID())),
						ui.TableCell(c.renderToSelect2(wnd, c.values[i], state)).Alignment(ui.Leading),
					),
					)
				} else {
					yield(ui.TableRow(
						ui.TableCell(ui.RadioButton(state.Get()).InputChecked(state).ID(state.ID())),
						ui.TableCell(c.renderToSelect2(wnd, c.values[i], state)).Alignment(ui.Leading),
					),
					)
				}

			}
		})...), func(table ui.TTable) ui.TTable {
			if !quickFilterVisible {
				table = table.HeaderDividerColor("")
				table = table.RowDividerColor("")
			} else {
				table = table.HeaderDividerColor(ui.ColorLine)
			}
			return table
		}),
		noEntries,
		c.detailView,
	).Frame(ui.Frame{}.FullWidth()), quickFilter
}

func (c TPicker[T]) DialogOptions(opts ...alert.Option) TPicker[T] {
	c.dlgOptions = append(c.dlgOptions, opts...)
	return c
}

// Dialog returns the dialog view as if pressed on the actual button.
func (c TPicker[T]) Dialog() core.View {

	var wnd core.Window
	selectedCount := 0
	for _, state := range c.checkboxStates {
		wnd = state.Window()
		if state.Get() {
			selectedCount++
		}
	}

	table, quickFilter := c.pickerTable(wnd)

	dlgOpts := make([]alert.Option, 0, 4+len(c.dlgOptions))
	dlgOpts = append(dlgOpts, alert.PreBody(quickFilter),
		alert.Cancel(func() {
			c.currentSelectedState.Set(c.targetSelectedState.Get())
			c.syncCheckboxStates(c.targetSelectedState)
		}),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			// positive case
			return ui.PrimaryButton(func() {
				c.targetSelectedState.Set(c.currentSelectedState.Get())
				c.targetSelectedState.Notify() // invoke observers
				close(true)
			}).Title(fmt.Sprintf("%d übernehmen", selectedCount))
		}))

	dlgOpts = append(dlgOpts, c.dlgOptions...)

	return alert.Dialog(c.title, table, c.pickerPresented, dlgOpts...)
}

func (c TPicker[T]) Render(ctx core.RenderContext) core.RenderNode {
	// TODO refactor me to use the picker.TButton
	borderColor := ui.Color("")
	backgroundColor := ui.Color("")
	if c.disabled {
		backgroundColor = ""
		borderColor = ui.ColorIconsMuted
	} else {
		borderColor = ui.ColorText
	}

	inner := ui.HStack(
		c.Dialog(),
		c.renderPicked(c, c.targetSelectedState.Get()),
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronRight).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).
		Gap(ui.L8).
		BackgroundColor(backgroundColor).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(borderColor).Width(ui.L1).Radius("0.375rem")).
		Padding(ui.Padding{}.All(ui.L8))

	if !c.disabled {
		inner = inner.(ui.THStack).Action(func() {
			if c.disabled {
				return
			}
			c.pickerPresented.Set(true)
		})

		inner = inner.(ui.THStack).HoveredBorder(ui.Border{}.Color(ui.I0).Width(ui.L1).Radius("0.375rem"))
	}

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.If(c.label != "", ui.Text(c.label).Font(ui.Font{Size: ui.L14, LineHeight: "1.25rem"}).Color(borderColor)),
			ui.HStack(
				ui.Image().StrokeColor(ui.SE0).Embed(heroSolid.XMark).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
				ui.Text(c.label).Font(ui.Font{Size: ui.L16}).Color(ui.SE0),
			),
		),
		inner,
		ui.IfElse(c.errorText == "",
			ui.If(c.supportingText != "", ui.Text(c.supportingText).Font(ui.Font{Size: "0.75rem"}).Color(ui.ST0)),
			ui.Text(c.errorText).Font(ui.Font{Size: "0.75rem"}).Color(ui.SE0),
		),
	).Alignment(ui.Leading).
		Gap(ui.L4).
		Visible(!c.invisible).
		Frame(c.frame).
		Render(ctx)
}

var regexMarkdownLink = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

func cleanMarkdown(str string) string {
	if maybeMarkdown(str) {
		return regexMarkdownLink.ReplaceAllStringFunc(str, func(s string) string {
			start := strings.IndexRune(s, '[')
			end := strings.IndexRune(s, ']')
			return s[start+1 : end]
		})
	}

	return str
}

func maybeMarkdown(str string) bool {
	return strings.IndexRune(str, '[') >= 0 && strings.IndexRune(str, ']') >= 0 && strings.IndexRune(str, '(') >= 0 && strings.IndexRune(str, ')') >= 0
}
