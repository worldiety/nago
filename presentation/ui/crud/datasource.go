package crud

import (
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/iter"
	"slices"
)

type dataSource[Entity data.Aggregate[ID], ID data.IDType] struct {
	it          iter.Seq2[Entity, error]
	errors      []error
	binding     *Binding[Entity, ID]
	sortByField *Field[Entity, ID]
	query       string
}

func (ds *dataSource[Entity, ID]) List() []Entity {
	predicate := rquery.SimplePredicate[string](ds.query)
	var res []Entity
	ds.it(func(e Entity, err error) bool {
		if err != nil {
			ds.errors = append(ds.errors, err)
		} else {
			if ds.query != "" {
				for _, field := range ds.binding.fields {
					if field.Stringer != nil {
						str := field.Stringer(e)
						if predicate(str) {
							res = append(res, e)
						}
					}
				}
			} else {
				res = append(res, e)
			}

		}

		return true
	})

	if ds.sortByField != nil && ds.sortByField.Comparator != nil {
		slices.SortFunc(res, ds.sortByField.Comparator)
	}

	return res
}
