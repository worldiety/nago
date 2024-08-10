package crud

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strings"
)

type ValidationResult struct {
	infrastructureError error
	validationHint      string
}

// WithError returns a ValidationResult with the state of an error in the infrastructure.
// The given error will be logged, but the details are never displayed to the user due to possible security concerns.
// E.g. if the error is passed from the infrastructure layer, it may expose confidential details to an attacker.
func (r ValidationResult) WithError(err error) ValidationResult {
	r.validationHint = ""
	r.infrastructureError = err
	return r
}

// Ok returns true if neither an infrastructure error has been defined nor a validation hint has been set.
// This is also the zero value.
func (r ValidationResult) Ok() bool {
	return r.infrastructureError == nil && r.validationHint == ""
}

// Error returns nil or an error.
func (r ValidationResult) Error() error {
	return r.infrastructureError
}

// ValidationHint returns either an empty string or a hint to show.
func (r ValidationResult) ValidationHint() string {
	return r.validationHint
}

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
	RenderTableCell func(self Field[T], entity T) ui.TTableCell

	// RenderCardElement may be nil, if it shall not be shown on a card.
	RenderCardElement func(self Field[T], entity T) ui.DecoredView

	// Window is needed to hold states while editing, to allow downloads and be responsive.
	// It must not be nil.
	Window core.Window

	// Validate may be nil, if nothing can be validated.
	Validate func(T) ValidationResult

	// Disabled is true, if the element shall be shown but must not be editable.
	Disabled bool

	// Comparator may be nil, if a field type can not be compared or must not be compared.
	Comparator func(a, b T) int

	// Stringer is used to create a string representation of the field. This is used for filtering or
	// other kinds of rendering.
	Stringer func(e T) string
}

func (f *Field[T]) DisableSorting() *Field[T] {
	f.Comparator = nil
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

func Text[E any, T ~string](label string, property func(*E) *T) Field[E] {
	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *E) ui.DecoredView {
			state := core.StateOf[string](self.Window, self.ID).From(func() string {
				return string(*property(entity))
			})

			state.Observe(func(newValue string) {
				f := property(entity)
				*f = T(newValue)
			})

			return ui.TextField(label, state.String()).InputValue(state)
		},
		RenderTableCell: func(self Field[E], entity E) ui.TTableCell {
			v := *property(&entity)
			return ui.TableCell(ui.Text(string(v)))
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return strings.Compare(string(av), string(bv))
		},
		Stringer: func(e E) string {
			return string(*property(&e))
		},
	}
}
