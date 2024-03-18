package ui

import "go.wdy.de/nago/container/slice"

type Datepicker struct {
	id               CID
	disabled         Bool
	label            String
	hint             String
	error            String
	expanded         Bool
	onToggleExpanded *Func
	properties       slice.Slice[Property]
}

func NewDatepicker(with func(datepicker *Datepicker)) *Datepicker {
	c := &Datepicker{
		id:               nextPtr(),
		disabled:         NewShared[bool]("disabled"),
		label:            NewShared[string]("label"),
		hint:             NewShared[string]("hint"),
		error:            NewShared[string]("error"),
		expanded:         NewShared[bool]("expanded"),
		onToggleExpanded: NewFunc("onToggleExpanded"),
	}

	c.properties = slice.Of[Property](c.disabled, c.label, c.hint, c.error, c.expanded, c.onToggleExpanded)
	if with != nil {
		with(c)
	}
	return c
}

func (c *Datepicker) ID() CID {
	return c.id
}

func (c *Datepicker) Type() string {
	return "Datepicker"
}

func (c *Datepicker) Disabled() Bool {
	return c.disabled
}

func (c *Datepicker) Label() String { return c.label }

func (c *Datepicker) Hint() String { return c.hint }

func (c *Datepicker) Error() String { return c.error }

func (c *Datepicker) Expanded() Bool {
	return c.expanded
}

func (c *Datepicker) OnToggleExpanded() *Func {
	return c.onToggleExpanded
}

func (c *Datepicker) Properties() slice.Slice[Property] {
	return c.properties
}
