package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"strings"
)

type Binding[Entity data.Aggregate[ID], ID data.IDType] struct {
	fields []Field[Entity, ID]
}

type Field[Entity data.Aggregate[ID], ID data.IDType] struct {
	Label          string
	RenderEditable func(wnd core.Window, entity Entity, validate func(Entity) (Entity, error)) core.View
	RenderViewOnly func(Entity) core.View
	Comparator     func(a, b Entity) int // not sortable, if nil
	Stringer       func(e Entity) string
}

func (f *Field[Entity, ID]) DisableSort() *Field[Entity, ID] {
	f.Comparator = nil
	return f
}

func NewBinding[Entity data.Aggregate[ID], ID data.IDType]() *Binding[Entity, ID] {
	return &Binding[Entity, ID]{}
}

func makeStateID[Entity data.Aggregate[ID], ID data.IDType](label string, entity Entity) string {
	return fmt.Sprintf("crud-field-%T-%v.%s", entity, entity.Identity(), label)
}

func Text[Entity data.Aggregate[ID], ID data.IDType, T ~string](label string, property func(*Entity) *T) Field[Entity, ID] {
	return Field[Entity, ID]{
		Label: label,
		RenderEditable: func(wnd core.Window, entity Entity, validate func(Entity) (Entity, error)) core.View {
			state := core.StateOf[string](wnd, makeStateID(label, entity)).From(func() string {
				return string(*property(&entity))
			})

			return ui.TextField(label, state.String()).InputValue(state)
		},
		RenderViewOnly: func(e Entity) core.View {
			v := *property(&e)
			return ui.Text(string(v))
		},
		CompareField: func(a, b Entity) int {
			av := *property(&a)
			bv := *property(&b)
			return strings.Compare(string(av), string(bv))
		},
		Stringer: func(e Entity) string {
			return string(*property(&e))
		},
	}
}
