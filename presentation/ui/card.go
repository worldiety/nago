package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type Card struct {
	Child    View              // first
	Children slice.Slice[View] // others
}

func (Card) isView() {}

func (v Card) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Views []View `json:"views"`
	}{
		Type:  "Card",
		Views: joinViews(v.Child, v.Children),
	})
}
