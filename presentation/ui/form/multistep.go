// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/stepper"
)

// TMultiSteps is a composite component (Multi Steps).
// This component manages and displays a sequence of steps,
// tracking the active step index, available steps, and a completion button.
// It can also apply custom logic to determine if a step can be shown.
type TMultiSteps struct {
	activeIndex  *core.State[int]
	buttonDone   core.View
	steps        []TStep
	canShow      func(currentIdx int, wantedIndex int) bool
	frame        ui.Frame
	colorCurrent ui.Color
	colorDone    ui.Color
	colorFuture  ui.Color
	style        stepper.Style
}

// MultiSteps creates a new TMultiSteps with the provided steps.
func MultiSteps(steps ...TStep) TMultiSteps {
	return TMultiSteps{steps: steps}
}

// InputValue binds the active step index state to the multi-steps component.
func (c TMultiSteps) InputValue(idx *core.State[int]) TMultiSteps {
	c.activeIndex = idx
	return c
}

// ButtonDone sets the view to display when the steps are completed.
func (c TMultiSteps) ButtonDone(view core.View) TMultiSteps {
	c.buttonDone = view
	return c
}

// CanShow sets a predicate to control whether a given step can be shown.
func (c TMultiSteps) CanShow(fn func(currentIdx int, wantedIndex int) bool) TMultiSteps {
	c.canShow = fn
	return c
}

// Frame sets the layout frame of the multi-steps component.
func (c TMultiSteps) Frame(frame ui.Frame) TMultiSteps {
	c.frame = frame
	return c
}

// ColorCurrent sets the color for the currently active step indicator.
func (c TMultiSteps) ColorCurrent(color ui.Color) TMultiSteps {
	c.colorCurrent = color
	return c
}

// ColorDone sets the color for completed step indicators.
func (c TMultiSteps) ColorDone(color ui.Color) TMultiSteps {
	c.colorDone = color
	return c
}

// ColorFuture sets the color for upcoming step indicators.
func (c TMultiSteps) ColorFuture(color ui.Color) TMultiSteps {
	c.colorFuture = color
	return c
}

// Style sets the display style of the stepper (Auto, Simple or Full).
func (c TMultiSteps) Style(style stepper.Style) TMultiSteps {
	c.style = style
	return c
}

// Render shows the current step with a stepper and nav buttons; clamps the index and respects CanShow.
func (c TMultiSteps) Render(ctx core.RenderContext) core.RenderNode {
	if c.activeIndex == nil {
		c.activeIndex = core.AutoState[int](ctx.Window())
	}

	if c.activeIndex.Get() < 0 {
		c.activeIndex.Set(0)
		c.activeIndex.Notify()
	}

	if c.activeIndex.Get() >= len(c.steps) {
		c.activeIndex.Set(len(c.steps) - 1)
		c.activeIndex.Notify()
	}

	if c.canShow == nil {
		c.canShow = func(currentIdx int, wantedIndex int) bool {
			return true
		}
	}

	var body core.View
	if len(c.steps) > 0 {
		body = c.steps[c.activeIndex.Get()].body
	}

	var buttons []core.View
	if c.activeIndex.Get() > 0 {
		enabled := c.canShow(c.activeIndex.Get(), c.activeIndex.Get()-1)
		buttons = append(buttons, ui.SecondaryButton(func() {
			c.activeIndex.Set(c.activeIndex.Get() - 1)
			c.activeIndex.Notify()
		}).Enabled(enabled).Title("Zurück"))
	}

	if c.activeIndex.Get() < len(c.steps)-1 {
		enabled := c.canShow(c.activeIndex.Get(), c.activeIndex.Get()+1)
		buttons = append(buttons, ui.PrimaryButton(func() {
			c.activeIndex.Set(c.activeIndex.Get() + 1)
			c.activeIndex.Notify()
		}).Enabled(enabled).Title("Weiter"))
	}

	if c.activeIndex.Get() == len(c.steps)-1 && c.buttonDone != nil {
		buttons = append(buttons, c.buttonDone)
	}

	s := stepper.Stepper(ui.ForEach(c.steps, func(t TStep) stepper.TStep {
		return stepper.Step().Headline(t.headline).SupportingText(t.supportingText)
	})...).Index(c.activeIndex.Get())

	if c.colorCurrent != "" {
		s = s.ColorCurrent(c.colorCurrent)
	}

	if c.colorDone != "" {
		s = s.ColorDone(c.colorDone)
	}

	if c.colorFuture != "" {
		s = s.ColorFuture(c.colorFuture)
	}

	if c.style != stepper.Auto {
		s = s.Style(c.style)
	}

	return ui.VStack(
		ui.HStack(
			s,
		).FullWidth(),
		ui.VStack().Frame(ui.Frame{Height: ui.L8}), // this is just a separator
		body,
		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			buttons...,
		).Gap(ui.L8).Alignment(ui.Trailing).FullWidth(),
	).Frame(c.frame).Render(ctx)
}

// TStep is a basic component (Step).
// Each step contains a body view with optional headline and supporting text.
type TStep struct {
	headline       string    // title of the step
	supportingText string    // descriptive text for the step
	body           core.View // main content of the step
}

// Step creates a new TStep with the given body view.
func Step(body core.View) TStep {
	return TStep{
		body: body,
	}
}

// Headline sets the headline text of the step.
func (c TStep) Headline(headline string) TStep {
	c.headline = headline
	return c
}

// SupportingText sets the supporting text of the step.
func (c TStep) SupportingText(supportingText string) TStep {
	c.supportingText = supportingText
	return c
}
