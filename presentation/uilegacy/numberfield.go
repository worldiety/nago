package uilegacy

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ora"
)

type NumberField struct {
	id             ora.Ptr
	label          String
	value          String // TODO @Lukas: können wir das protokollseitig auf Integer forcieren?
	placeholder    String
	hint           String
	error          String
	simple         Bool
	disabled       Bool
	visible        Bool
	onValueChanged *Func
	properties     []core.Property
}

func NewNumberField(with func(numberField *NumberField)) *NumberField {
	c := &NumberField{
		id:             nextPtr(),
		label:          NewShared[string]("label"),
		value:          NewShared[string]("value"),
		placeholder:    NewShared[string]("placeholder"),
		hint:           NewShared[string]("hint"),
		error:          NewShared[string]("error"),
		simple:         NewShared[bool]("simple"),
		disabled:       NewShared[bool]("disabled"),
		visible:        NewShared[bool]("visible"),
		onValueChanged: NewFunc("onValueChanged"),
	}

	c.properties = []core.Property{c.label, c.value, c.placeholder, c.hint, c.error, c.simple, c.disabled, c.disabled, c.onValueChanged, c.visible}
	c.visible.Set(true)
	if with != nil {
		with(c)
	}

	return c
}

func (l *NumberField) OnValueChanged() *Func {
	return l.onValueChanged
}

func (l *NumberField) ID() ora.Ptr {
	return l.id
}

func (l *NumberField) Value() String {
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

func (l *NumberField) Visible() Bool {
	return l.visible
}

func (l *NumberField) Type() ora.ComponentType {
	return ora.NumberFieldT
}

func (l *NumberField) Properties(yield func(core.Property) bool) {
	for _, property := range l.properties {
		if !yield(property) {
			return
		}
	}
}

func (l *NumberField) Render() ora.Component {
	return l.render()
}

func (l *NumberField) render() ora.NumberField {
	return ora.NumberField{
		Ptr:            l.id,
		Type:           ora.NumberFieldT,
		Label:          l.label.render(),
		Hint:           l.hint.render(),
		Error:          l.error.render(),
		Value:          l.value.render(),
		Placeholder:    l.placeholder.render(),
		Disabled:       l.disabled.render(),
		Simple:         l.simple.render(),
		Visible:        l.visible.render(),
		OnValueChanged: renderFunc(l.onValueChanged),
	}
}
