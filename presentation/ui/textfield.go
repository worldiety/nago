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
	functions     slice.Slice[*Func]
}

func NewTextField() *TextField {
	l := &TextField{
		id:            nextPtr(),
		label:         NewShared[string]("label"),
		value:         NewShared[string]("value"),
		hint:          NewShared[string]("hint"),
		error:         NewShared[string]("error"),
		disabled:      NewShared[bool]("disabled"),
		onTextChanged: NewFunc("onTextChanged"),
	}

	l.properties = slice.Of[Property](l.label, l.value, l.hint, l.error, l.disabled, l.disabled, l.onTextChanged)
	l.functions = slice.Of[*Func](l.onTextChanged)
	return l
}

func (l *TextField) OnTextChanged() *Func {
	return l.onTextChanged
}

func (l *TextField) With(f func(t *TextField)) *TextField {
	f(l)
	return l
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

func (l *TextField) Children() slice.Slice[LiveComponent] {
	return slice.Of[LiveComponent]()
}

func (l *TextField) Functions() slice.Slice[*Func] {
	return l.functions
}
