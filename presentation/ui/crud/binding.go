package crud

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
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
	RenderFormElement func(self Field[T], entity *core.State[T]) ui.DecoredView // TODO this state does not make any sense, each render scope has its own state anyway

	// RenderTableCell may be nil, if it shall not be shown in tables.
	RenderTableCell func(self Field[T], entity *core.State[T]) ui.TTableCell // TODO this state does not make any sense, each render scope has its own state anyway

	// RenderCardElement may be nil, if it shall not be shown on a card.
	RenderCardElement func(self Field[T], entity *core.State[T]) ui.DecoredView // TODO this state does not make any sense, each render scope has its own state anyway

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

	// may be nil
	parent *Binding[T]

	// metaRefID is used by the grouping fields hack, like [Section]
	metaRefID string
}

func (f Field[T]) requiresValidation() bool {
	return f.parent != nil && f.parent.forceValidation != nil && f.parent.forceValidation.Get()
}

func (f Field[T]) ReadOnly(readOnly bool) Field[T] {
	f.Disabled = readOnly
	return f
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

func (f Field[T]) WithStringer(fn func(e T) string) Field[T] {
	f.Stringer = fn
	return f
}

type Binding[T any] struct {
	id                      string
	wnd                     core.Window
	fields                  []Field[T]
	fieldValidationObserver map[int]func(field Field[T], errorText string, infrastructureError error)
	lastObserverId          int
	forceValidation         *core.State[bool]
}

// NewBinding allocates a new binding using the given window.
// See also [Binding.SetID].
func NewBinding[T any](wnd core.Window) *Binding[T] {
	return &Binding[T]{
		wnd:             wnd,
		forceValidation: core.AutoState[bool](wnd), // TODO this may break to easily in loops etc.
	}
}

// tableFields only returns those fields, which have a table renderer
func (b *Binding[T]) tableFields() []Field[T] {
	res := make([]Field[T], 0, len(b.fields))
	for _, field := range b.fields {
		if field.RenderTableCell != nil {
			res = append(res, field)
		}
	}

	return res
}

func (b *Binding[T]) lastField() std.Option[Field[T]] {
	if len(b.fields) == 0 {
		return std.None[Field[T]]()
	}

	return std.Some(b.fields[len(b.fields)-1])
}

func (b *Binding[T]) mutField(idx int) *Field[T] {
	return &b.fields[idx]
}

func (b *Binding[T]) CountTableColumns() int {
	var i int
	for _, field := range b.fields {
		if field.RenderTableCell != nil {
			i++
		}
	}

	return i
}

// Inherit returns a defensive copy with the new id set.
func (b *Binding[T]) Inherit(id string) *Binding[T] {
	cpy := &Binding[T]{
		id:              id,
		wnd:             b.wnd,
		fields:          make([]Field[T], 0, len(b.fields)),
		forceValidation: b.forceValidation,
	}

	for _, field := range b.fields {
		field.ID = ""
		cpy.Add(field)
	}

	return cpy
}

// SetDisabledByLabel updates all fields named by the given label
func (b *Binding[T]) SetDisabledByLabel(label string, disabled bool) {
	for i, field := range b.fields {
		if field.Label == label {
			b.fields[i].Disabled = disabled
		}
	}
}

// AddFieldValidationObserver is called for each validated field event.
func (b *Binding[T]) AddFieldValidationObserver(onFieldValidated func(field Field[T], errorText string, infrastructureError error)) (remove func()) {
	b.lastObserverId++
	hnd := b.lastObserverId
	if b.fieldValidationObserver == nil {
		b.fieldValidationObserver = make(map[int]func(field Field[T], errorText string, infrastructureError error))
	}

	b.fieldValidationObserver[hnd] = onFieldValidated

	return func() {
		delete(b.fieldValidationObserver, hnd)
	}
}

// ResetValidation switches validation behavior back to default. See [Binding.Validates].
func (b *Binding[T]) ResetValidation() {
	b.forceValidation.Set(false)
}

// Validates triggers the validation of all fields. This is required, to trigger form validation before saving stuff.
// This method triggers from now on immediate field validations. Use [Binding.ResetValidation] to reset to default
// behavior, in which the user has to make the first entering.
func (b *Binding[T]) Validates(value T) bool {
	b.forceValidation.Set(true)

	anyValidationErr := false
	for _, field := range b.fields {
		if fn := field.Validate; fn != nil {
			if msg, err := fn(value); msg != "" || err != nil {
				anyValidationErr = true
				break
			}
		}
	}

	return !anyValidationErr
}

// FieldByLabel returns the first field value which has the given label
func (b *Binding[T]) FieldByLabel(label string) (Field[T], bool) {
	for _, field := range b.fields {
		if field.Label == label {
			return field, true
		}
	}

	return Field[T]{}, false
}

// UpdateFieldByLabel replaces the first field which has the given label with the given field.
func (b *Binding[T]) UpdateFieldByLabel(field Field[T]) {
	for i, f := range b.fields {
		if f.Label == field.Label {
			field.parent = b
			b.fields[i] = field
			return
		}
	}
}

// SetID sets the internal binding id, which is used to render a field binding.
// If you bind a specific entity, just use Identity as id.
// If you don't have an identity, it may work, if left empty,
// but ensure you read the doc at [Binding.Add] and understood
// the state-render mechanics to see potential unwanted side effects.
func (b *Binding[T]) SetID(id string) *Binding[T] {
	b.id = id
	return b
}

// Add appends the given fields to this binding container.
// If the field ID is empty, an automatic internal ID based on T and the field index
// is assigned, which should normally be enough. However, there may be corner situations like prev/next navigation
// between entities on the same form, which may cause unwanted state re-usage, if you do not provide a unique id
// to the binding constructor.
func (b *Binding[T]) Add(fields ...Field[T]) *Binding[T] {
	off := len(b.fields)
	for i, field := range fields {
		if field.ID == "" {
			var zero T
			field.ID = fmt.Sprintf("crud.field.%T@%s.%d", zero, b.id, i+off)
		}
		field.Window = b.wnd
		field.parent = b
		b.fields = append(b.fields, field)
	}

	return b
}

func handleValidation[E any](self Field[E], entity *core.State[E], errMsg *core.State[string]) {
	if self.Validate != nil {
		errText, err := self.Validate(entity.Get())
		if err != nil {
			if errText == "" {
				var tmp [16]byte
				if _, err := rand.Read(tmp[:]); err != nil {
					panic(err)
				}
				incidentToken := hex.EncodeToString(tmp[:])
				errText = fmt.Sprintf("Unerwarteter Infrastrukturfehler: %s", incidentToken)
			}

			slog.Error(errText, "err", err)
		}

		errMsg.Set(errText)

		if self.parent != nil {
			for _, f := range self.parent.fieldValidationObserver {
				f(self, errText, err)
			}

		}
	}
}
