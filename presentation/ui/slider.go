package ui

import "go.wdy.de/nago/container/slice"

type Slider struct {
	id         CID
	disabled   Bool
	label      String
	hint       String
	error      String
	value      Float
	min        Int
	max        Int
	stepsize   Float
	properties slice.Slice[Property]
}

func NewSlider(with func(slider *Slider)) *Slider {
	c := &Slider{
		id:       nextPtr(),
		disabled: NewShared[bool]("disabled"),
		label:    NewShared[string]("label"),
		hint:     NewShared[string]("hint"),
		error:    NewShared[string]("error"),
		value:    NewShared[float64]("value"),
		min:      NewShared[int64]("min"),
		max:      NewShared[int64]("max"),
		stepsize: NewShared[float64]("stepsize"),
	}

	c.properties = slice.Of[Property](c.disabled, c.label, c.hint, c.error, c.value, c.min, c.max, c.stepsize)
	if with != nil {
		with(c)
	}
	return c
}

func (c *Slider) ID() CID {
	return c.id
}

func (c *Slider) Type() string {
	return "Slider"
}

func (c *Slider) Disabled() Bool { return c.disabled }

func (c *Slider) Label() String { return c.label }

func (c *Slider) Hint() String { return c.hint }

func (c *Slider) Error() String { return c.error }

func (c *Slider) Value() Float {
	return c.value
}

func (c *Slider) Min() Int {
	return c.min
}

func (c *Slider) Max() Int {
	return c.max
}

func (c *Slider) Stepsize() Float {
	return c.stepsize
}

func (c *Slider) Properties() slice.Slice[Property] {
	return c.properties
}
