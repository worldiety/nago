package crud

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/image"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/timepicker"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type AutoBindingOptions struct {
	ButtonEditForwardTo core.NavigationPath
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

			if field.Name == "_" {
				fieldsBuilder.Append(Label[E](label))
				continue
			}

			var values []string
			if array, ok := field.Tag.Lookup("values"); ok {
				err := json.Unmarshal([]byte(array), &values)
				if err != nil {
					slog.Error("cannot parse values from struct field", "type", reflect.TypeFor[E]().String(), "field", field.Name, "literal", array, "err", err)
				}
			}

			switch field.Type.Kind() {
			case reflect.Slice:
				switch field.Type.Elem().Kind() {
				case reflect.String:

					source, ok := field.Tag.Lookup("source")
					if ok {
						listAll, ok := core.SystemServiceWithName[UseCaseListAny](wnd.Application(), source)
						if !ok {
							slog.Error("can not find list by system service", "source", source)
							continue
						}

						fieldsBuilder.Append(OneToMany[E](OneToManyOptions[AnyEntity, string]{Label: label, ForeignEntities: listAll(wnd.Subject())}, PropertyFuncs(
							func(obj *E) []string {
								slice := reflect.ValueOf(obj).Elem().FieldByName(field.Name)
								tmp := make([]string, 0, slice.Len())
								for i := 0; i < slice.Len(); i++ {
									tmp = append(tmp, slice.Index(i).String())
								}

								return tmp
							},
							func(dst *E, v []string) {
								slice := reflect.MakeSlice(field.Type, 0, len(v))
								for _, strVal := range v {
									newValue := reflect.New(field.Type.Elem()).Elem()
									newValue.SetString(strVal)
									slice = reflect.Append(slice, newValue)
								}

								reflect.ValueOf(dst).Elem().FieldByName(field.Name).Set(slice)
							},
						)))
					}

				}
			case reflect.Struct:
				switch field.Type {
				case reflect.TypeFor[xtime.TimeFrame]():
					fieldsBuilder.Append(TimeFrame(TimeFrameOptions{Label: label, Location: time.Local}, PropertyFuncs(
						func(e *E) xtime.TimeFrame {
							value := reflect.ValueOf(e).Elem().FieldByName(field.Name).Interface()
							return value.(xtime.TimeFrame)
						},
						func(dst *E, v xtime.TimeFrame) {
							value := reflect.ValueOf(v)
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).Set(value)
						},
					)))

				case reflect.TypeFor[xtime.Date]():
					fieldsBuilder.Append(Date(DateOptions{Label: label}, PropertyFuncs(
						func(e *E) xtime.Date {
							value := reflect.ValueOf(e).Elem().FieldByName(field.Name).Interface()
							return value.(xtime.Date)
						},
						func(dst *E, v xtime.Date) {
							value := reflect.ValueOf(v)
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).Set(value)
						},
					)))
				}
			case reflect.String:
				switch field.Type {
				case reflect.TypeFor[image.ID]():
					var styleOpts PickOneImageStyle
					switch field.Tag.Get("style") {
					case "avatar":
						styleOpts = PickOneImageStyleAvatar
					default:
						styleOpts = PickOneImageStyleSingle
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
						var styleOpts PickOneStyle
						switch field.Tag.Get("style") {
						case "radio":
							styleOpts = PickOneStyleWithRadioButton
						default:
							styleOpts = PickOneStyleWithPicker
						}

						var pickedEntryList []taggedValueEntry
						for _, value := range values {
							idAndName := strings.Split(value, "=")
							if len(idAndName) == 1 {
								pickedEntryList = append(pickedEntryList, taggedValueEntry{id: idAndName[0], text: idAndName[0]})
							} else {
								pickedEntryList = append(pickedEntryList, taggedValueEntry{id: idAndName[0], text: idAndName[1]})
							}
						}

						fieldsBuilder.Append(PickOne[E, taggedValueEntry](PickOneOptions[taggedValueEntry]{Label: label, Values: pickedEntryList, Style: styleOpts}, PropertyFuncs(
							func(obj *E) std.Option[taggedValueEntry] {
								fmt.Println(obj)
								id := reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
								for _, entry := range pickedEntryList {
									if entry.id == id {
										return std.Some(entry)
									}
								}

								return std.None[taggedValueEntry]()
							}, func(dst *E, v std.Option[taggedValueEntry]) {
								id := v.UnwrapOr(taggedValueEntry{}).id
								reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(id)
							})))
					} else {

						source, ok := field.Tag.Lookup("source")
						if ok {
							listAll, ok := core.SystemServiceWithName[UseCaseListAny](wnd.Application(), source)
							if !ok {
								slog.Error("can not find list by system service", "source", source)
								continue
							}

							fieldsBuilder.Append(OneToOne[E](OneToOneOptions[AnyEntity, string]{Label: label, ForeignEntities: listAll(wnd.Subject())}, PropertyFuncs(
								func(obj *E) std.Option[string] {
									v := reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
									if v == "" {
										return std.None[string]()
									}

									return std.Some(v)
								},
								func(dst *E, v std.Option[string]) {
									reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v.UnwrapOr(""))
								},
							)))

						} else {

							var lines int
							if str, ok := field.Tag.Lookup("lines"); ok {
								lines, _ = strconv.Atoi(str)
							}

							switch field.Tag.Get("style") {
							case "secret":
								fieldsBuilder.Append(Password[E, string](PasswordOptions{Label: label}, PropertyFuncs(
									func(obj *E) string {
										return reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
									}, func(dst *E, v string) {
										reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v)
									})))

							default:
								fieldsBuilder.Append(Text[E, string](TextOptions{Label: label, Lines: lines, SupportingText: field.Tag.Get("supportingText")}, PropertyFuncs(
									func(obj *E) string {
										return reflect.ValueOf(obj).Elem().FieldByName(field.Name).String()
									}, func(dst *E, v string) {
										reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetString(v)
									})))

							}

						}
					}

				}

			case reflect.Int:
				fallthrough
			case reflect.Int64:
				switch field.Type {
				case reflect.TypeFor[time.Duration]():
					var displayFormat timepicker.PickerFormat
					switch field.Tag.Get("style") {
					case "decomposed":
						displayFormat = timepicker.DecomposedFormat
					case "clock":
						displayFormat = timepicker.ClockFormat
					}

					fieldsBuilder.Append(Time[E](TimeOptions{Label: label, DisplayFormat: displayFormat}, PropertyFuncs(
						func(obj *E) time.Duration {
							return time.Duration(reflect.ValueOf(obj).Elem().FieldByName(field.Name).Int())
						},
						func(dst *E, v time.Duration) {
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetInt(int64(v))
						})))
				default:
					fieldsBuilder.Append(Int[E, int64](IntOptions{Label: label, SupportingText: field.Tag.Get("supportingText")}, PropertyFuncs(
						func(obj *E) int64 {
							return reflect.ValueOf(obj).Elem().FieldByName(field.Name).Int()
						}, func(dst *E, v int64) {
							reflect.ValueOf(dst).Elem().FieldByName(field.Name).SetInt(v)
						})))
				}

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
		if opts.ButtonEditForwardTo == "" {
			aggregateOpts = append(aggregateOpts, ButtonEdit[E, ID](bnd, func(model E) (errorText string, infrastructureError error) {

				_, err := useCases.Save(wnd.Subject(), model)
				if err != nil {
					return "", err // The UI will hide the error
					// from the user and will show a general tracking.SupportRequestDialog
				}

				return "", nil
			}))
		} else {
			aggregateOpts = append(aggregateOpts, ButtonEditForwardTo[E](bnd, func(wnd core.Window, entity E) {
				wnd.Navigation().ForwardTo(opts.ButtonEditForwardTo, core.Values{"id": data.Idtos(entity.Identity())})
			}))
		}

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
	//

	//typ := reflect.TypeOf(zero)
	//for i := 0; i < typ.NumField(); i++ {
	for _, field := range reflect.VisibleFields(reflect.TypeOf(zero)) {
		//field := typ.Field(i)

		if flag, ok := field.Tag.Lookup("visible"); ok && flag == "false" {
			continue
		}

		if field.Name != "_" && !field.IsExported() {
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

type taggedValueEntry struct {
	id   string
	text string
}

func (e taggedValueEntry) String() string {
	return e.text
}
