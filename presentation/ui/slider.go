package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Slider struct {
	id          CID
	disabled    Bool
	label       String
	hint        String
	error       String
	value       Float
	min         Float
	max         Float
	stepsize    Float
	initialized Bool
	onChanged   *Func
	properties  []core.Property
}

func NewSlider(with func(slider *Slider)) *Slider {
	c := &Slider{
		id:          nextPtr(),
		disabled:    NewShared[bool]("disabled"),
		label:       NewShared[string]("label"),
		hint:        NewShared[string]("hint"),
		error:       NewShared[string]("error"),
		value:       NewShared[float64]("value"),
		min:         NewShared[float64]("min"),
		max:         NewShared[float64]("max"),
		stepsize:    NewShared[float64]("stepsize"),
		initialized: NewShared[bool]("initialized"),
		onChanged:   NewFunc("onChanged"),
	}

	c.properties = []core.Property{c.disabled, c.label, c.hint, c.error, c.value, c.min, c.max, c.stepsize, c.initialized, c.onChanged}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Slider) ID() CID {
	return c.id
}

func (c *Slider) Type() ora.ComponentType {
	return ora.SliderT
}

func (c *Slider) Disabled() Bool { return c.disabled }

func (c *Slider) Label() String { return c.label }

func (c *Slider) Hint() String { return c.hint }

func (c *Slider) Error() String { return c.error }

func (c *Slider) Value() Float {
	return c.value
}

func (c *Slider) Min() Float {
	return c.min
}

func (c *Slider) Max() Float {
	return c.max
}

func (c *Slider) Stepsize() Float {
	return c.stepsize
}

func (c *Slider) Initialized() Bool {
	return c.initialized
}

func (c *Slider) OnChanged() *Func { return c.onChanged }

func (c *Slider) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Slider) Render() ora.Component {
	panic("not implemented")
}
