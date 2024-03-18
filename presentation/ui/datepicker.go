package ui

import "go.wdy.de/nago/container/slice"

type Datepicker struct {
	id         CID
	properties slice.Slice[Property]
}

func NewDatepicker(with func(datepicker *Datepicker)) *Datepicker {
	c := &Datepicker{
		id: nextPtr(),
	}

	c.properties = slice.Of[Property]()
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

func (c *Datepicker) Properties() slice.Slice[Property] {
	return c.properties
}
