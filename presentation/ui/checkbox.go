package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Checkbox struct {
	id         ora.Ptr
	selected   Bool
	onClicked  *Func
	disabled   Bool
	visible    Bool
	properties []core.Property
}

func NewCheckbox(with func(chb *Checkbox)) *Checkbox {
	c := &Checkbox{
		id:        nextPtr(),
		selected:  NewShared[bool]("selected"),
		onClicked: NewFunc("action"),
		disabled:  NewShared[bool]("disabled"),
		visible:   NewShared[bool]("visible"),
	}

	c.properties = []core.Property{c.selected, c.onClicked, c.disabled, c.visible}
	c.visible.Set(true)
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

func (c *Checkbox) OnClicked() *Func {
	return c.onClicked
}

func (c *Checkbox) Disabled() Bool {
	return c.disabled
}

func (c *Checkbox) Visible() Bool {
	return c.visible
}

func (c *Checkbox) renderCheckbox() ora.Checkbox {
	return ora.Checkbox{
		Ptr:       c.id,
		Type:      ora.CheckboxT,
		Disabled:  c.disabled.render(),
		Selected:  c.selected.render(),
		Visible:   c.visible.render(),
		OnClicked: renderFunc(c.onClicked),
	}
}
