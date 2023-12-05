package ui

import "go.wdy.de/nago/container/slice"

type Stepper struct {
	id            CID
	steps         *SharedList[*StepInfo]
	properties    slice.Slice[Property]
	selectedIndex Int
}

func NewStepper(with func(stepper *Stepper)) *Stepper {
	c := &Stepper{
		id:            nextPtr(),
		steps:         NewSharedList[*StepInfo]("steps"),
		selectedIndex: NewShared[int64]("selectedIndex"),
	}
	c.properties = slice.Of[Property](c.steps, c.selectedIndex)

	if with != nil {
		with(c)
	}

	return c
}

// SelectedIndex marks the step at the given index as active. If index is out of range, nothing is selected.
func (c *Stepper) SelectedIndex() Int {
	return c.selectedIndex
}

func (c *Stepper) Steps() *SharedList[*StepInfo] {
	return c.steps
}

func (c *Stepper) ID() CID {
	return c.id
}

func (c *Stepper) Type() string {
	return "Stepper"
}

func (c *Stepper) Properties() slice.Slice[Property] {
	return c.properties
}

type StepInfo struct {
	id         CID
	number     String // what is in the bubble
	caption    String
	details    String
	properties slice.Slice[Property]
}

func NewStepInfo(with func(step *StepInfo)) *StepInfo {
	c := &StepInfo{
		id:      nextPtr(),
		number:  NewShared[string]("number"),
		caption: NewShared[string]("caption"),
		details: NewShared[string]("details"),
	}

	c.properties = slice.Of[Property](c.number, c.caption, c.details)

	if with != nil {
		with(c)
	}

	return c
}

func (c *StepInfo) ID() CID {
	return c.id
}

func (c *StepInfo) Type() string {
	return "StepInfo"
}

func (c *StepInfo) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *StepInfo) Number() String {
	return c.number
}

func (c *StepInfo) Caption() String {
	return c.caption
}

func (c *StepInfo) Details() String {
	return c.details
}
