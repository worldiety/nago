package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

// A Field is a binding to a field of T. All members are intentionally public, to make customization as flexible as
// possible in most typical CRUD situations. Try to stick to the prebuild factories:
//   - [Text]
type Field[T any] struct {
	// ID must be a unique identifier, at least when multiple binding are rendered at the same time or if different
	// bindings are displayed immediately after each other. A good candidate may be an entity ID.
	// Otherwise, the internal state handling cannot detect
	// that the bindings are different and will re-use them causing weired state bugs.
	ID string

	// Label of the field, which should be unique in the entire set. Otherwise, accessibility is broken.
	Label string

	// SupportingText is optional.
	SupportingText string

	// RenderFormElement may be nil, if it shall not be shown in a form.
	RenderFormElement func(self Field[T], entity *T) ui.DecoredView

	// RenderTableCell may be nil, if it shall not be shown in tables.
	RenderTableCell func(self Field[T], entity *T) ui.TTableCell

	// RenderCardElement may be nil, if it shall not be shown on a card.
	RenderCardElement func(self Field[T], entity *T) ui.DecoredView

	// Window is needed to hold states while editing, to allow downloads and be responsive.
	// It must not be nil.
	Window core.Window

	// Validate may be nil, if nothing can be validated.
	// The given error will be logged, but the details are never displayed to the user due to possible security concerns.
	// E.g. if the error is passed from the infrastructure layer, it may expose confidential details to an attacker.
	// Return an errorText, if you want to mark the field as errornous.
	Validate func(T) (errorText string, infrastructureError error)

	// Disabled is true, if the element shall be shown but must not be editable.
	Disabled bool

	// Comparator may be nil, if a field type can not be compared or must not be compared.
	Comparator func(a, b T) int

	// Stringer is used to create a string representation of the field. This is used for filtering or
	// other kinds of rendering.
	Stringer func(e T) string
}

func (f Field[T]) WithoutSorting() Field[T] {
	f.Comparator = nil
	return f
}

func (f Field[T]) WithoutCard() Field[T] {
	f.RenderCardElement = nil
	return f
}

func (f Field[T]) WithoutForm() Field[T] {
	f.RenderFormElement = nil
	return f
}

func (f Field[T]) WithoutTable() Field[T] {
	f.RenderTableCell = nil
	return f
}

func (f Field[T]) WithValidation(fn func(T) (errorText string, infrastructureError error)) Field[T] {
	f.Validate = fn
	return f
}

func (f Field[T]) WithSupportingText(str string) Field[T] {
	f.SupportingText = str
	return f
}

type Binding[T any] struct {
	id     string
	wnd    core.Window
	fields []Field[T]
}

// NewBinding allocates a new binding using the given window and id.
// If you bind a specific entity, just use Identity as id.
// If you don't have an identity, it may work, if left empty,
// but ensure you read the doc at [Binding.Add] and understood
// the state-render mechanics to see potential unwanted side effects.
func NewBinding[T any](wnd core.Window, id string) *Binding[T] {
	return &Binding[T]{
		id:  id,
		wnd: wnd,
	}
}

// Add appends the given fields to this binding container.
// If the field ID is empty, an automatic internal ID based on T and the field index
// is assigned, which should normally be enough. However, there may be corner situations like prev/next navigation
// between entities on the same form, which may cause unwanted state re-usage, if you do not provide a unique id
// to the binding constructor.
func (b *Binding[T]) Add(fields ...Field[T]) {
	off := len(b.fields)
	for i, field := range fields {
		if field.ID == "" {
			var zero T
			field.ID = fmt.Sprintf("crud.field.%T@%s.%d", zero, b.id, i+off)
		}
		field.Window = b.wnd
		b.fields = append(b.fields, field)
	}
}
