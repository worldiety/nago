package ui

import (
	"fmt"
	"go.wdy.de/nago/container/slice"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type InputField interface {
	intoInput() inputType
	setValue(v string) InputField
}

type SwitchField struct {
	Label    string
	Value    bool
	Error    string
	Hint     string
	Disabled bool
}

func (t SwitchField) setValue(v string) InputField {
	b, _ := strconv.ParseBool(v)
	t.Value = b
	return t
}

func (t SwitchField) intoInput() inputType {
	return inputType{
		Type:     "SwitchField",
		Label:    t.Label,
		Hint:     t.Hint,
		Error:    t.Error,
		Value:    strconv.FormatBool(t.Value),
		Disabled: t.Disabled,
	}
}

type SelectField[ID ~string] struct {
	Label       string
	SelectedIDs []ID // the preselected value
	Error       string
	Hint        string
	Multiple    bool
	Disabled    bool
	List        slice.Slice[SelectItem[ID]]
}

func (t SelectField[ID]) setValue(v string) InputField {
	if v == "" {
		t.SelectedIDs = nil
		return t
	}

	t.SelectedIDs = mapStrs2Ids[ID](strings.Split(v, ","))
	for i, d := range t.SelectedIDs {
		t.SelectedIDs[i] = ID(strings.TrimSpace(string(d))) //TODO bad model
	}
	return t
}

func (t SelectField[ID]) intoInput() inputType {
	return inputType{
		Type:           "SelectField",
		Label:          t.Label,
		SelectValues:   mapIds2Strs(t.SelectedIDs),
		Hint:           t.Hint,
		Error:          t.Error,
		Disabled:       t.Disabled,
		SelectMultiple: t.Multiple,
		SelectItems: slice.UnsafeUnwrap(slice.Map(t.List, func(idx int, v SelectItem[ID]) inputSelectItem {
			return inputSelectItem{
				ID:      string(v.ID),
				Caption: v.Caption,
			}
		})),
	}
}

func mapStrs2Ids[ID ~string](s []string) []ID {
	var res []ID
	for _, s2 := range s {
		res = append(res, ID(s2))
	}
	return res
}

func mapIds2Strs[ID ~string](s []ID) []string {
	var res []string
	for _, id := range s {
		res = append(res, string(id))
	}
	return res
}

type SelectItem[ID ~string] struct {
	ID      ID
	Caption string
}

type DateField struct {
	Label    string
	Value    time.Time
	Error    string
	Hint     string
	Disabled bool
}

func (t DateField) setValue(v string) InputField {
	d, err := time.Parse(time.DateOnly, v) // TODO decide about timezone behavior
	if err != nil {
		slog.Default().Error(fmt.Sprintf("cannot parse date field format '%s': %v", v, err))
	}
	t.Value = d
	return t
}

func (t DateField) intoInput() inputType {
	return inputType{
		Type:     "DateField",
		Label:    t.Label,
		Value:    t.Value.Format(time.DateOnly),
		Hint:     t.Hint,
		Error:    t.Error,
		Disabled: t.Disabled,
	}
}

type TextAreaField struct {
	Label    string
	Value    string
	Error    string
	Hint     string
	Disabled bool
}

func (t TextAreaField) setValue(v string) InputField {
	t.Value = v
	return t
}

func (t TextAreaField) intoInput() inputType {
	return inputType{
		Type:     "TextAreaField",
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
