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
		Caption:  field.Caption,
		Stringer: field.Stringer,
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

	f.Form = formElement[Model]{
		Component: ui.NewTextField(func(textField *ui.TextField) {
			textField.Label().Set(f.Caption)
		}),
		FromModel: func(model Model) {
			f.Form.Component.(*ui.TextField).Value().Set(f.Stringer(model))
		},
		IntoModel: func(model Model) (Model, error) {
			tf := f.Form.Component.(*ui.TextField)
			return field.IntoModel(model, tf.Value().Get())
		},

		SetError: func(err string) {
			f.Form.Component.(*ui.TextField).Error().Set(err)
		},
	}

	b.fields = append(b.fields, f)
}

func (b *Binding[T]) Form() core.Component {
	return ui.NewHStack(func(hstack *ui.FlexContainer) {
		for _, field := range b.fields {
			hstack.Append(field.Form.Component)
		}
	})

}

type Field[Model, Presentation any] struct {
	Caption   string               // e.g. a caption for a Column or Field rendering
	Compare   func(a, b Model) int // Compares the two for ordering, if applicable, otherwise nil
	ReadOnly  bool
	Hidden    bool
	Action    func(Model) error // an arbitrary action, instead of a field Rendering
	Validate  func(Model) error // an arbitrary validation callback. The returned error is shown in the UI
	IntoModel func(model Model, value Presentation) (Model, error)
	FromModel func(Model) Presentation
	Stringer  func(Model) string
}

type anyField[T any] struct {
	Caption      string
	Form         formElement[T]
	ReadOnly     bool
	Hidden       bool
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
