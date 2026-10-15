// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	"go.wdy.de/nago/presentation/ui/stepper"
)

// StepperLayout is the layout of a stepper, like stepper.StepperLayout of nago.
type StepperLayout int32

const (
	StepperLayoutAuto       StepperLayout = 0
	StepperLayoutHorizontal StepperLayout = 1
	StepperLayoutVertical   StepperLayout = 2
	StepperLayoutSimple     StepperLayout = 3
	StepperLayoutSimpleList StepperLayout = 4
)

// TStepper is a stepper, which shows the steps of a process, see Stepper.
type TStepper struct{ ref js.Ref }

func (TStepper) isView() {}

// TStepperStep is a step of a stepper, see StepperStep.
type TStepperStep struct{ ref js.Ref }

// Stepper returns a stepper of the steps, like stepper.Stepper of nago.
//
//gostrict:host
func Stepper(steps ...TStepperStep) TStepper {
	return TStepper{ref: wrap(stepper.Stepper(stepperStepsOf(steps)...))}
}

// Value sets the index of the current step.
//
//gostrict:host
func (c TStepper) Value(value int32) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).Value(int(value)))}
}

// Layout sets a fixed layout for the stepper.
//
//gostrict:host
func (c TStepper) Layout(layout StepperLayout) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).Layout(stepper.StepperLayout(layout)))}
}

// Steps sets the steps of the stepper, which overwrites the previous ones.
//
//gostrict:host
func (c TStepper) Steps(steps ...TStepperStep) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).Steps(stepperStepsOf(steps)...))}
}

// SimpleTextPattern overwrites the text pattern of the simple layout. Use %d for the current step and for the total
// number of steps.
//
//gostrict:host
func (c TStepper) SimpleTextPattern(pattern string) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).SimpleTextPattern(pattern))}
}

// CompletedTextPattern overwrites the text pattern of the completed stepper. Use %d for the total number of steps.
//
//gostrict:host
func (c TStepper) CompletedTextPattern(pattern string) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).CompletedTextPattern(pattern))}
}

// Numbers shows or hides the step numbers.
//
//gostrict:host
func (c TStepper) Numbers(b bool) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).Numbers(b))}
}

// Lines shows or hides the lines of the simple layouts.
//
//gostrict:host
func (c TStepper) Lines(b bool) TStepper {
	return TStepper{ref: wrap(get[stepper.TStepper](c.ref).Lines(b))}
}

// StepperStep returns an empty step of a stepper, like stepper.Step of nago.
//
//gostrict:host
func StepperStep() TStepperStep {
	return TStepperStep{ref: wrap(stepper.Step())}
}

// Headline sets the headline of the step.
//
//gostrict:host
func (c TStepperStep) Headline(headline string) TStepperStep {
	return TStepperStep{ref: wrap(get[stepper.TStep](c.ref).Headline(headline))}
}

// SupportingText sets the supporting text of the step.
//
//gostrict:host
func (c TStepperStep) SupportingText(text string) TStepperStep {
	return TStepperStep{ref: wrap(get[stepper.TStep](c.ref).SupportingText(text))}
}
