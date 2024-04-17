package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type Dropdown struct {
	id              CID
	selectedIndices *SharedList[int64]
	items           *SharedList[*DropdownItem]
	multiselect     Bool
	expanded        Bool
	disabled        Bool
	label           String
	hint            String
	error           String
	onClicked       *Func
	properties      []core.Property
}

func NewDropdown(with func(dropdown *Dropdown)) *Dropdown {
	c := &Dropdown{
		id:              nextPtr(),
		selectedIndices: NewSharedList[int64]("selectedIndices"),
		items:           NewSharedList[*DropdownItem]("items"),
		multiselect:     NewShared[bool]("multiselect"),
		expanded:        NewShared[bool]("expanded"),
		disabled:        NewShared[bool]("disabled"),
		label:           NewShared[string]("label"),
		hint:            NewShared[string]("hint"),
		error:           NewShared[string]("error"),
		onClicked:       NewFunc("onClicked"),
	}

	c.properties = []core.Property{c.selectedIndices, c.items, c.multiselect, c.expanded, c.disabled, c.label, c.hint, c.error, c.onClicked}
	if with != nil {
		with(c)
	}
	return c
}

func (c *Dropdown) ID() CID {
	return c.id
}

func (c *Dropdown) SelectedIndices() *SharedList[int64] {
	return c.selectedIndices
}

// indexOf determines the index of a dropdown item based on its DropdownItem.id
func (c *Dropdown) indexOf(item *DropdownItem) int64 {
	index := -1
	counter := 0
	c.items.Iter(func(it *DropdownItem) bool {
		if it.id == item.id {
			index = counter
		}
		counter++
		return true
	})

	return int64(index)
}

// Toggle toggles a dropdown item's selected state.
// If the dropdown is in multiselect mode, multiple items may be selected at the same time.
// Otherwise, only a single item may be selected at the same time.
func (c *Dropdown) Toggle(item *DropdownItem) {
	itemIndex := c.indexOf(item)
	if c.Multiselect().Get() != true {
		c.SelectedIndices().Clear()
		c.SelectedIndices().Append(itemIndex)
		c.Expanded().Set(false)
		return
	}

	contains := false
	c.SelectedIndices().Iter(func(index int64) bool {
		if itemIndex == index {
			contains = true
			return false
		}

		return true
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

func (c *Dropdown) OnClicked() *Func {
	return c.onClicked
}

func (c *Dropdown) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *Dropdown) Render() ora.Component {
	return c.render()
}

func (c *Dropdown) render() ora.Dropdown {
	var items []ora.DropdownItem
	c.items.Iter(func(it *DropdownItem) bool {
		items = append(items, it.render())
		return true
	})

	return ora.Dropdown{
		Ptr:  c.id,
		Type: ora.DropdownT,
		Items: ora.Property[[]ora.DropdownItem]{
			Ptr:   c.items.ID(),
			Value: items,
		},
		SelectedIndices: c.selectedIndices.render(),
		Multiselect:     c.multiselect.render(),
		Expanded:        c.expanded.render(),
		Disabled:        c.disabled.render(),
		Label:           c.label.render(),
		Hint:            c.hint.render(),
		Error:           c.error.render(),
		OnClicked:       renderFunc(c.onClicked),
	}
}

type DropdownItem struct {
	id         CID
	content    String
	onClicked  *Func
	properties []core.Property
}

func NewDropdownItem(with func(dropdownItem *DropdownItem)) *DropdownItem {
	c := &DropdownItem{
		id:        nextPtr(),
		content:   NewShared[string]("content"),
		onClicked: NewFunc("onClicked"),
	}

	c.properties = []core.Property{c.content, c.onClicked}

	if with != nil {
		with(c)
	}

	return c
}

func (c *DropdownItem) ID() CID {
	return c.id
}

func (c *DropdownItem) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *DropdownItem) Content() String {
	return c.content
}

func (c *DropdownItem) OnClicked() *Func {
	return c.onClicked
}

func (c *DropdownItem) Render() ora.Component {
	return c.render()
}

func (c *DropdownItem) render() ora.DropdownItem {
	return ora.DropdownItem{
		Ptr:       c.id,
		Type:      ora.DropdownItemT,
		Content:   c.content.render(),
		OnClicked: renderFunc(c.onClicked),
	}
}
