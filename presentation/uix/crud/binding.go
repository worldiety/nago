package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"math"
	slices2 "slices"
	"strconv"
	"strings"
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

func OneToOne[Model any, Foreign data.Aggregate[ForeignKey], ForeignKey data.IDType](b *Binding[Model], foreignKeyIter iter.Seq2[Foreign, error], stringer func(Foreign) string, field Field[Model, []ForeignKey]) {
	oneToN[Model, Foreign, ForeignKey](b, 1, foreignKeyIter, stringer, field)
}

// OneToMany converts foreign key (or ID) values into actual values by collecting them from the given iterator.
func OneToMany[Model any, Foreign data.Aggregate[ForeignKey], ForeignKey data.IDType](b *Binding[Model], foreignKeyIter iter.Seq2[Foreign, error], stringer func(Foreign) string, field Field[Model, []ForeignKey]) {
	oneToN[Model, Foreign, ForeignKey](b, math.MaxInt, foreignKeyIter, stringer, field)
}

func oneToN[Model any, Foreign data.Aggregate[ForeignKey], ForeignKey data.IDType](b *Binding[Model], n int, foreignKeyIter iter.Seq2[Foreign, error], stringer func(Foreign) string, field Field[Model, []ForeignKey]) {
	f := anyField[Model]{
		Caption:     field.Caption,
		RenderHints: field.RenderHints,
	}

	var itemSlice []Foreign
	initDataSet := func() {
		var err error
		itemSlice = slices.Collect(iter.BreakOnError(&err, foreignKeyIter))
		if err != nil {
			slog.Error("cannot get entity slice from iter for foreign keys", "err", err)
			return
		}

		if field.FromModel == nil {
			panic(fmt.Errorf("cannot process OneToMany declaration without FromModel func"))
		}

	}

	strSliceOf := func(model Model) []string {
		var tmp []string
		fks := field.FromModel(model)
		for _, fk := range fks {
			for _, foreign := range itemSlice {
				if foreign.Identity() == fk {
					tmp = append(tmp, stringer(foreign))
				}
			}
		}

		slices2.Sort(tmp)
		return tmp
	}

	if field.Stringer == nil || field.isPtrStringer {
		field.Stringer = func(model Model) string {
			return strings.Join(strSliceOf(model), ", ")
		}
	}

	f.Stringer = field.Stringer

	// we need to init the dataset, because the field members are also used by other views without calling the factory
	initDataSet()

	f.FormFactory = func(variant RenderHint) formElement[Model] {
		// we need to re-init the entire foreign dataset, because it may have been invalidated by the preceeding modification.
		// E.g. a Person with Person as friends will otherwise not update the list on itself
		// TODO this may become very expensive, perhaps we should accept that bug?
		initDataSet()
		component := ui.NewDropdown(func(dropdown *ui.Dropdown) {
			dropdown.Label().Set(f.Caption)
			switch variant {
			case Visible:
				dropdown.Visible().Set(true)
			case ReadOnly:
				dropdown.Disabled().Set(true)
			case Hidden:
				dropdown.Visible().Set(false)
			}

			if n > 1 {
				dropdown.Multiselect().Set(true)
			}

			dropdown.OnClicked().Set(func() {
				dropdown.Expanded().Set(!dropdown.Expanded().Get())
			})

		})

		if len(itemSlice) > 10 {
			component.Searchable().Set(true)
		}

		// always populate items

		var selectedFKs []ForeignKey

		for _, item := range itemSlice {
			component.Items().Append(
				ui.NewDropdownItem(func(dropdownItem *ui.DropdownItem) {
					dropdownItem.Content().Set(stringer(item))
					dropdownItem.OnClicked().Set(func() {
						component.Toggle(dropdownItem)
						selectedFKs = nil

						component.SelectedIndices().Iter(func(i int64) bool {
							selectedFKs = append(selectedFKs, itemSlice[i].Identity())
							return true
						})

					})
				}),
			)
		}

		return formElement[Model]{
			Component: component,
			FromModel: func(model Model) {
				selectedFKs = nil
				for i, item := range itemSlice {
					isSelected := false
					for _, id := range field.FromModel(model) {
						if id == item.Identity() {
							isSelected = true
							break
						}
					}

					if isSelected {
						selectedFKs = append(selectedFKs, itemSlice[i].Identity())
						component.SelectedIndices().Append(int64(i))
					}

				}

			},
			IntoModel: func(model Model) (Model, error) {
				return field.IntoModel(model, selectedFKs)
			},

			SetError: func(err string) {
				component.Error().Set(err)
			},
		}
	}

	b.fields = append(b.fields, f)
}

func Text[Model any, T ~string](b *Binding[Model], field Field[Model, T]) {
	text[Model, T](b, false, field)
}

func Secret[Model any, T ~string](b *Binding[Model], field Field[Model, T]) {
	text[Model, T](b, true, field)
}

func text[Model any, T ~string](b *Binding[Model], isSecret bool, field Field[Model, T]) {
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

	type textOrPassfield interface {
		Value() ui.String
		Error() ui.String
		Visible() ui.Bool
		Disabled() ui.Bool
		Label() ui.String
	}

	f.FormFactory = func(variant RenderHint) formElement[Model] {
		var component textOrPassfield
		if isSecret {
			component = ui.NewPasswordField(nil)
		} else {
			component = ui.NewTextField(nil)
		}

		component.Label().Set(f.Caption)
		switch variant {
		case Visible:
			component.Visible().Set(true)
		case ReadOnly:
			component.Disabled().Set(true)
		case Hidden:
			component.Visible().Set(false)
		}

		return formElement[Model]{
			Component: component.(core.Component),
			FromModel: func(model Model) {
				component.Value().Set(f.Stringer(model))
			},
			IntoModel: func(model Model) (Model, error) {
				return field.IntoModel(model, T(component.Value().Get()))
			},

			SetError: func(err string) {
				component.Error().Set(err)
			},
		}
	}

	b.fields = append(b.fields, f)
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Int[Model any, T Integer](b *Binding[Model], field Field[Model, T]) {
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
		component := ui.NewNumberField(func(textField *ui.NumberField) {
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
				v, err := strconv.ParseInt(component.Value().Get(), 10, 64)
				if err != nil {
					slog.Error("cannot parse number from numberfield", "err", err)
				}
				return field.IntoModel(model, T(v))
			},

			SetError: func(err string) {
				component.Error().Set(err)
			},
		}
	}

	b.fields = append(b.fields, f)
}

func Bool[Model any, T ~bool](b *Binding[Model], field Field[Model, T]) {
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
		component := ui.NewToggle(func(textField *ui.Toggle) {
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
				flag := field.FromModel(model)
				component.Checked().Set(bool(flag))
			},
			IntoModel: func(model Model) (Model, error) {

				return field.IntoModel(model, T(component.Checked().Get()))
			},

			SetError: func(err string) {
				//component.Error().Set(err) TODO currently missing
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
	IntoModel     func(model Model, value Presentation) (Model, error)
	FromModel     func(Model) Presentation
	Stringer      func(Model) string
	isPtrStringer bool
}

func (f *Field[Model, Presentation]) setRenderHints(hints RenderHints) {
	f.RenderHints = hints
}

type FOption interface {
	apply(interface{ setRenderHints(hints RenderHints) })
}

func FromPtr[Model, Presentation any](caption string, property func(*Model) *Presentation, opts ...FOption) Field[Model, Presentation] {

	f := Field[Model, Presentation]{
		Caption: caption,
		Stringer: func(model Model) string {
			return fmt.Sprintf("%v", *property(&model))
		},
		isPtrStringer: true,
		IntoModel: func(model Model, value Presentation) (Model, error) {
			*property(&model) = value
			return model, nil
		},
		FromModel: func(model Model) Presentation {
			return *property(&model)
		},
	}
	for _, opt := range opts {
		opt.apply(&f)
	}

	return f
}

type RenderVariant int

const (
	Overview RenderVariant = iota + 1
	Create
	Update
)

type RenderHints map[RenderVariant]RenderHint

func (r RenderHints) apply(i interface{ setRenderHints(hints RenderHints) }) {
	i.setRenderHints(r)
}

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
