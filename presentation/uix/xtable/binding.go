// deprecated: use crud package
package xtable

import "log/slog"

type SortOrder int

const (
	None SortOrder = iota
	Asc
	Desc
)

// deprecated: use crud package
type Binding[T any] struct {
	Columns []AnyColumn
}

func (b *Binding[T]) byCaption(caption string) (AnyColumn, bool) {
	for _, elem := range b.Columns {
		if elem.Caption == caption {
			return elem, true
		}
	}

	return AnyColumn{}, false
}

// deprecated: use crud package
type Column[Aggregate any] struct {
	Caption      string
	Sortable     bool
	MapField     func(Aggregate) string
	CompareField func(a, b Aggregate) int
}

// deprecated: use crud package
type AnyColumn struct {
	Caption      string
	Sortable     bool
	MapField     func(any) string
	CompareField func(a, b any) int
}

// deprecated: use crud package
// NewBinding create a custom model-column binding configuration.
// See also [NewModelBinding] and choose what suits best.
func NewBinding[T any]() *Binding[T] {
	b := &Binding[T]{}
	return b
}

func (b *Binding[T]) AddColumn(opts Column[T]) *Binding[T] {
	b.Columns = append(b.Columns, AnyColumn{
		Caption:  opts.Caption,
		Sortable: opts.Sortable,
		MapField: func(a any) string {
			if opts.MapField == nil {
				return ""
			}

			return opts.MapField(a.(T))
		},
		CompareField: func(a, b any) int {
			if opts.CompareField == nil {
				slog.Error("cannot compare field because no field comparator has been defined", "field", opts.Caption)
				return 0
			}

			return opts.CompareField(a.(T), b.(T))
		},
	})

	return b
}
