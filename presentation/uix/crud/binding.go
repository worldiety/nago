package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Binding[T any] struct {
	fields []anyField[T]
}

func NewBinding[T any](with func(bnd *Binding[T])) *Binding[T] {
	b := &Binding[T]{}
	if with != nil {
		with(b)
	}
	return b
}

func Text[Model any](b *Binding[Model], field Field[Model, string]) {
	f := anyField[Model]{
		Caption:     field.Caption,
		Stringer:    field.Stringer,
		RenderHints: field.RenderHints,
	}

	if f.Stringer == nil {
		f.Stringer = func(model Model) string {
			return fmt.Sprintf("%v", model)
		}
	}

	if f.FromModel == nil {
		f.FromModel = func(model Model) any {
			return f.Stringer(model)
		}
	}

	f.FormFactory = func(variant RenderHint) formElement[Model] {
		component := ui.NewTextField(func(textField *ui.TextField) {
			textField.Label().Set(f.Caption)
			switch variant {
			case Visible:
				textField.Visible().Set(true)
			case ReadOnly:
				textField.Disabled().Set(true)
			case Hidden:
				textField.Visible().Set(false)
			}
		})

		return formElement[Model]{
			Component: component,
			FromModel: func(model Model) {
				component.Value().Set(f.Stringer(model))
			},
			IntoModel: func(model Model) (Model, error) {
				return field.IntoModel(model, component.Value().Get())
			},

			SetError: func(err string) {
				component.Error().Set(err)
			},
		}
	}

	b.fields = append(b.fields, f)
}

type Form[Model any] struct {
	Component core.Component
	Fields    []formElement[Model]
}

func (b *Binding[T]) NewForm(variant RenderVariant) Form[T] {
	var fields []formElement[T]
	root := ui.NewVStack(func(hstack *ui.FlexContainer) {
		for _, field := range b.fields {
			hint := field.RenderHints[variant]
			allocField := field.FormFactory(hint)
			fields = append(fields, allocField)
			hstack.Append(allocField.Component)
		}
	})

	return Form[T]{
		Component: root,
		Fields:    fields,
	}

}

type Field[Model, Presentation any] struct {
	Caption     string               // e.g. a caption for a Column or Field rendering
	Compare     func(a, b Model) int // Compares the two for ordering, if applicable, otherwise nil
	RenderHints RenderHints
	//Action      func(Model) error // an arbitrary action, instead of a field Rendering
	//Validate    func(Model) error // an arbitrary validation callback. The returned error is shown in the UI
	IntoModel func(model Model, value Presentation) (Model, error)
	FromModel func(Model) Presentation
	Stringer  func(Model) string
}

type RenderVariant int

const (
	Overview RenderVariant = iota + 1
	Create
	Update
)

type RenderHints map[RenderVariant]RenderHint

type RenderHint int

const (
	Visible RenderHint = iota
	ReadOnly
	Hidden
)

type anyField[T any] struct {
	Caption      string
	FormFactory  func(variant RenderHint) formElement[T] // we need to allocate that individually
	RenderHints  RenderHints
	Sortable     bool
	CompareField func(a, b T) int
	IntoModel    func(model T, value any) (T, error)
	FromModel    func(T) any
	Stringer     func(T) string
}

type formElement[Model any] struct {
	Component core.Component
	FromModel func(Model)
	IntoModel func(Model) (Model, error)
	SetError  func(err string)
}
