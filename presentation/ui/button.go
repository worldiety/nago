package ui

import "encoding/json"

type Button struct {
	Title   Texter
	OnClick any
}

func (Button) isView() {}

func (b Button) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string `json:"type"`
		Title   Texter `json:"title"`
		OnClick any    `json:"onClick"`
	}{
		Type:    "Button",
		Title:   b.Title,
		OnClick: b.OnClick,
	})
}
