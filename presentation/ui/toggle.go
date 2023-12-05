package ui

import "go.wdy.de/nago/container/slice"

// Toggle is like a checkbox, which is either on or off.
type Toggle struct {
	id               CID
	label            String
	checked          Bool
	disabled         Bool
	properties       slice.Slice[Property]
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

	c.properties = slice.Of[Property](c.label, c.disabled, c.checked, c.onCheckedChanged)
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

func (c *Toggle) Properties() slice.Slice[Property] {
	return c.properties
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
