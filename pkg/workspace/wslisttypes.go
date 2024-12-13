package workspace

import (
	"go.wdy.de/nago/auth"
	"maps"
	"slices"
)

var permListTypes = annotation.Permission[ListTypes]("de.worldiety.nago.workspace.types.list")

// ListTypes returns all available types sorted by type.
type ListTypes func(subject auth.Subject) ([]Type, error)

func NewListTypes(repo Repository) ListTypes {
	return func(subject auth.Subject) ([]Type, error) {
		if err := subject.Audit(permList.Identity()); err != nil {
			return nil, err
		}

		unique := map[Type]bool{}
		for ws, err := range repo.All() {
			if err != nil {
				return nil, err
			}

			unique[ws.Type] = true
		}

		types := slices.Collect(maps.Keys(unique))
		slices.Sort(types)

		return types, nil
	}
}
