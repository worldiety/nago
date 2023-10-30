package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

// Match expressions are evaluated at the client side without trigger any event.
type Match struct {
	Regex   string `json:"regex"`   // features and dialect depend on the client
	Message string `json:"message"` // text is either shown as Supporting or as Error text, if matched.
}

type InputText struct {
	Name              string             // the unique name of this input within a form
	Value             string             // the pre-filled text, as if it has been entered by the user
	Label             string             // the caption of the field
	Supporting        string             // text which describes the required information in a sentence
	Error             string             // text which signals something wrong
	Disabled          bool               // if true, the field cannot be modified by the user
	OnMatchSupporting slice.Slice[Match] // the first match sets the supporting text without causing a roundtrip
	OnMatchError      slice.Slice[Match] // the first match sets the error text without causing a roundtrip
}

func (InputText) isInputType() {}

func (InputText) isView() {}

func (v InputText) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type              string  `json:"type"`
		Name              string  `json:"name"`
		Value             string  `json:"value"`
		Label             string  `json:"label"`
		Supporting        string  `json:"supporting"`
		Error             string  `json:"error"`
		Disabled          bool    `json:"disabled"`
		OnMatchSupporting []Match `json:"onMatchSupporting"`
		OnMatchError      []Match `json:"onMatchError"`
	}{
		Type:              "InputText",
		Name:              v.Name,
		Value:             v.Value,
		Label:             v.Label,
		Supporting:        v.Supporting,
		Error:             v.Error,
		Disabled:          v.Disabled,
		OnMatchSupporting: slice.UnsafeUnwrap(v.OnMatchSupporting),
		OnMatchError:      slice.UnsafeUnwrap(v.OnMatchError),
	})
}
