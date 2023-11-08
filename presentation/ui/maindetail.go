package ui

import (
	"go.wdy.de/nago/container/slice"
	"net/http"
)

type ListLoader func() ListView
type DetailLoader func(item ListItem) View

// A MainDetail view is either a stacked (mobile) or a two column side by side layout (desktop).
type MainDetail struct {
	Main   ListLoader
	Detail DetailLoader
}

func (m MainDetail) MarshalJSON() ([]byte, error) {
	return marshalJSON(m)
}

func (MainDetail) isView() {}

type MainDetailAdapter interface {
	Main(r http.Request) slice.Slice[ListItem]
	Detail(r http.Request, idx int) View
}
