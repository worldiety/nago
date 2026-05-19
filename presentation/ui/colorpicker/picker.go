// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package colorpicker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
)

// TPalettePicker is a composite component(Palette Picker).
// This component allows users to select a color
// from a predefined palette. It is typically used in design tools or
// configuration interfaces where color choices are limited to a fixed set.
type TPalettePicker struct {
	palette            Palette
	paletteCols        int
	frame              ui.Frame
	padding            ui.Padding
	accessibilityLabel string
	border             ui.Border
	visible            bool
	value              ui.Color
	state              *core.State[ui.Color] // is nil, if no interactivity is required
	currentState       *core.State[ui.Color]
	pickerPresented    *core.State[bool] // also nil, if state is nil
	customColorState   *core.State[string]
	title              string
	supportingText     string
	errorText          string
	disabled           bool
	label              string
}

func PalettePicker(label string, p Palette) TPalettePicker {
	return TPalettePicker{label: label, palette: p, paletteCols: 7, padding: ui.Padding{}.All(ui.L8), visible: true}
}

// Value sets the selected value. An empty Color selects none.
func (c TPalettePicker) Value(color ui.Color) TPalettePicker {
	c.value = color
	return c
}

// State attaches the given state to the interaction process of selecting a value.
// A nil state signals read-only.
func (c TPalettePicker) State(state *core.State[ui.Color]) TPalettePicker {
	c.state = state
	if c.state != nil {
		c.pickerPresented = core.DerivedState[bool](state, "palette-picker-presented")
		c.currentState = core.DerivedState[ui.Color](state, "palette-picker-current").Init(func() ui.Color {
			return c.value
		})
	}

	return c
}

func (c TPalettePicker) AccessibilityLabel(label string) ui.DecoredView {
	c.accessibilityLabel = label
	return c
}

func (c TPalettePicker) Padding(padding ui.Padding) ui.DecoredView {
	c.padding = padding
	return c
}

func (c TPalettePicker) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

func (c TPalettePicker) WithFrame(fn func(ui.Frame) ui.Frame) ui.DecoredView {
	c.frame = fn(c.frame)
	return c
}

func (c TPalettePicker) Border(border ui.Border) ui.DecoredView {
	c.border = border
	return c
}

func (c TPalettePicker) Visible(visible bool) ui.DecoredView {
	c.visible = visible
	return c
}
func (c TPalettePicker) Title(title string) TPalettePicker {
	c.title = title
	return c
}

func (c TPalettePicker) SupportingText(text string) TPalettePicker {
	c.supportingText = text
	return c
}

func (c TPalettePicker) ErrorText(text string) TPalettePicker {
	c.errorText = text
	return c
}

func (c TPalettePicker) Disabled(disabled bool) TPalettePicker {
	c.disabled = disabled
	return c
}

func (c TPalettePicker) Render(ctx core.RenderContext) core.RenderNode {
	isCustomColor := c.state.Get() != "" && !slices.Contains(c.palette, c.state.Get())
	c.customColorState = core.AutoState[string](ctx.Window()).Init(func() string {
		if isCustomColor {
			return string(c.state.Get())
		}
		return ""
	}).Observe(func(newValue string) {
		if strings.TrimSpace(newValue) == "" {
			if c.state.Get() == "" {
				c.currentState.Set("")
			}
		}
	})
	colors := core.Colors[ui.Colors](ctx.Window())
	borderColor := ui.Color("")
	backgroundColor := ui.Color("")
	if c.disabled {
		borderColor = ""
		backgroundColor = colors.Disabled
	} else {
		borderColor = option.Must(colors.I1.WithChromaAndTone(16, 75))
	}

	inner := ui.HStack(
		c.Dialog(c.pickerPresented),
		renderColor(c.palette, c.value),
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronRight).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).Action(func() {
		if c.disabled {
			return
		}
		if c.state.Get() != "" && !slices.Contains(c.palette, c.state.Get()) {
			c.currentState.Set(c.state.Get())
			c.customColorState.Set(string(c.state.Get()))
		}
		c.pickerPresented.Set(true)
	}).HoveredBorder(ui.Border{}.Color(borderColor).Width(ui.L1).Radius("0.375rem")).
		FocusedBorder(ui.Border{}.Color(ui.I0).Width(ui.L1)).
		Gap(ui.L8).
		BackgroundColor(backgroundColor).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.M8).Width(ui.L1).Radius("0.375rem")).
		Padding(c.padding)

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
		Visible(c.visible).
		Frame(c.frame).
		Render(ctx)
}

func (c TPalettePicker) pickerTable() core.View {
	return ui.VStack(ui.Each(xiter.Chunks(slices.Values(c.palette), c.paletteCols), func(t []ui.Color) core.View {
		return ui.HStack(ui.Each(slices.Values(t), func(t ui.Color) core.View {

			var borderColor ui.Color
			if c.currentState != nil && t == c.currentState.Get() && c.customColorState.Get() == "" {
				// mark whatever is selected
				borderColor = ui.ColorInputBorder
			} else {
				borderColor = "#00000000"
			}

			return ui.HStack(renderColor(c.palette, t)).Action(func() {
				c.currentState.Set(t)
				c.customColorState.Set("")
			}).
				Padding(ui.Padding{}.All(ui.L2)).
				Border(ui.Border{}.Circle().Color(borderColor).Width(ui.L2))

		})...).Gap(ui.L12)
	})...).Gap(ui.L12).Frame(ui.Frame{}.FullWidth())
}

// Dialog returns the dialog view as if pressed on the actual button.
func (c TPalettePicker) Dialog(pickerPresented *core.State[bool]) core.View {
	return alert.Dialog(
		c.title,
		ui.VStack(
			c.pickerTable(),
			c.customColorView(),
		).Gap(ui.L12).FullWidth(),
		pickerPresented,
		alert.Cancel(func() {
			if c.currentState != nil {
				c.currentState.Set(c.state.Get())
			}
			// if the current color is in palette there is no need to retain the custom color
			if slices.Contains(c.palette, c.currentState.Get()) {
				c.customColorState.Set("")
			}
		}),

		alert.Custom(func(close func(closeDlg bool)) core.View {
			// positive case
			return ui.PrimaryButton(func() {
				if strings.TrimSpace(c.customColorState.Get()) != "" {
					c.currentState.Set(ui.Color(strings.TrimSpace(c.customColorState.Get())))
				} else {
					c.customColorState.Set("")
					if !slices.Contains(c.palette, c.currentState.Get()) {
						// current color is not in palette which means customColor is empty and no color from palette is selected
						c.currentState.Set("")
					}
				}
				c.state.Set(c.currentState.Get())
				c.state.Notify() // invoke observers
				c.customColorState.Notify()
				close(true)
			}).Title(fmt.Sprintf("Farbe wählen"))
		}),
	)
}

// customColorView returns the view where the user can define a custom color with hexcode
func (c TPalettePicker) customColorView() core.View {
	raw := strings.TrimSpace(c.customColorState.Get())
	var valid bool
	if raw != "" && IsValidHexColor(raw) {
		valid = true
	}

	var errorText string
	if !valid && raw != "" {
		errorText = "Ungültiger Hexcode (z.B. #FF0000)"
	}

	var preview core.View
	if valid {
		preview = ui.VStack(
			ui.VStack().
				BackgroundColor(ui.Color(raw)).
				Border(ui.Border{}.Circle()).
				Frame(ui.Frame{}.Size(ui.L24, ui.L24)),
		).
			Padding(ui.Padding{}.All(ui.L1)).
			Border(ui.Border{}.Circle().Color(ui.ColorInputBorder).Width(ui.L2))
	} else {
		preview = ui.ImageIcon(heroOutline.NoSymbol)
	}

	return ui.HStack(
		ui.HStack(preview).Alignment(ui.Leading).Frame(ui.Frame{MaxWidth: ui.L40}.FullWidth()),
		ui.TextField("Eigene Farbe", c.customColorState.Get()).
			InputValue(c.customColorState).
			SupportingText("Hexcode (z. B. #FF0000)").
			ErrorText(errorText).
			FullWidth(),
	).Gap(ui.L12).FullWidth().Padding(ui.Padding{}.Horizontal(ui.L40))
}

// renderColor renders a no-color sign, if c is empty or if c is not in palette. If palette is nil, the color is
// always treated as if it would be in the palette.
func renderColor(palette Palette, c ui.Color) ui.DecoredView {
	//if c == "" || (palette != nil && !slices.Contains(palette, c)) {
	if c == "" || palette == nil {
		return renderNone()
	}

	return ui.VStack().
		BackgroundColor(c).
		Frame(ui.Frame{}.Size(ui.L24, ui.L24)).
		Border(ui.Border{}.Circle())
}

func renderNone() ui.DecoredView {
	return ui.Image().
		Embed(heroOutline.NoSymbol).
		Frame(ui.Frame{}.Size(ui.L24, ui.L24))
}
