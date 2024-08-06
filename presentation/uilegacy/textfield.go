package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
	"time"
)

type TextField struct {
	id                     ora.Ptr
	label                  String
	value                  String
	placeholder            String
	hint                   String
	help                   String
	error                  String
	disabled               Bool
	simple                 Bool
	visible                Bool
	onTextChanged          *Func
	onDebouncedTextChanged *Func
	debounceTime           *Shared[time.Duration]
	properties             []core.Property
	frame                  ora.Frame
}

func NewTextField(with func(textField *TextField)) *TextField {
	c := &TextField{
		id:                     nextPtr(),
		label:                  NewShared[string]("label"),
		value:                  NewShared[string]("value"),
		placeholder:            NewShared[string]("placeholder"), // TODO remove me, does not make sense from UX perspective, we have Label and Hint
		hint:                   NewShared[string]("hint"),
		error:                  NewShared[string]("error"),
		disabled:               NewShared[bool]("disabled"),
		help:                   NewShared[string]("help"), // TODO remove me, does not make sense from UX perspective, we have Label and Hint
		simple:                 NewShared[bool]("simple"), // TODO what is that?
		visible:                NewShared[bool]("visible"),
		onTextChanged:          NewFunc("onTextChanged"),
		onDebouncedTextChanged: NewFunc("onDebouncedTextChanged"),
		debounceTime:           NewShared[time.Duration]("debounceTime"),
	}

	c.properties = []core.Property{c.label, c.value, c.placeholder, c.hint, c.help, c.error, c.disabled, c.simple, c.onTextChanged, c.visible, c.debounceTime, c.onTextChanged}
	c.visible.Set(true)
	c.debounceTime.Set(time.Millisecond * 500)
	if with != nil {
		with(c)
	}

	return c
}

func (l *TextField) OnTextChanged() *Func {
	return l.onTextChanged
}

func (l *TextField) OnDebouncedTextChanged() *Func {
	return l.onDebouncedTextChanged
}

func (l *TextField) DebounceTime() *Shared[time.Duration] {
	return l.debounceTime
}

func (l *TextField) ID() ora.Ptr {
	return l.id
}

func (l *TextField) Value() String {
	return l.value
}

func (l *TextField) SetValue(s String) {
	l.value = s
	l.properties = append(l.properties, s)
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

func (l *TextField) Frame() *ora.Frame { return &l.frame }

func (l *TextField) Properties(yield func(core.Property) bool) {
	for _, property := range l.properties {
		if !yield(property) {
			return
		}
	}
}

func (l *TextField) Visible() Bool {
	return l.visible
}

func (l *TextField) Render() ora.Component {
	//return ora.TextField{
	//	Ptr:                    l.id,
	//	Type:                   ora.TextFieldT,
	//	Label:                  l.label.render(),
	//	Hint:                   l.hint.render(),
	//	Help:                   l.help.render(),
	//	Error:                  l.error.render(),
	//	Value:                  l.value.render(),
	//	Placeholder:            l.placeholder.render(),
	//	Disabled:               l.disabled.render(),
	//	Simple:                 l.simple.render(),
	//	Visible:                l.visible.render(),
	//	DebounceTime:           l.debounceTime.render(),
	//	OnDebouncedTextChanged: renderFunc(l.onDebouncedTextChanged),
	//	OnTextChanged:          renderFunc(l.onTextChanged),
	//	Frame:                  l.frame,
	//}
	return nil
}
