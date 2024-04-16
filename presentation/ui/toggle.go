package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

// Toggle is like a checkbox, which is either on or off.
type Toggle struct {
	id               CID
	label            String
	checked          Bool
	disabled         Bool
	properties       []core.Property
	onCheckedChanged *Func
}

func NewToggle(with func(tgl *Toggle)) *Toggle {
	c := &Toggle{
		id:               nextPtr(),
		label:            NewShared[string]("label"),
		disabled:         NewShared[bool]("disabled"),
		checked:          NewShared[bool]("checked"),
		onCheckedChanged: NewFunc("onCheckedChanged"),
	}

	c.properties = []core.Property{c.label, c.disabled, c.checked, c.onCheckedChanged}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Toggle) ID() CID {
	return c.id
}

func (c *Toggle) Type() string {
	return "Toggle"
}

func (c *Toggle) OnCheckedChanged() *Func {
	return c.onCheckedChanged
}

func (c *Toggle) Label() String {
	return c.label
}

func (c *Toggle) Disabled() Bool {
	return c.disabled
}

func (c *Toggle) Checked() Bool {
	return c.checked
}

func (c *Toggle) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Toggle) Render() ora.Component {
	return c.render()
}

func (c *Toggle) render() ora.Toggle {
	return ora.Toggle{
		Ptr:              c.id,
		Type:             ora.ToggleT,
		Label:            c.label.render(),
		Checked:          c.checked.render(),
		Disabled:         c.disabled.render(),
		OnCheckedChanged: renderFunc(c.onCheckedChanged),
	}
}
