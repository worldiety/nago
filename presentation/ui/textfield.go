package ui

import (
	"go.wdy.de/nago/container/slice"
)

type TextField struct {
	id            CID
	label         String
	value         String
	hint          String
	error         String
	disabled      Bool
	onTextChanged *Func
	properties    slice.Slice[Property]
}

func NewTextField(with func(t *TextField)) *TextField {
	c := &TextField{
		id:            nextPtr(),
		label:         NewShared[string]("label"),
		value:         NewShared[string]("value"),
		hint:          NewShared[string]("hint"),
		error:         NewShared[string]("error"),
		disabled:      NewShared[bool]("disabled"),
		onTextChanged: NewFunc("onTextChanged"),
	}

	c.properties = slice.Of[Property](c.label, c.value, c.hint, c.error, c.disabled, c.disabled, c.onTextChanged)

	if with != nil {
		with(c)
	}

	return c
}

func (l *TextField) OnTextChanged() *Func {
	return l.onTextChanged
}

func (l *TextField) ID() CID {
	return l.id
}

func (l *TextField) Value() String {
	return l.value
}

func (l *TextField) Label() String {
	return l.label
}

func (l *TextField) Hint() String {
	return l.hint
}

func (l *TextField) Error() String {
	return l.error
}

func (l *TextField) Disabled() Bool {
	return l.disabled
}

func (l *TextField) Type() string {
	return "TextField"
}

func (l *TextField) Properties() slice.Slice[Property] {
	return l.properties
}
