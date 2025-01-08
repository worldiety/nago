package group

import (
	"fmt"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/pkg/xiter"
	"go.wdy.de/nago/pkg/xslices"
	"iter"
)

func NewFindMyGroups(repository Repository) FindMyGroups {
	return func(subject permission.Auditable) iter.Seq2[Group, error] {
		type groupOwner interface {
			HasGroup(ID) bool
			Groups() iter.Seq[ID]
		}

		owner, ok := subject.(groupOwner)
		if !ok {
			return xiter.WithError[Group](fmt.Errorf("subject %T is not GroupOwner", subject))
		}

		var tmp []Group
		for id := range owner.Groups() {
			optGroup, err := repository.FindByID(id)
			if err != nil {
				return xiter.WithError[Group](err)
			}

			if optGroup.IsSome() {
				tmp = append(tmp, optGroup.Unwrap())
			}
		}

		return xslices.Values2[[]Group, Group, error](tmp)
	}
}
