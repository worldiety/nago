package ui

import "go.wdy.de/nago/container/slice"

type Dropdown struct {
	id               CID
	selectedIndices  *SharedList[int64]
	items            *SharedList[*DropdownItem]
	multiselect      Bool
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
		selectedIndices:  NewSharedList[int64]("selectedIndices"),
		items:            NewSharedList[*DropdownItem]("items"),
		multiselect:      NewShared[bool]("multiselect"),
		expanded:         NewShared[bool]("expanded"),
		disabled:         NewShared[bool]("disabled"),
		label:            NewShared[string]("label"),
		hint:             NewShared[string]("hint"),
		error:            NewShared[string]("error"),
		onToggleExpanded: NewFunc("onToggleExpanded"),
	}

	c.properties = slice.Of[Property](c.selectedIndices, c.items, c.multiselect, c.expanded, c.disabled, c.label, c.hint, c.error, c.onToggleExpanded)
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

func (c *Dropdown) SelectedIndices() *SharedList[int64] {
	return c.selectedIndices
}

// Toggle toggles a dropdown item's selected state.
// If the dropdown is in multiselect mode, multiple items may be selected at the same time.
// Otherwise, only a single item may be selected at the same time.
func (c *Dropdown) Toggle(item *DropdownItem) {
	itemIndex := item.ItemIndex().Get()

	if c.Multiselect().Get() != true {
		c.SelectedIndices().Clear()
		c.SelectedIndices().Append(itemIndex)
		c.Expanded().Set(false)
		return
	}

	contains := false
	c.SelectedIndices().Each(func(index int64) {
		if itemIndex == index {
			contains = true
			return
		}
	})
	if contains {
		c.SelectedIndices().Remove(itemIndex)
	} else {
		c.SelectedIndices().Append(itemIndex)
	}
}

func (c *Dropdown) Items() *SharedList[*DropdownItem] {
	return c.items
}

func (c *Dropdown) Multiselect() Bool {
	return c.multiselect
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
