// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package stepper

import (
	"fmt"

	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

type StepperLayout int

const (
	StepperLayoutAuto       StepperLayout = StepperLayout(proto.StepperLayoutAuto)
	StepperLayoutHorizontal StepperLayout = StepperLayout(proto.StepperLayoutHorizontal)
	StepperLayoutVertical   StepperLayout = StepperLayout(proto.StepperLayoutVertical)
	StepperLayoutSimple     StepperLayout = StepperLayout(proto.StepperLayoutSimple)
	StepperLayoutSimpleList StepperLayout = StepperLayout(proto.StepperLayoutSimpleList)
)

// TStepper is a composite component (Stepper).
// It visually represents a sequence of steps in a process, highlighting
// completed, current, and upcoming steps with distinct colors and styles.
// Each step can display a label, and the layout adapts between simple
// or full-sized step representations.
type TStepper struct {
	value             int              // current step index
	inputValue        *core.State[int] // optional external state for controlled component behavior
	layout            StepperLayout    // visual layout of the stepper (e.g., horizontal or vertical)
	steps             []TStep          // list of steps making up the stepper
	simpleTextPattern string           // label format for simplified step display
	numbers           bool             // defines whether to display step numbers
	lines             bool             // defines whether to display lines in simple and simple list layouts
}

func Stepper(steps ...TStep) TStepper {
	stepper := TStepper{
		value:             0,
		layout:            StepperLayoutAuto,
		steps:             steps,
		simpleTextPattern: "Schritt %d von %d",
		numbers:           true,
		lines:             true,
	}

	return stepper
}

func (c TStepper) Value(value int) TStepper {
	c.value = value
	return c
}

func (c TStepper) InputValue(state *core.State[int]) TStepper {
	c.inputValue = state
	return c
}

func (c TStepper) Layout(layout StepperLayout) TStepper {
	c.layout = layout
	return c
}

func (c TStepper) Steps(steps ...TStep) TStepper {
	c.steps = steps
	return c
}

func (c TStepper) SimpleTextPattern(pattern string) TStepper {
	c.simpleTextPattern = pattern
	return c
}

func (c TStepper) Numbers(b bool) TStepper {
	c.numbers = b
	return c
}

func (c TStepper) Lines(b bool) TStepper {
	c.lines = b
	return c
}

type TStep struct {
	headline       string
	supportingText string
}

func Step() TStep {
	return TStep{}
}

func (c TStep) Headline(headline string) TStep {
	c.headline = headline
	return c
}

func (c TStep) SupportingText(text string) TStep {
	c.supportingText = text
	return c
}

func (c TStepper) Render(ctx core.RenderContext) core.RenderNode {
	wnd := ctx.Window()

	value := c.value
	if c.inputValue != nil {
		value = c.inputValue.Get()
	}

	simpleText := ""
	if c.simpleTextPattern != "" {
		simpleText = fmt.Sprintf(c.simpleTextPattern, value+1, len(c.steps))
	}

	layout := c.layout
	if layout == StepperLayoutAuto {
		if wnd.Info().SizeClass < core.SizeClassMedium {
			if simpleText != "" {
				layout = StepperLayoutSimple
			} else {
				layout = StepperLayoutVertical
			}
		} else {
			layout = StepperLayoutHorizontal
		}
	}

	stepper := &proto.Stepper{
		InputValue: c.inputValue.Ptr(),
		Value:      proto.Uint(value),
		Steps:      make(proto.StepperSteps, 0),
		Layout:     proto.StepperLayout(layout),
		SimpleText: proto.Str(simpleText),
		Numbers:    proto.Bool(c.numbers),
		Lines:      proto.Bool(c.lines),
	}

	for _, step := range c.steps {
		stepper.Steps = append(stepper.Steps, proto.StepperStep{
			Title:    proto.Str(step.headline),
			Subtitle: proto.Str(step.supportingText),
		})
	}

	return stepper
}
