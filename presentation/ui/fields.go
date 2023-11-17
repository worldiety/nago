package ui

import "io"

type InputField interface {
	intoInput() inputType
	setValue(v string) InputField
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
	File     ReceivedFile // only for receiving
	Error    string
	Hint     string
	Disabled bool
}

type ReceivedFile struct {
	Data io.Reader
	Name string
	Size int64
}

func (t FileUploadField) intoInput() inputType {
	return inputType{
		Type:     "FileUploadField",
		Label:    t.Label,
		Hint:     t.Hint,
		Error:    t.Error,
		Disabled: t.Disabled,
	}
}

func (t FileUploadField) setValue(v string) InputField {
	return t
}
