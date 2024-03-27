package xtable

import (
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	slices2 "slices"
)

type row[T any] struct {
	binding *Binding[T]
	values  []string
	model   T
}

func getData[E any](rows iter.Seq2[E, error], binding *Binding[E], settings Settings) ([]row[E], error) {
	// build up functional data pipeline
	var err error
	repoItems := iter.BreakOnError(&err, rows)

	colSeq := iter.Map(func(in E) row[E] {
		res := row[E]{
			binding: binding,
			model:   in,
		}

		for _, binder := range binding.elems {
			res.values = append(res.values, binder.MapField(in))
		}
		return res
	}, repoItems)

	predicate := rquery.SimplePredicate[any](settings.LastQuery)
	filtered := iter.Filter(func(model row[E]) bool {
		return predicate(model.values)
	}, colSeq)

	// until now, not a single I/O has taken place, apply the entire functional pipeline and allocate that memory
	items := slices.Collect(filtered)

	// the error may have manifested
	if err != nil {
		return nil, err
	}

	// apply sorting on in-memory data
	if sortCol, ok := binding.byCaption(settings.SortByColumName); settings.SortByColumName != "" && ok {
		slices2.SortFunc(items, func(a, b row[E]) int {
			dir := 1
			if !settings.SortAsc {
				dir = -1
			}
			return sortCol.CompareField(a.model, b.model) * dir
		})
	}

	return items, nil
}
