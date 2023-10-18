package ui

import "encoding/json"

// Text is a string whose style depends on its context.
type Text string

func (Text) isText() {}
func (Text) isView() {}

func (v Text) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "Text",
		Value: string(v),
	})
}

type Texter interface {
	isText()
}

// AttributedText takes the string and some style options.
type AttributedText struct {
	Value string
}

func (AttributedText) isText() {}

func (v AttributedText) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "AttributedText",
		Value: v.Value,
	})
}
