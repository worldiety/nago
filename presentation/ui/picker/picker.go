package picker

import (
	"fmt"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"reflect"
	"slices"
)

type TPicker[T any] struct {
	renderPicked         func([]T) core.View
	renderToSelect       func(T) core.View
	stringer             func(T) string
	label                string
	supportingText       string
	errorText            string
	values               []T
	frame                ui.Frame
	title                string
	selectAllCheckbox    *core.State[bool]
	targetSelectedState  *core.State[[]T]
	currentSelectedState *core.State[[]T]
	pickerPresented      *core.State[bool]
	multiselect          *core.State[bool]
	checkboxStates       []*core.State[bool]
	quickSearch          *core.State[string]
	selectAllSupported   bool
	quickFilterSupported bool
	disabled             bool
	detailView           core.View
}

// Picker takes the given slice and state to represent the selection. Internally, it uses deep equals, to determine
// the unique set of selected elements and coordinate that with the UI state.
func Picker[T any](label string, values []T, selectedState *core.State[[]T]) TPicker[T] {
	c := TPicker[T]{
		selectAllCheckbox:    core.DerivedState[bool](selectedState, "pck.all"),
		multiselect:          core.DerivedState[bool](selectedState, ".pck.ms"),
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

	textColor := ui.Color(ui.ST0)

	c.renderPicked = func(t []T) core.View {
		switch len(t) {
		case 0:
			return ui.Text("nichts gewählt").Color(textColor)
		case 1:
			return ui.Text(fmt.Sprintf("%v", t[0])).Color(textColor)
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

func (c TPicker[T]) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker[T]) Visible(visible bool) ui.DecoredView {
	//TODO implement me
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
	c.renderPicked = fn
	return c
}

// ItemRenderer can be customized to return a non-text view for the given T. This is
// shown within the picker popup. If fn is nil, the default fallback rendering will be applied.
func (c TPicker[T]) ItemRenderer(fn func(T) core.View) TPicker[T] {
	c.renderToSelect = fn
	if fn == nil {
		c.renderToSelect = func(t T) core.View {
			return ui.HStack(ui.Text(fmt.Sprintf("%v", t))).Padding(ui.Padding{}.Vertical(ui.L16))
		}
	}

	return c
}

func (c TPicker[D]) Disabled(disabled bool) TPicker[D] {
	c.disabled = disabled
	return c
}

func (c TPicker[T]) pickerTable() (table core.View, quickFilter core.View) {
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
		Frame(ui.Frame{}.FullWidth()).Padding(ui.Padding{Left: ui.L8, Bottom: ui.L16})

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
						ui.TableCell(ui.Checkbox(state.Get()).InputChecked(state)),
						ui.TableCell(c.renderToSelect(c.values[i])).Alignment(ui.Leading),
					),
					)
				} else {
					yield(ui.TableRow(
						ui.TableCell(ui.RadioButton(state.Get()).InputChecked(state)),
						ui.TableCell(c.renderToSelect(c.values[i])).Alignment(ui.Leading),
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

// Dialog returns the dialog view as if pressed on the actual button.
func (c TPicker[T]) Dialog() core.View {

	selectedCount := 0
	for _, state := range c.checkboxStates {
		if state.Get() {
			selectedCount++
		}
	}

	table, quickFilter := c.pickerTable()

	return alert.Dialog(c.title, table, c.pickerPresented, alert.Cancel(func() {
		c.currentSelectedState.Set(c.targetSelectedState.Get())
		c.syncCheckboxStates(c.targetSelectedState)
	}),
		alert.PreBody(quickFilter),
		alert.Custom(func(close func(closeDlg bool)) core.View {
			// positive case
			return ui.PrimaryButton(func() {
				c.targetSelectedState.Set(c.currentSelectedState.Get())
				c.targetSelectedState.Notify() // invoke observers
				close(true)
			}).Title(fmt.Sprintf("%d übernehmen", selectedCount))
		}))
}

func (c TPicker[T]) Render(ctx core.RenderContext) core.RenderNode {
	// TODO refactor me to use the picker.TButton
	colors := core.Colors[ui.Colors](ctx.Window())
	borderColor := ui.Color("")
	backgroundColor := ui.Color("")
	if c.disabled {
		borderColor = ""
		backgroundColor = colors.Disabled
	} else {
		borderColor = colors.I1.WithLuminosity(0.75)
	}

	inner := ui.HStack(
		c.Dialog(),
		c.renderPicked(c.targetSelectedState.Get()),
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronDown).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).Action(func() {
		if c.disabled {
			return
		}
		c.pickerPresented.Set(true)
	}).HoveredBorder(ui.Border{}.Color(borderColor).Width(ui.L1).Radius("0.375rem")).
		Gap(ui.L8).
		BackgroundColor(backgroundColor).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.M8).Width(ui.L1).Radius("0.375rem")).
		Padding(ui.Padding{}.All(ui.L8))

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L14}),
			ui.HStack(
				ui.Image().StrokeColor(ui.SE0).Embed(heroSolid.XMark).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
				ui.Text(c.label).Font(ui.Font{Size: ui.L16}).Color(ui.SE0),
			),
		),
		inner,
		ui.IfElse(c.errorText == "",
			ui.Text(c.supportingText).Font(ui.Font{Size: "0.75rem"}).Color(ui.ST0),
			ui.Text(c.errorText).Font(ui.Font{Size: "0.75rem"}).Color(ui.SE0),
		),
	).Alignment(ui.Leading).
		Gap(ui.L4).
		Frame(c.frame).
		Render(ctx)
}
