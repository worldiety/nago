package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type Navbar struct {
	Caption   View
	MenuItems slice.Slice[View]
}

func (Navbar) isView() {}

func (v Navbar) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      string            `json:"type"`
		Caption   View              `json:"caption"`
		MenuItems slice.Slice[View] `json:"menuItems"`
	}{
		Type:      "Navbar",
		Caption:   v.Caption,
		MenuItems: v.MenuItems,
	})
}
