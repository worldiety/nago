package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type TextArea struct {
	id            ora.Ptr
	label         String
	value         String
	hint          String
	error         String
	visible       Bool
	rows          Int
	disabled      Bool
	onTextChanged *Func
	properties    []core.Property
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
		visible:       NewShared[bool]("visible"),
		onTextChanged: NewFunc("onTextChanged"),
	}

	c.properties = []core.Property{c.label, c.value, c.hint, c.error, c.disabled, c.disabled, c.onTextChanged, c.rows, c.visible}
	c.visible.Set(true)
	if with != nil {
		with(c)
	}

	return c
}

func (c *TextArea) OnTextChanged() *Func {
	return c.onTextChanged
}

func (c *TextArea) ID() ora.Ptr {
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

func (c *TextArea) Visible() Bool {
	return c.visible
}

func (c *TextArea) Type() string {
	return "TextArea"
}

func (c *TextArea) Rows() Int {
	return c.rows
}

func (c *TextArea) Properties(yield func(core.Property) bool) {
	for _, property := range c.properties {
		if !yield(property) {
			return
		}
	}
}

func (c *TextArea) Render() ora.Component {
	return c.render()
}

func (c *TextArea) render() ora.TextArea {
	return ora.TextArea{
		Ptr:           c.id,
		Type:          ora.TextAreaT,
		Label:         c.label.render(),
		Hint:          c.hint.render(),
		Error:         c.error.render(),
		Value:         c.value.render(),
		Rows:          c.rows.render(),
		Disabled:      c.disabled.render(),
		Visible:       c.visible.render(),
		OnTextChanged: renderFunc(c.onTextChanged),
	}
}
