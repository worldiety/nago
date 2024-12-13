package permission

import (
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
	"slices"
)

func NewFindAll() FindAll {
	return func(subject Auditable) iter.Seq2[Permission, error] {
		if err := subject.Audit(PermFindAll); err != nil {
			return xiter.WithError[Permission](err)
		}

		return xslices.ValuesWithError(slices.Collect(All()), nil)
	}
}
