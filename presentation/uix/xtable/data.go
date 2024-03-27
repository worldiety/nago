package xtable

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/pkg/slices"
	"log/slog"
	"reflect"
	slices2 "slices"
	"strings"
)

type MapF[From, To any] func(From) To

type holder[Original, ViewModel any] struct {
	Original  Original
	ViewModel ViewModel
}

func (h holder[Original, ViewModel]) comparator(colName string) func(a, b holder[Original, ViewModel]) int {
	rType := reflect.TypeOf(h.ViewModel)
	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		caption, ok := field.Tag.Lookup("caption")
		if !ok {
			caption = field.Name
		}

		if caption == colName {
			switch field.Type.Kind() {
			case reflect.String:
				return func(a, b holder[Original, ViewModel]) int {
					x := reflect.ValueOf(a.ViewModel).Field(i).String()
					y := reflect.ValueOf(b.ViewModel).Field(i).String()

					return strings.Compare(x, y)
				}

			case reflect.Int:
				fallthrough
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				return func(a, b holder[Original, ViewModel]) int {
					x := reflect.ValueOf(a.ViewModel).Field(i).Int()
					y := reflect.ValueOf(b.ViewModel).Field(i).Int()
					return int(x - y)
				}
			default:
				return func(a, b holder[Original, ViewModel]) int {
					x := reflect.ValueOf(a.ViewModel).Field(i).Interface()
					y := reflect.ValueOf(b.ViewModel).Field(i).Interface()

					return strings.Compare(fmt.Sprintf("%v", x), fmt.Sprintf("%v", y))
				}
			}
		}
	}

	slog.Error("holder reflect comparator has not found column", slog.String("col", colName))

	return func(a, b holder[Original, ViewModel]) int {
		return 0
	}
}

func getData[E data.Aggregate[ID], ID data.IDType, ColumnModel any](repo data.Repository[E, ID], intoModel MapF[E, ColumnModel], opts Options[E, ID], settings Settings) ([]rowData[E, ColumnModel], error) {
	// build up functional data pipeline
	var err error
	repoItems := iter.BreakOnError(&err, repo.Each)
	colSeq := iter.Map(func(in E) holder[E, ColumnModel] {
		return holder[E, ColumnModel]{
			Original:  in,
			ViewModel: intoModel(in),
		}

	}, repoItems)

	predicate := rquery.SimplePredicate[ColumnModel](settings.LastQuery)
	filtered := iter.Filter(func(model holder[E, ColumnModel]) bool {
		return predicate(model.ViewModel)
	}, colSeq)

	// until now, not a single I/O has taken place, apply the entire functional pipeline and allocate that memory
	items := slices.Collect(filtered)

	// the error may have manifested
	if err != nil {
		return nil, err
	}

	// apply sorting on in-memory data
	if settings.SortByColumName != "" {
		comparator := holder[E, ColumnModel]{}.comparator(settings.SortByColumName)
		slices2.SortFunc(items, func(a, b holder[E, ColumnModel]) int {
			dir := 1
			if !settings.SortAsc {
				dir = -1
			}
			return comparator(a, b) * dir
		})
	}

	var res []rowData[E, ColumnModel]
	for _, item := range items {
		cols := getColData(item.ViewModel)
		var row rowData[E, ColumnModel]
		row.holder = item
		for _, col := range cols {
			row.cols = append(row.cols, col)
		}

		res = append(res, row)
	}

	return res, nil
}

type rowData[Original, ViewModel any] struct {
	holder holder[Original, ViewModel]
	cols   []colData
}

type colData struct {
	value string
}
