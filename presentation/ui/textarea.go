package ui

import (
	"go.wdy.de/nago/container/slice"
)

type TextArea struct {
	id            CID
	label         String
	value         String
	hint          String
	error         String
	rows          Int
	disabled      Bool
	onTextChanged *Func
	properties    slice.Slice[Property]
}

func NewTextArea(with func(textArea *TextArea)) *TextArea {
	c := &TextArea{
		id:            nextPtr(),
		label:         NewShared[string]("label"),
		value:         NewShared[string]("value"),
		hint:          NewShared[string]("hint"),
		error:         NewShared[string]("error"),
		disabled:      NewShared[bool]("disabled"),
		rows:          NewShared[int64]("rows"),
		onTextChanged: NewFunc("onTextChanged"),
	}

	c.properties = slice.Of[Property](c.label, c.value, c.hint, c.error, c.disabled, c.disabled, c.onTextChanged, c.rows)

	if with != nil {
		with(c)
	}

	return c
}

func (c *TextArea) OnTextChanged() *Func {
	return c.onTextChanged
}

func (c *TextArea) ID() CID {
	return c.id
}

func (c *TextArea) Value() String {
	return c.value
}

func (c *TextArea) Label() String {
	return c.label
}

func (c *TextArea) Hint() String {
	return c.hint
}

func (c *TextArea) Error() String {
	return c.error
}

func (c *TextArea) Disabled() Bool {
	return c.disabled
}

func (c *TextArea) Type() string {
	return "TextArea"
}

func (c *TextArea) Rows() Int {
	return c.rows
}

func (c *TextArea) Properties() slice.Slice[Property] {
	return c.properties
}
