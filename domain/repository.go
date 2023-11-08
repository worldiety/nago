package dm

import (
	"cmp"
	"go.wdy.de/nago/container/data"
	"go.wdy.de/nago/container/slice"
)

type Repository[E Entity[ID], ID cmp.Ordered] interface {
	Filter(p func(E) bool) (slice.Slice[E], error)
	Delete(id ID) error
	DeleteByFunc(f func(E) bool) error
	Save(entities ...E) error
	// FindOne either returns the Entity or data.None or fails with a technical error.
	FindOne(id ID) (data.Option[E], error)
}
