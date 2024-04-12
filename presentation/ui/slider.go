package ui

import "go.wdy.de/nago/container/slice"

type Slider struct {
	id          CID
	disabled    Bool
	label       String
	hint        String
	error       String
	startValue  Float
	endValue    Float
	min         Float
	max         Float
	stepsize    Float
	initialized Bool
	onChanged   *Func
	properties  slice.Slice[Property]
}

func NewSlider(with func(slider *Slider)) *Slider {
	c := &Slider{
		id:          nextPtr(),
		disabled:    NewShared[bool]("disabled"),
		label:       NewShared[string]("label"),
		hint:        NewShared[string]("hint"),
		error:       NewShared[string]("error"),
		startValue:  NewShared[float64]("startValue"),
		endValue:    NewShared[float64]("endValue"),
		min:         NewShared[float64]("min"),
		max:         NewShared[float64]("max"),
		stepsize:    NewShared[float64]("stepsize"),
		initialized: NewShared[bool]("initialized"),
		onChanged:   NewFunc("onChanged"),
	}

	c.properties = slice.Of[Property](c.disabled, c.label, c.hint, c.error, c.startValue, c.endValue, c.min, c.max, c.stepsize, c.initialized, c.onChanged)
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

func (c *Slider) StartValue() Float {
	return c.startValue
}

func (c *Slider) EndValue() Float {
	return c.endValue
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

func (c *Slider) Properties() slice.Slice[Property] {
	return c.properties
}
