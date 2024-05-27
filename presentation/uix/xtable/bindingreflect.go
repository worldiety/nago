package xtable

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// NewModelBinding creates automatically columns based on the given ViewModel type.
// The ViewModel can provide the following custom tags:
//   - "caption" provides the actual column name to use
//   - "ignore" removes the column from rendering
//   - "sortable" enables or disables the sorting of the column
//
// See also [NewBinding] for more control.
func NewModelBinding[ViewModel any]() *Binding[ViewModel] {
	b := &Binding[ViewModel]{}
	var zero ViewModel
	rType := reflect.TypeOf(zero)

	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		if _, ok := field.Tag.Lookup("ignore"); ok {
			continue
		}

		caption, ok := field.Tag.Lookup("caption")
		if !ok {
			caption = field.Name
		}

		strSortable := field.Tag.Get("sortable")
		sortable, _ := strconv.ParseBool(strSortable)

		colOpts := AnyColumn{
			Caption:  caption,
			Sortable: sortable,
		}

		switch field.Type.Kind() {
		case reflect.String:
			colOpts.MapField = func(a any) string {
				return reflect.ValueOf(a).Field(i).String()
			}
			colOpts.CompareField = func(a, b any) int {
				x := reflect.ValueOf(a).Field(i).String()
				y := reflect.ValueOf(b).Field(i).String()

				return strings.Compare(x, y)
			}

		case reflect.Int:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			colOpts.MapField = func(a any) string {
				return strconv.FormatInt(reflect.ValueOf(a).Field(i).Int(), 10)
			}
			colOpts.CompareField = func(a, b any) int {
				x := reflect.ValueOf(a).Field(i).Int()
				y := reflect.ValueOf(b).Field(i).Int()
				return int(x - y)
			}

		default:
			colOpts.MapField = func(a any) string {
				return fmt.Sprintf("%v", reflect.ValueOf(a).Field(i).Interface())
			}
			colOpts.CompareField = func(a, b any) int {
				x := reflect.ValueOf(a).Field(i).Interface()
				y := reflect.ValueOf(b).Field(i).Interface()

				return strings.Compare(fmt.Sprintf("%v", x), fmt.Sprintf("%v", y))
			}
		}

		b.Columns = append(b.Columns, colOpts)
	}

	return b
}
