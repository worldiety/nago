package ui

import "go.wdy.de/nago/container/slice"

type Dropdown struct {
	id               CID
	selectedIndexes  *SharedList[int64]
	items            *SharedList[*DropdownItem]
	expanded         Bool
	disabled         Bool
	label            String
	hint             String
	error            String
	onToggleExpanded *Func
	properties       slice.Slice[Property]
}

func NewDropdown(with func(dropdown *Dropdown)) *Dropdown {
	c := &Dropdown{
		id:               nextPtr(),
		selectedIndexes:  NewSharedList[int64]("selectedIndexes"),
		items:            NewSharedList[*DropdownItem]("items"),
		expanded:         NewShared[bool]("expanded"),
		disabled:         NewShared[bool]("disabled"),
		label:            NewShared[string]("label"),
		hint:             NewShared[string]("hint"),
		error:            NewShared[string]("error"),
		onToggleExpanded: NewFunc("onToggleExpanded"),
	}

	c.properties = slice.Of[Property](c.selectedIndexes, c.items, c.expanded, c.disabled, c.label, c.hint, c.error, c.onToggleExpanded)
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

func (c *Dropdown) SelectedIndexes() *SharedList[int64] {
	return c.selectedIndexes
}

func (c *Dropdown) Items() *SharedList[*DropdownItem] {
	return c.items
}

func (c *Dropdown) Expanded() Bool {
	return c.expanded
}

func (c *Dropdown) Disabled() Bool { return c.disabled }

func (c *Dropdown) Label() String { return c.label }

func (c *Dropdown) Hint() String { return c.hint }

func (c *Dropdown) Error() String { return c.error }

func (c *Dropdown) OnToggleExpanded() *Func {
	return c.onToggleExpanded
}

func (c *Dropdown) Properties() slice.Slice[Property] {
	return c.properties
}

type DropdownItem struct {
	id         CID
	itemIndex  Int
	content    String
	onSelected *Func
	properties slice.Slice[Property]
}

func NewDropdownItem(with func(dropdownItem *DropdownItem)) *DropdownItem {
	c := &DropdownItem{
		id:         nextPtr(),
		itemIndex:  NewShared[int64]("itemIndex"),
		content:    NewShared[string]("content"),
		onSelected: NewFunc("onSelected"),
	}

	c.properties = slice.Of[Property](c.itemIndex, c.content, c.onSelected)

	if with != nil {
		with(c)
	}

	return c
}

func (c *DropdownItem) ID() CID {
	return c.id
}

func (c *DropdownItem) Type() string {
	return "DropdownItem"
}

func (c *DropdownItem) Properties() slice.Slice[Property] {
	return c.properties
}

func (c *DropdownItem) ItemIndex() Int {
	return c.itemIndex
}

func (c *DropdownItem) Content() String {
	return c.content
}

func (c *DropdownItem) OnSelected() *Func {
	return c.onSelected
}
