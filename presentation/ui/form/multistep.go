// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"go.wdy.de/nago/application/localization/rstring"
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
	onStepChange func(from, to int) bool
	backLabel    string
	nextLabel    string
	frame        ui.Frame
	colorCurrent ui.Color
	colorDone    ui.Color
	colorFuture  ui.Color
	layout       stepper.StepperLayout
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

// OnStepChange sets a callback which is invoked when the user navigates from one step to another using the back or
// next button. It is called within the event loop and before the step changes. Return false to stay at the current
// step, e.g. because the validation of its inputs failed. A validation which takes longer should disable the
// next button through its own state and [TMultiSteps.CanShow] instead of blocking the event loop.
func (c TMultiSteps) OnStepChange(fn func(from, to int) bool) TMultiSteps {
	c.onStepChange = fn
	return c
}

// BackLabel replaces the localized default label of the back button.
func (c TMultiSteps) BackLabel(label string) TMultiSteps {
	c.backLabel = label
	return c
}

// NextLabel replaces the localized default label of the next button.
func (c TMultiSteps) NextLabel(label string) TMultiSteps {
	c.nextLabel = label
	return c
}

// Frame sets the layout frame of the multi-steps component.
func (c TMultiSteps) Frame(frame ui.Frame) TMultiSteps {
	c.frame = frame
	return c
}

// ColorCurrent sets the color for the currently active step indicator.
//
// Deprecated: has no effect, because the stepper does not support custom colors.
func (c TMultiSteps) ColorCurrent(color ui.Color) TMultiSteps {
	c.colorCurrent = color
	return c
}

// ColorDone sets the color for completed step indicators.
//
// Deprecated: has no effect, because the stepper does not support custom colors.
func (c TMultiSteps) ColorDone(color ui.Color) TMultiSteps {
	c.colorDone = color
	return c
}

// ColorFuture sets the color for upcoming step indicators.
//
// Deprecated: has no effect, because the stepper does not support custom colors.
func (c TMultiSteps) ColorFuture(color ui.Color) TMultiSteps {
	c.colorFuture = color
	return c
}

// Layout sets the layout of the stepper, which is chosen automatically by default.
func (c TMultiSteps) Layout(layout stepper.StepperLayout) TMultiSteps {
	c.layout = layout
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

	wnd := ctx.Window()
	backLabel := c.backLabel
	if backLabel == "" {
		backLabel = rstring.ActionBack.Get(wnd)
	}

	nextLabel := c.nextLabel
	if nextLabel == "" {
		nextLabel = rstring.ActionContinue.Get(wnd)
	}

	// the target step is captured by the render, thus a stale second click of a double click cannot skip a step
	current := c.activeIndex.Get()
	goTo := func(to int) {
		if c.activeIndex.Get() != current {
			return
		}

		if c.onStepChange != nil && !c.onStepChange(current, to) {
			return
		}

		c.activeIndex.Set(to)
		c.activeIndex.Notify()
	}

	var buttons []core.View
	if current > 0 {
		enabled := c.canShow(current, current-1)
		buttons = append(buttons, ui.SecondaryButton(func() {
			goTo(current - 1)
		}).Enabled(enabled).Title(backLabel))
	}

	if current < len(c.steps)-1 {
		enabled := c.canShow(current, current+1)
		buttons = append(buttons, ui.PrimaryButton(func() {
			goTo(current + 1)
		}).Enabled(enabled).Title(nextLabel))
	}

	if c.activeIndex.Get() == len(c.steps)-1 && c.buttonDone != nil {
		buttons = append(buttons, c.buttonDone)
	}

	s := stepper.Stepper(ui.ForEach(c.steps, func(t TStep) stepper.TStep {
		return stepper.Step().Headline(t.headline).SupportingText(t.supportingText)
	})...).Value(c.activeIndex.Get())

	if c.layout != stepper.StepperLayoutAuto {
		s = s.Layout(c.layout)
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
