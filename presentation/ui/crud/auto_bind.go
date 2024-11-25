package crud

import (
	"encoding/json"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"reflect"
)

type AutoBindingOptions struct {
}

// AutoBinding takes the crud use cases and creates a naive binding for it.
// You can additionally tweak the binding, using the following field tags:
//   - label for an alternative name
//   - hidden to omit it completely
//
// To automatically also create a CRUD component e.g. for an entire page, see also [AutoView].
func AutoBinding[E Aggregate[E, ID], ID ~string](opts AutoBindingOptions, wnd core.Window, useCases UseCases[E, ID]) *Binding[E] {
	var zero E
	bnd := NewBinding[E](wnd)
	for _, group := range groupedFields[E]() {
		var fieldsBuilder xslices.Builder[Field[E]]
		for _, field := range group.fields {

			fieldTableVisible := true
			if flag, ok := field.Tag.Lookup("table-visible"); ok && flag == "false" {
				fieldTableVisible = false
			}

			label := field.Name
			if name, ok := field.Tag.Lookup("label"); ok {
				label = name
			}

			var values []string
			if array, ok := field.Tag.Lookup("values"); ok {
				err := json.Unmarshal([]byte(array), &values)
				if err != nil {
					slog.Error("cannot parse values from struct field", "type", reflect.TypeFor[E]().String(), "field", field.Name, "literal", array, "err", err)
				}
			}

			switch field.Type.Kind() {
			case reflect.String:
				switch field.Type {
				case reflect.TypeFor[image.ID]():
					var styleOpts PickOneImageStyle
					switch field.Tag.Get("style") {
					case "avatar":
						styleOpts = PickOneImageStyleAvatar
					default:
						styleOpts = PickOneImageStyleTeaser
					}

					fieldsBuilder.Append(PickOneImage(PickOneImageOptions[E, image.ID]{Label: label, Style: styleOpts}, PropertyFuncs(
						func(e *E) std.Option[image.ID] {
							value := reflect.ValueOf(e).Elem().FieldByName(field.Name).String()
							if value == "" {
								return std.None[image.ID]()
							}

							return std.Some(image.ID(value))
						},
						func(dst *E, v std.Option[image.ID]) {
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(string(v.UnwrapOr("")))
						},
					)))

				case reflect.TypeFor[ui.Color]():
					fieldsBuilder.Append(PickOneColor(PickOneColorOptions{Label: label}, PropertyFuncs(
						func(e *E) std.Option[ui.Color] {
							value := reflect.ValueOf(e).Elem().FieldByName(field.Name).String()
							if value == "" {
								return std.None[ui.Color]()
							}

							return std.Some(ui.Color(value))
						},
						func(dst *E, v std.Option[ui.Color]) {
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(string(v.UnwrapOr("")))
						},
					)))

				default:
					if len(values) > 0 {
						fieldsBuilder.Append(PickOne[E, string](PickOneOptions[string]{Label: label, Values: values}, PropertyFuncs(
							func(obj *E) std.Option[string] {
								return std.Some(reflect.ValueOf(obj).Elem().FieldByName(field.Name).String())
							}, func(dst *E, v std.Option[string]) {
								reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v.UnwrapOr(""))
							})))
					} else {
						fieldsBuilder.Append(Text[E, string](TextOptions{Label: label}, PropertyFuncs(
							func(obj *E) string {
								return reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
							}, func(dst *E, v string) {
								reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v)
							})))
					}

				}

			case reflect.Int:
				fallthrough
			case reflect.Int64:
				fieldsBuilder.Append(Int[E, int64](IntOptions{Label: label}, PropertyFuncs(
					func(obj *E) int64 {
						return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Int()
					}, func(dst *E, v int64) {
						reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetInt(v)
					})))
			case reflect.Float64:
				fieldsBuilder.Append(Float[E, float64](FloatOptions{Label: label}, PropertyFuncs(
					func(obj *E) float64 {
						return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Float()
					}, func(dst *E, v float64) {
						reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetFloat(v)
					})))
			case reflect.Bool:
				fieldsBuilder.Append(Bool[E, bool](BoolOptions{Label: label}, PropertyFuncs(
					func(obj *E) bool {
						return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Bool()
					}, func(dst *E, v bool) {
						reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetBool(v)
					})))
			default:
				slog.Info("unsupported auto binding field type", "type", reflect.TypeOf(zero), "field", field.Name, "type", field.Type)
			}

			if fieldsBuilder.Len() > 0 && !fieldTableVisible {
				if lastField, ok := fieldsBuilder.RemoveLast(); ok {
					lastField.RenderTableCell = nil
					fieldsBuilder.Append(lastField)
				}
			}
		}

		fields := fieldsBuilder.Collect()
		if group.name == "" {
			bnd.Add(fields...)
		} else {
			bnd.Add(Section(group.name, fields...)...)
		}
	}

	var aggregateOpts []ElementViewFactory[E]
	if canSave(bnd, useCases) {
		aggregateOpts = append(aggregateOpts, ButtonEdit[E, ID](bnd, func(model E) (errorText string, infrastructureError error) {

			_, err := useCases.Save(wnd.Subject(), model)
			if err != nil {
				return "", err // The UI will hide the error
				// from the user and will show a general tracking.SupportRequestDialog
			}

			return "", nil
		}))
	}

	if canDelete(bnd, useCases) {
		aggregateOpts = append(aggregateOpts, ButtonDelete[E, ID](wnd, func(model E) error {

			err := useCases.DeleteByID(wnd.Subject(), model.Identity())
			if err != nil {
				return err // The UI will hide the error
				// from the user and will show a general tracking.SupportRequestDialog
			}

			return nil
		}))
	}

	if len(aggregateOpts) > 0 {
		bnd.Add(
			AggregateActions[E]("Optionen", aggregateOpts...),
		)
	}

	return bnd
}

type fieldGroup struct {
	name   string
	fields []reflect.StructField
}

func groupedFields[E any]() []fieldGroup {
	var zero E
	var res []fieldGroup
	for _, field := range reflect.VisibleFields(reflect.TypeOf(zero)) {
		if flag, ok := field.Tag.Lookup("visible"); ok && flag == "false" {
			continue
		}

		sec := field.Tag.Get("section")
		var grp *fieldGroup
		for idx := range res {
			g := &res[idx]
			if g.name == sec {
				grp = g
				break
			}
		}

		if grp == nil {
			res = append(res, fieldGroup{
				name: sec,
			})

			grp = &res[len(res)-1]
		}

		grp.fields = append(grp.fields, field)
	}

	return res
}
