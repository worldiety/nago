package ui

import "go.wdy.de/nago/container/slice"

// A Chip is like a badge but removable.
type Chip struct {
	id         CID
	caption    String
	action     *Func
	onClose    *Func
	color      *Shared[Color]
	properties slice.Slice[Property]
}

func NewChip(with func(chip *Chip)) *Chip {
	c := &Chip{
		id:      nextPtr(),
		caption: NewShared[string]("caption"),
		action:  NewFunc("action"),
		onClose: NewFunc("onClose"),
		color:   NewShared[Color]("color"),
	}

	c.properties = slice.Of[Property](c.caption, c.action, c.onClose, c.color)

	if with != nil {
		with(c)
	}

	return c
}

func (c *Chip) ID() CID {
	return c.id
}

func (c *Chip) Type() string {
	return "Chip"
}

func (c *Chip) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *Chip) Caption() String {
	return c.caption
}

func (c *Chip) Action() *Func {
	return c.action
}

func (c *Chip) OnClose() *Func {
	return c.onClose
}

// TBD: red, green, yellow
func (c *Chip) Color() *Shared[Color] {
	return c.color
}
