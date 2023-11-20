package ui

import (
	"go.wdy.de/nago/container/slice"
	"io"
	"strings"
)

type InputField interface {
	intoInput() inputType
	setValue(v string) InputField
}

type SelectField struct {
	Label       string
	SelectedIDs []string // the preselected value
	Error       string
	Hint        string
	Multiple    bool
	Disabled    bool
	List        slice.Slice[SelectItem]
}

func (t SelectField) setValue(v string) InputField {
	if v == "" {
		t.SelectedIDs = nil
		return t
	}

	t.SelectedIDs = strings.Split(v, ",")
	for i, d := range t.SelectedIDs {
		t.SelectedIDs[i] = strings.TrimSpace(d) //TODO bad model
	}
	return t
}

func (t SelectField) intoInput() inputType {
	return inputType{
		Type:           "SelectField",
		Label:          t.Label,
		SelectValues:   t.SelectedIDs,
		Hint:           t.Hint,
		Error:          t.Error,
		Disabled:       t.Disabled,
		SelectMultiple: t.Multiple,
		SelectItems: slice.UnsafeUnwrap(slice.Map(t.List, func(idx int, v SelectItem) inputSelectItem {
			return inputSelectItem{
				ID:      v.ID,
				Caption: v.Caption,
			}
		})),
	}
}

type SelectItem struct {
	ID      string
	Caption string
}

type TextField struct {
	Label    string
	Value    string
	Error    string
	Hint     string
	Disabled bool
}

func (t TextField) setValue(v string) InputField {
	t.Value = v
	return t
}

func (t TextField) intoInput() inputType {
	return inputType{
		Type:     "TextField",
		Label:    t.Label,
		Value:    t.Value,
		Hint:     t.Hint,
		Error:    t.Error,
		Disabled: t.Disabled,
	}
}

type FileUploadField struct {
	Label    string
	Files    []ReceivedFile // only for receiving
	Error    string
	Hint     string
	Multiple bool
	Accept   string
	Disabled bool
}

type ReceivedFile struct {
	Data io.Reader
	Name string
	Size int64
}

func (t FileUploadField) intoInput() inputType {
	return inputType{
		Type:         "FileUploadField",
		Label:        t.Label,
		Hint:         t.Hint,
		Error:        t.Error,
		Disabled:     t.Disabled,
		FileAccept:   t.Accept,
		FileMultiple: t.Multiple,
	}
}

func (t FileUploadField) setValue(v string) InputField {
	return t
}
