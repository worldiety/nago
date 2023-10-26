package ui

import (
	"encoding/json"
	"reflect"
)

type Button struct {
	Title   Texter
	OnClick any
}

func (Button) isView() {}

func (b Button) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string      `json:"type"`
		Title   Texter      `json:"title"`
		OnClick eventSource `json:"onClick"`
	}{
		Type:  "Button",
		Title: b.Title,
		OnClick: eventSource{
			EventType: reflect.TypeOf(b.OnClick).String(),
			Data:      b.OnClick,
		},
	})
}

type eventSource struct {
	EventType string `json:"eventType"`
	Data      any    `json:"data"`
}
