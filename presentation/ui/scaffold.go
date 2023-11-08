package ui

import (
	"go.wdy.de/nago/container/slice"
)

type Scaffold struct {
	Title string
	Menu  slice.Slice[ListItem1L]
	Body  View
}

func (s Scaffold) MarshalJSON() ([]byte, error) {
	return marshalJSON(s)
}

func (Scaffold) isView() {
}
