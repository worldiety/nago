package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/protocol"
)

type TextField struct {
	id            CID
	label         String
	value         String
	placeholder   String
	hint          String
	help          String
	error         String
	disabled      Bool
	simple        Bool
	onTextChanged *Func
	properties    []core.Property
}

func NewTextField(with func(textField *TextField)) *TextField {
	c := &TextField{
		id:            nextPtr(),
		label:         NewShared[string]("label"),
		value:         NewShared[string]("value"),
		placeholder:   NewShared[string]("placeholder"),
		hint:          NewShared[string]("hint"),
		help:          NewShared[string]("help"),
		error:         NewShared[string]("error"),
		disabled:      NewShared[bool]("disabled"),
		simple:        NewShared[bool]("simple"),
		onTextChanged: NewFunc("onTextChanged"),
	}

	c.properties = []core.Property{c.label, c.value, c.placeholder, c.hint, c.help, c.error, c.disabled, c.disabled, c.simple, c.onTextChanged}

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

func (l *TextField) Placeholder() String { return l.placeholder }

func (l *TextField) Label() String {
	return l.label
}

func (l *TextField) Hint() String {
	return l.hint
}

func (l *TextField) Help() String {
	return l.help
}

func (l *TextField) Error() String {
	return l.error
}

func (l *TextField) Disabled() Bool {
	return l.disabled
}

func (l *TextField) Simple() Bool { return l.simple }

func (l *TextField) Properties(yield func(core.Property) bool) {
	for _, property := range l.properties {
		if !yield(property) {
			return
		}
	}
}

func (l *TextField) Render() protocol.Component {
	panic("not implemented")
}
