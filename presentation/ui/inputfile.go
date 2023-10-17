package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
)

type InputType interface {
	isInputType()
}

// InputFile allows the client to pick device-local files and upload them to the server.
type InputFile struct {
	Name     string
	Multiple bool                // if true, multiple files can be selected
	Accept   slice.Slice[string] // filter patterns e.g. image/* or .pdf

}

func (InputFile) isInputType() {}

func (InputFile) isView() {}

func (v InputFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string   `json:"type"`
		Name     string   `json:"name"`
		Multiple bool     `json:"multiple"`
		Accept   []string `json:"accept"`
	}{
		Type:     "InputFile",
		Name:     v.Name,
		Multiple: v.Multiple,
		Accept:   slice.UnsafeUnwrap(v.Accept),
	})
}
