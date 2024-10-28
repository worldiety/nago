package data

import (
	"go.wdy.de/nago/pkg/std"
	"iter"
)

// ReadOnly wraps a given repository to become a read-only repository, so that even type assertions cannot peek through
// the abstraction to perform writes.
func ReadOnly[E Aggregate[ID], ID IDType](wr Repository[E, ID]) ReadRepository[E, ID] {
	return readOnlyAdapter[E, ID]{wr}
}

type readOnlyAdapter[E Aggregate[ID], ID IDType] struct {
	other ReadRepository[E, ID]
}

func (r readOnlyAdapter[E, ID]) FindByID(id ID) (std.Option[E], error) {
	return r.other.FindByID(id)
}

func (r readOnlyAdapter[E, ID]) FindAllByID(ids iter.Seq[ID]) iter.Seq2[E, error] {
	return r.FindAllByID(ids)
}

func (r readOnlyAdapter[E, ID]) All() iter.Seq2[E, error] {
	return r.other.All()
}

func (r readOnlyAdapter[E, ID]) Count() (int, error) {
	return r.other.Count()
}
