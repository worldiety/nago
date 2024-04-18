package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Slider struct {
	id               ora.Ptr
	disabled         Bool
	label            String
	hint             String
	error            String
	startValue       Float
	endValue         Float
	min              Float
	max              Float
	stepsize         Float
	startInitialized Bool
	endInitialized   Bool
	onChanged        *Func
	properties       []core.Property
}

func NewSlider(with func(slider *Slider)) *Slider {
	c := &Slider{
		id:               nextPtr(),
		disabled:         NewShared[bool]("disabled"),
		label:            NewShared[string]("label"),
		hint:             NewShared[string]("hint"),
		error:            NewShared[string]("error"),
		startValue:       NewShared[float64]("startValue"),
		endValue:         NewShared[float64]("endValue"),
		min:              NewShared[float64]("min"),
		max:              NewShared[float64]("max"),
		stepsize:         NewShared[float64]("stepsize"),
		startInitialized: NewShared[bool]("startInitialized"),
		endInitialized:   NewShared[bool]("endInitialized"),
		onChanged:        NewFunc("onChanged"),
	}

	c.properties = []core.Property{c.disabled, c.label, c.hint, c.error, c.startValue, c.endValue, c.min, c.max, c.stepsize, c.startInitialized, c.endInitialized, c.onChanged}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Slider) ID() ora.Ptr {
	return c.id
}

func (c *Slider) Type() ora.ComponentType {
	return ora.SliderT
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

func (c *Slider) StartInitialized() Bool {
	return c.startInitialized
}

func (c *Slider) EndInitialized() Bool { return c.endInitialized }

func (c *Slider) OnChanged() *Func { return c.onChanged }

func (c *Slider) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Slider) Render() ora.Component {
	return c.render()
}

func (c *Slider) render() ora.Slider {
	return ora.Slider{
		Ptr:              c.id,
		Type:             ora.SliderT,
		Disabled:         c.disabled.render(),
		Label:            c.label.render(),
		Hint:             c.hint.render(),
		Error:            c.error.render(),
		StartValue:       c.startValue.render(),
		EndValue:         c.endValue.render(),
		Min:              c.min.render(),
		Max:              c.max.render(),
		Stepsize:         c.stepsize.render(),
		StartInitialized: c.startInitialized.render(),
		EndInitialized:   c.endInitialized.render(),
		OnChanged:        renderFunc(c.onChanged),
	}
}
