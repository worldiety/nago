package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Stepper struct {
	id            ora.Ptr
	steps         *SharedList[*StepInfo]
	properties    []core.Property
	selectedIndex Int
}

func NewStepper(with func(stepper *Stepper)) *Stepper {
	c := &Stepper{
		id:            nextPtr(),
		steps:         NewSharedList[*StepInfo]("steps"),
		selectedIndex: NewShared[int64]("selectedIndex"),
	}
	c.properties = []core.Property{c.steps, c.selectedIndex}

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

func (c *Stepper) ID() ora.Ptr {
	return c.id
}

func (c *Stepper) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Stepper) Render() ora.Component {
	return c.render()
}

func (c *Stepper) render() ora.Stepper {
	var steps []ora.StepInfo
	c.steps.Iter(func(info *StepInfo) bool {
		steps = append(steps, info.render())
		return true
	})

	return ora.Stepper{
		Ptr:  c.id,
		Type: ora.StepperT,
		Steps: ora.Property[[]ora.StepInfo]{
			Ptr:   c.steps.ID(),
			Value: steps,
		},
		SelectedIndex: c.selectedIndex.render(),
	}
}

type StepInfo struct {
	id         ora.Ptr
	number     String // what is in the bubble
	caption    String
	details    String
	properties []core.Property
}

func NewStepInfo(with func(step *StepInfo)) *StepInfo {
	c := &StepInfo{
		id:      nextPtr(),
		number:  NewShared[string]("number"),
		caption: NewShared[string]("caption"),
		details: NewShared[string]("details"),
	}

	c.properties = []core.Property{c.number, c.caption, c.details}

	if with != nil {
		with(c)
	}

	return c
}

func (c *StepInfo) ID() ora.Ptr {
	return c.id
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

func (c *StepInfo) Properties(yield func(property core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *StepInfo) Render() ora.Component {
	return c.render()
}

func (c *StepInfo) render() ora.StepInfo {
	return ora.StepInfo{
		Ptr:     c.id,
		Type:    ora.StepInfoT,
		Number:  c.number.render(),
		Caption: c.caption.render(),
		Details: c.details.render(),
	}
}
