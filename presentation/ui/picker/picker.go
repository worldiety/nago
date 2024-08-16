package picker

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"slices"
)

type TPicker[T any] struct {
	renderPicked      func([]T) core.View
	renderToSelect    func(T) core.View
	label             string
	supportingText    string
	errorText         string
	values            []T
	frame             ui.Frame
	title             string
	selectAllCheckbox *core.State[bool]
	selectedState     *core.State[[]T]
	pickerPresented   *core.State[bool]
	multiselect       *core.State[bool]
	checkboxStates    []*core.State[bool]
}

func Picker[T any](label string, values []T, selectedState *core.State[[]T]) TPicker[T] {
	c := TPicker[T]{
		selectAllCheckbox: core.DerivedState[bool](selectedState, "pck.all"),
		multiselect:       core.DerivedState[bool](selectedState, ".pck.ms"),
		pickerPresented:   core.DerivedState[bool](selectedState, ".pck.pre"),
		label:             label,
		values:            values,
		selectedState:     selectedState,
		checkboxStates:    make([]*core.State[bool], 0, len(values)),
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

				// better re-allocate, we don't know what the owner does with it and it may break value-semantics otherwise
				selected := make([]T, 0, len(c.values))
				count := 0
				for i2, state := range c.checkboxStates {
					if state.Get() {
						selected = append(selected, c.values[i2])
						count++
					}
				}
				c.selectedState.Set(selected)

				c.selectAllCheckbox.Set(count == len(c.values))
			}))
	}

	c.selectAllCheckbox.Observe(func(newValue bool) {
		for _, state := range c.checkboxStates {
			state.Set(newValue)
		}

		if !newValue {
			c.selectedState.Set(nil)
		} else {
			c.selectedState.Set(append([]T{}, c.values...))
		}
	})

	return c
}

func (c TPicker[T]) Title(title string) TPicker[T] {
	c.title = title
	return c
}

func (c TPicker[T]) Frame(frame ui.Frame) TPicker[T] {
	c.frame = frame
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

func (c TPicker[T]) SingleSelect() TPicker[T] {
	c.multiselect.Set(false)
	return c
}

func (c TPicker[T]) MultiSelect() TPicker[T] {
	c.multiselect.Set(true)
	return c
}

func (c TPicker[T]) pickerTable() core.View {
	return ui.VStack(
		ui.HStack(
			ui.Image().
				Embed(heroSolid.MagnifyingGlass).
				Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
			ui.TextField("", "").
				Style(ui.TextFieldReduced).
				SupportingText(fmt.Sprintf("%d Einträge durchsuchen", len(c.values))).
				Frame(ui.Frame{}.FullWidth()),
		).Gap(ui.L4).
			Frame(ui.Frame{}.FullWidth()).Padding(ui.Padding{Left: ui.L8, Bottom: ui.L16}),

		ui.Table(
			ui.TableColumn(ui.Checkbox(c.selectAllCheckbox.Get()).InputChecked(c.selectAllCheckbox).Visible(c.multiselect.Get())).Width(ui.L20),
			ui.TableColumn(ui.Text("alles auswählen").Visible(c.multiselect.Get())).Width(ui.L320),
		).CellPadding(ui.Padding{}).
			BackgroundColor(ui.M1).Rows(slices.Collect(func(yield func(row ui.TTableRow) bool) {
			for i, value := range c.values {
				state := c.checkboxStates[i]
				yield(ui.TableRow(
					ui.TableCell(ui.Checkbox(state.Get()).InputChecked(state)),
					ui.TableCell(c.renderToSelect(value))),
				)
			}
		})...),
	)
}

func (c TPicker[T]) Render(ctx core.RenderContext) core.RenderNode {
	colors := core.Colors[ui.Colors](ctx.Window())
	if c.renderPicked == nil {
		c.renderPicked = func(t []T) core.View {

			return ui.Text(fmt.Sprintf("%v", t))
		}
	}

	if c.renderToSelect == nil {
		c.renderToSelect = func(t T) core.View {
			return ui.HStack(ui.Text(fmt.Sprintf("%v", t))).Padding(ui.Padding{}.Vertical(ui.L16))
		}
	}

	selectedCount := 0
	for _, state := range c.checkboxStates {
		if state.Get() {
			selectedCount++
		}
	}

	inner := ui.HStack(
		alert.Dialog(c.title, c.pickerTable(), c.pickerPresented, alert.Cancel(nil), alert.Custom(func(close func(closeDlg bool)) core.View {
			return ui.PrimaryButton(func() {
				close(true)
			}).Title(fmt.Sprintf("%d auswählen", selectedCount))
		})),
		c.renderPicked(c.selectedState.Get()),
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronDown).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).Action(func() {
		c.pickerPresented.Set(true)
	}).HoveredBorder(ui.Border{}.Color(colors.I1.WithBrightness(75)).Width(ui.L2).Radius(ui.L8)).
		Gap(ui.L8).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.I1).Width(ui.L2).Radius(ui.L8)).
		Padding(ui.Padding{}.All(ui.L8))

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L16}),
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
