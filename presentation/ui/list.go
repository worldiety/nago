package ui

import "go.wdy.de/nago/container/slice"

// ListItem as defined by m3 see https://m3.material.io/components/lists/specs#888e18b6-581d-43a5-a878-1920a136358a
type ListItem interface {
	isView()
	isListItem()
}

type HorizontalDivider struct{}

func (h HorizontalDivider) MarshalJSON() ([]byte, error) {
	return marshalJSON(h)
}

func (h HorizontalDivider) isView() {}

func (h HorizontalDivider) isListItem() {}

// ListItem1L is one line item.
type ListItem1L struct {
	Headline    string // required
	LeadingIcon Image  // optional
	ActionEvent any    // optional
}

func (l ListItem1L) MarshalJSON() ([]byte, error) {
	return marshalJSON(l)
}

func (l ListItem1L) isView() {}

func (l ListItem1L) isListItem() {}

// ListItem2L is a two line item with a supporting text.
type ListItem2L struct {
	Headline       string // required
	SupportingText string // required
	LeadingIcon    Image  // optional
	ActionEvent    any    // optional
}

func (l ListItem2L) MarshalJSON() ([]byte, error) {
	return marshalJSON(l)
}

func (l ListItem2L) isView() {}

func (l ListItem2L) isListItem() {}

type ListView struct {
	Items slice.Slice[ListItem]
}

func (l ListView) isView() {}
