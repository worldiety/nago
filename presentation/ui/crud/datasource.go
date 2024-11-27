package crud

import (
	"errors"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"iter"
	"slices"
)

type dataSource[Entity data.Aggregate[ID], ID data.IDType] struct {
	it                    iter.Seq2[Entity, error]
	errors                []error
	binding               *Binding[Entity]
	sortByField           *Field[Entity]
	sortOrder             sortDir
	query                 string
	disableDefaultSorting bool
	totalCount            int
}

func (ds *dataSource[Entity, ID]) Error() error {
	if len(ds.errors) > 0 {
		return ds.errors[0]
	}

	return nil
}

func (ds *dataSource[Entity, ID]) PermissionDenied() bool {
	type permissionDeniedEntity interface {
		PermissionDenied() bool
	}

	var perr permissionDeniedEntity
	for _, err := range ds.errors {
		if errors.As(err, &perr) && perr.PermissionDenied() {
			return true
		}
	}

	return false
}

func (ds *dataSource[Entity, ID]) List() []Entity {
	predicate := rquery.SimplePredicate[string](ds.query)
	var res []Entity

	ds.totalCount = 0
	// collect what is possible, e.g. ignore errors for single entries, which may be caused
	// by broken json persistence models
	ds.it(func(e Entity, err error) bool {
		ds.totalCount++
		if err != nil {
			// we could argue that we should collect errors per id, however general IO errors
			// are not id related, and also AnyEntity is totally undefined in error case.
			ds.errors = append(ds.errors, err)
		} else {
			if ds.query != "" {
				// only add those entries which apply to the given predicate filter
				for _, field := range ds.binding.fields {
					if field.Stringer != nil {
						// micro-optimization to avoid unneeded string allocations for non-filter conditions
						str := field.Stringer(e)
						if predicate(str) {
							res = append(res, e)
							break
						}
					}
				}
			} else {
				res = append(res, e)
			}

		}

		return true
	})

	// only apply sorting on the pre-filtered entries, if applicable
	if ds.sortByField != nil && ds.sortByField.Comparator != nil {
		slices.SortFunc(res, func(a, b Entity) int {
			v := ds.sortByField.Comparator(a, b)
			if v != 0 {
				return v
			}

			// the fields are equal, thus ensure that we have a stable return order for each rendering
			if a.Identity() > b.Identity() {
				return 1
			} else if a.Identity() < b.Identity() {
				return -1
			}

			return 0 // actually, can only happen for non-unique iterator
		})

		if !ds.sortOrder {
			slices.Reverse(res)
		}
	} else if !ds.disableDefaultSorting {
		// ensure that we have a stable return order for each rendering, e.g. map-based sources have a random order
		slices.SortFunc(res, func(a, b Entity) int {
			if a.Identity() > b.Identity() {
				return 1
			} else if a.Identity() < b.Identity() {
				return -1
			}

			return 0 // actually, can only happen for non-unique iterator
		})
	}

	return res
}
