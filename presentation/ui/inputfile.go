package ui

import (
	"encoding/json"
	"go.wdy.de/nago/container/slice"
	"go.wdy.de/nago/xtime"
	"strings"
)

// An InputName must be unique in the entire component tree which represents a page. Input types with an empty name
// are omitted.
type InputName string

type InputType interface {
	isInputType()
}

// InputFile allows the client to pick device-local files and upload them to the server.
type InputFile struct {
	Name     InputName
	Multiple bool                // if true, multiple files can be selected
	Accept   slice.Slice[string] // filter patterns e.g. image/* or .pdf. Note that for file extensions you have to omit the *

}

func (InputFile) isInputType() {}

func (InputFile) isView() {}

func (v InputFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string `json:"type"`
		Name     string `json:"name"`
		Multiple bool   `json:"multiple"`
		Accept   string `json:"accept"`
	}{
		Type:     "InputFile",
		Name:     string(v.Name),
		Multiple: v.Multiple,
		Accept:   strings.Join(slice.UnsafeUnwrap(v.Accept), ","),
	})
}

type File struct {
	Name         string                 `json:"name"`
	LastModified xtime.UnixMilliseconds `json:"lastModified"`
	Size         int64                  `json:"size"`
	Type         string                 `json:"type"`
	Data         []byte                 `json:"data"`
}
