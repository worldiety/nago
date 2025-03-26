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

type TMultiSteps struct {
	activeIndex *core.State[int]
	buttonDone  core.View
	steps       []TStep
	canShow     func(currentIdx int, wantedIndex int) bool
	frame       ui.Frame
}

func MultiSteps(steps ...TStep) TMultiSteps {
	return TMultiSteps{steps: steps}
}

func (c TMultiSteps) InputValue(idx *core.State[int]) TMultiSteps {
	c.activeIndex = idx
	return c
}

func (c TMultiSteps) ButtonDone(view core.View) TMultiSteps {
	c.buttonDone = view
	return c
}

func (c TMultiSteps) CanShow(fn func(currentIdx int, wantedIndex int) bool) TMultiSteps {
	c.canShow = fn
	return c
}

func (c TMultiSteps) Frame(frame ui.Frame) TMultiSteps {
	c.frame = frame
	return c
}

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

	return ui.VStack(
		ui.HStack(
			stepper.Stepper(ui.ForEach(c.steps, func(t TStep) stepper.TStep {
				return stepper.Step().Headline(t.headline).SupportingText(t.supportingText)
			})...).Index(c.activeIndex.Get()),
		).FullWidth(),
		ui.VStack().Frame(ui.Frame{Height: ui.L8}), // this is just a separator
		body,
		ui.HLineWithColor(ui.ColorAccent),
		ui.HStack(
			buttons...,
		).Gap(ui.L8).Alignment(ui.Trailing).FullWidth(),
	).Frame(c.frame).Render(ctx)
}

type TStep struct {
	headline       string
	supportingText string
	body           core.View
}

func Step(body core.View) TStep {
	return TStep{
		body: body,
	}
}

func (c TStep) Headline(headline string) TStep {
	c.headline = headline
	return c
}

func (c TStep) SupportingText(supportingText string) TStep {
	c.supportingText = supportingText
	return c
}
