package ui

import "go.wdy.de/nago/container/slice"

type Dropdown struct {
	id         CID
	caption    String
	preIcon    EmbeddedSVG
	postIcon   EmbeddedSVG
	color      *Shared[Color]
	action     *Func
	disabled   Bool
	properties slice.Slice[Property]

	selectedIndex Int
	items         *SharedList[*DropdownItem]
}

func NewDropdown(with func(btn *Dropdown)) *Dropdown {
	c := &Dropdown{
		id:       nextPtr(),
		caption:  NewShared[string]("caption"),
		preIcon:  NewShared[SVGSrc]("preIcon"),
		postIcon: NewShared[SVGSrc]("postIcon"),
		color:    NewShared[Color]("color"),
		disabled: NewShared[bool]("disabled"),
		action:   NewFunc("action"),
	}

	c.properties = slice.Of[Property](c.caption, c.preIcon, c.postIcon, c.color, c.disabled, c.action)
	if with != nil {
		with(c)
	}
	return c
}

func (c *Dropdown) ID() CID {
	return c.id
}

func (c *Dropdown) Type() string {
	return "Dropdown"
}

func (c *Dropdown) SelectedIndex() Int {
	return c.selectedIndex
}

func (c *Dropdown) Items() *SharedList[*DropdownItem] {
	return c.items
}

type DropdownItem struct {
	id         CID
	content    String
	onSelected *Func
	properties slice.Slice[Property]
}

func NewDropdownItem(with func(dropdownItem *DropdownItem)) *DropdownItem {
	c := &DropdownItem{
		id:         nextPtr(),
		content:    NewShared[string]("content"),
		onSelected: NewFunc("onSelected"),
	}

	c.properties = slice.Of[Property](c.content, c.onSelected)

	if with != nil {
		with(c)
	}

	return c
}

func (c *StepInfo) ID() CID {
	return c.id
}

func (c *StepInfo) Type() string {
	return "StepInfo"
}

func (c *StepInfo) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *StepInfo) Number() String {
	return c.number
}

func (c *StepInfo) Caption() String {
	return c.caption
}

func (c *StepInfo) Details() String {
	return c.details
}
