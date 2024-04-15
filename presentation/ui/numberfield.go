package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type NumberField struct {
	id             CID
	label          String
	value          Int
	placeholder    String
	hint           String
	error          String
	simple         Bool
	disabled       Bool
	onValueChanged *Func
	properties     []core.Property
}

func NewNumberField(with func(numberField *NumberField)) *NumberField {
	c := &NumberField{
		id:             nextPtr(),
		label:          NewShared[string]("label"),
		value:          NewShared[int64]("value"),
		placeholder:    NewShared[string]("placeholder"),
		hint:           NewShared[string]("hint"),
		error:          NewShared[string]("error"),
		simple:         NewShared[bool]("simple"),
		disabled:       NewShared[bool]("disabled"),
		onValueChanged: NewFunc("onValueChanged"),
	}

	c.properties = []core.Property{c.label, c.value, c.placeholder, c.hint, c.error, c.simple, c.disabled, c.disabled, c.onValueChanged}

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

func (l *NumberField) Simple() Bool {
	return l.simple
}

func (l *NumberField) Disabled() Bool {
	return l.disabled
}

func (l *NumberField) Type() protocol.ComponentType {
	return protocol.NumberFieldT
}

func (l *NumberField) Properties(yield func(core.Property) bool) {
	for _, property := range l.properties {
		if !yield(property) {
			return
		}
	}
}

func (l *NumberField) Render() protocol.Component {
	panic("not implemented")
}
