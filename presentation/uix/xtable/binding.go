package xtable

import "log/slog"

type SortOrder int

const (
	None SortOrder = iota
	Asc
	Desc
)

type Binding[T any] struct {
	elems []dynColumn
}

func (b *Binding[T]) byCaption(caption string) (dynColumn, bool) {
	for _, elem := range b.elems {
		if elem.Caption == caption {
			return elem, true
		}
	}

	return dynColumn{}, false
}

type Column[Aggregate any] struct {
	Caption      string
	Sortable     bool
	MapField     func(Aggregate) string
	CompareField func(a, b Aggregate) int
}

type dynColumn struct {
	Caption      string
	Sortable     bool
	MapField     func(any) string
	CompareField func(a, b any) int
}

// NewBinding create a custom model-column binding configuration.
// See also [NewModelBinding] and choose what suits best.
func NewBinding[T any]() *Binding[T] {
	b := &Binding[T]{}
	return b
}

func (b *Binding[T]) AddColumn(opts Column[T]) *Binding[T] {
	b.elems = append(b.elems, dynColumn{
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
