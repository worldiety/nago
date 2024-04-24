package ui

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type PasswordField struct {
	id                ora.Ptr
	label             String
	value             String
	revealed          Bool
	placeholder       String
	hint              String
	help              String
	error             String
	disabled          Bool
	simple            Bool
	onPasswordChanged *Func
	properties        []core.Property
}

func NewPasswordField(with func(passwordField *PasswordField)) *PasswordField {
	c := &PasswordField{
		id:                nextPtr(),
		label:             NewShared[string]("label"),
		value:             NewShared[string]("value"),
		revealed:          NewShared[bool]("revealed"),
		placeholder:       NewShared[string]("placeholder"), // TODO remove me, does not make sense from UX perspective, we have Label and Hint
		hint:              NewShared[string]("hint"),
		error:             NewShared[string]("error"),
		disabled:          NewShared[bool]("disabled"),
		help:              NewShared[string]("help"), // TODO remove me, does not make sense from UX perspective, we have Label and Hint
		simple:            NewShared[bool]("simple"), // TODO what is that?
		onPasswordChanged: NewFunc("onPasswordChanged"),
	}

	c.properties = []core.Property{c.label, c.value, c.revealed, c.placeholder, c.hint, c.help, c.error, c.disabled, c.simple, c.onPasswordChanged}

	if with != nil {
		with(c)
	}

	return c
}

func (l *PasswordField) OnPasswordChanged() *Func {
	return l.onPasswordChanged
}

func (l *PasswordField) ID() ora.Ptr {
	return l.id
}

func (l *PasswordField) Value() String {
	return l.value
}

func (l *PasswordField) Revealed() Bool {
	return l.revealed
}

func (l *PasswordField) Placeholder() String { return l.placeholder }

func (l *PasswordField) Label() String {
	return l.label
}

func (l *PasswordField) Hint() String {
	return l.hint
}

func (l *PasswordField) Help() String {
	return l.help
}

func (l *PasswordField) Error() String {
	return l.error
}

func (l *PasswordField) Disabled() Bool {
	return l.disabled
}

func (l *PasswordField) Simple() Bool { return l.simple }

func (l *PasswordField) Properties(yield func(core.Property) bool) {
	for _, property := range l.properties {
		if !yield(property) {
			return
		}
	}
}

func (l *PasswordField) Render() ora.Component {
	return ora.PasswordField{
		Ptr:               l.id,
		Type:              ora.PasswordFieldT,
		Label:             l.label.render(),
		Hint:              l.hint.render(),
		Help:              l.help.render(),
		Error:             l.error.render(),
		Value:             l.value.render(),
		Revealed:          l.revealed.render(),
		Placeholder:       l.placeholder.render(),
		Disabled:          l.disabled.render(),
		Simple:            l.simple.render(),
		OnPasswordChanged: renderFunc(l.onPasswordChanged),
	}
}
