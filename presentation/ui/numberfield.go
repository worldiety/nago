package ui

import (
	"go.wdy.de/nago/container/slice"
)

type NumberField struct {
	id             CID
	label          String
	value          Int
	placeholder    String
	hint           String
	error          String
	disabled       Bool
	onValueChanged *Func
	properties     slice.Slice[Property]
}

func NewNumberField(with func(numberField *NumberField)) *NumberField {
	c := &NumberField{
		id:             nextPtr(),
		label:          NewShared[string]("label"),
		value:          NewShared[int64]("value"),
		placeholder:    NewShared[string]("placeholder"),
		hint:           NewShared[string]("hint"),
		error:          NewShared[string]("error"),
		disabled:       NewShared[bool]("disabled"),
		onValueChanged: NewFunc("onValueChanged"),
	}

	c.properties = slice.Of[Property](c.label, c.value, c.placeholder, c.hint, c.error, c.disabled, c.disabled, c.onValueChanged)

	if with != nil {
		with(c)
	}

	return c
}

func (l *NumberField) OnValueChanged() *Func {
	return l.onValueChanged
}

func (l *NumberField) ID() CID {
	return l.id
}

func (l *NumberField) Value() Int {
	return l.value
}

func (l *NumberField) Placeholder() String { return l.placeholder }

func (l *NumberField) Label() String {
	return l.label
}

func (l *NumberField) Hint() String {
	return l.hint
}

func (l *NumberField) Error() String {
	return l.error
}

func (l *NumberField) Disabled() Bool {
	return l.disabled
}

func (l *NumberField) Type() string {
	return "NumberField"
}

func (l *NumberField) Properties() slice.Slice[Property] {
	return l.properties
}
