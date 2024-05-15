package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Checkbox struct {
	id         ora.Ptr
	selected   Bool
	clicked    *Func
	disabled   Bool
	properties []core.Property
}

func NewCheckbox(with func(chb *Checkbox)) *Checkbox {
	c := &Checkbox{
		id:       nextPtr(),
		selected: NewShared[bool]("selected"),
		clicked:  NewFunc("action"),
		disabled: NewShared[bool]("disabled"),
	}

	c.properties = []core.Property{c.selected, c.clicked, c.disabled}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Checkbox) ID() ora.Ptr {
	return c.id
}

func (c *Checkbox) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Checkbox) Render() ora.Component {
	return c.renderCheckbox()
}

func (c *Checkbox) Selected() Bool { return c.selected }

func (c *Checkbox) Clicked() *Func {
	return c.clicked
}

func (c *Checkbox) Disabled() Bool {
	return c.disabled
}

func (c *Checkbox) renderCheckbox() ora.Checkbox {
	return ora.Checkbox{
		Ptr:      c.id,
		Type:     ora.CheckboxT,
		Disabled: c.disabled.render(),
		Selected: c.selected.render(),
		Clicked:  renderFunc(c.clicked),
	}
}
