// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package form

import (
	"encoding/json"
	"fmt"
	"go.wdy.de/nago/application/image"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"go.wdy.de/nago/presentation/ui/colorpicker"
	"go.wdy.de/nago/presentation/ui/picker"
	"go.wdy.de/nago/presentation/ui/timepicker"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type AutoOptions struct {
	SectionPadding std.Option[ui.Padding]
	ViewOnly       bool
	IgnoreFields   []string
	Window         core.Window
}

// Auto is similar to [crud.AutoBinding], however it does much less and just creates a form using
// reflection from the given type. It does not require or understand entities and identities.
// Also note, that the concrete type is inspected at runtime and not the given template T, which
// is only needed for your convenience and to satisfy any concrete state type. Internally, everything gets evaluated
// as [any]. T maybe also be an interface, thus ensure, that the state contains not a nil interface.
//
// The current implementation only supports:
//   - string fields
//   - integer fields (literally)
//
// Other features, which are supported by [crud.Auto] are not (yet) supported.
func Auto[T any](opts AutoOptions, state *core.State[T]) ui.DecoredView {
	// TODO can we unify this with the crud package, but it is so different under the hood and equal at the same time?
	value := any(state.Get())
	if value == nil {
		var zero T
		value = zero
	}

	if value == nil {
		return ui.VStack(alert.Banner("implementation error", "no type information available for [form.Auto]"))
	}

	var rootViews xslices.Builder[core.View]
	structType := reflect.TypeOf(value)
	for _, group := range GroupsOf(structType, opts.IgnoreFields...) {
		var fieldsBuilder xslices.Builder[core.View]
		for _, field := range group.Fields {
			/*fieldTableVisible := true
			if flag, ok := field.Tag.Lookup("table-visible"); ok && flag == "false" {
				fieldTableVisible = false
			}*/

			disabled := false
			if flag, ok := field.Tag.Lookup("disabled"); ok && flag == "true" {
				disabled = true
			}

			if opts.ViewOnly {
				disabled = true
			}

			label := field.Name
			if name, ok := field.Tag.Lookup("label"); ok {
				label = name
			}

			if label == "---" {
				fieldsBuilder.Append(ui.HLine())
				continue
			}

			if strings.HasPrefix(field.Name, "_") && label != "_" {
				fieldsBuilder.Append(ui.Text(label).FullWidth().TextAlignment(ui.TextAlignStart))
				continue
			} else if label == "_" {
				continue
			}

			id := field.Tag.Get("id")

			var values []string
			if array, ok := field.Tag.Lookup("values"); ok {
				err := json.Unmarshal([]byte(array), &values)
				if err != nil {
					slog.Error("cannot parse values from struct field", "type", structType.String(), "field", field.Name, "literal", array, "err", err)
				}
			}

			switch field.Type.Kind() {
			case reflect.Slice:
				switch field.Type.Elem().Kind() {
				case reflect.String:
					source, ok := field.Tag.Lookup("source")
					if ok {
						if opts.Window == nil {
							slog.Error("form.Auto requires AutoOptions.Window but is nil")
						}

						listAll, ok := core.SystemServiceWithName[UseCaseListAny](opts.Window.Application(), source)
						if !ok {
							slog.Error("can not find list by system service", "source", source)
							continue
						}

						values, err := xslices.Collect2(listAll(opts.Window.Subject()))
						if err != nil {
							slog.Error("can not collect list", "source", source, "err", err)
						}

						strState := core.DerivedState[[]AnyEntity](state, field.Name).Init(func() []AnyEntity {
							src := state.Get()
							slice := reflect.ValueOf(src).FieldByName(field.Name)
							tmp := make([]AnyEntity, 0, slice.Len())
							for _, id := range slice.Seq2() {
								id := id.String()

								for _, v := range values {
									if v.id == id {
										tmp = append(tmp, v)
										break
									}
								}
							}

							return tmp
						})

						strState.Observe(func(v []AnyEntity) {
							slice := reflect.MakeSlice(field.Type, 0, len(v))
							for _, strVal := range v {
								newValue := reflect.New(field.Type.Elem()).Elem()
								newValue.SetString(strVal.id)
								slice = reflect.Append(slice, newValue)
							}

							dst := state.Get()
							dst = setFieldValue(dst, field.Name, slice.Interface()).(T)
							state.Set(dst)
							state.Notify()
						})

						fieldsBuilder.Append(picker.Picker[AnyEntity](label, values, strState).
							Title(label).
							MultiSelect(true).
							Disabled(disabled).
							SupportingText(field.Tag.Get("supportingText")).
							Frame(ui.Frame{}.FullWidth()))

					} else {
						// just show a multi line textfield
						var lines int
						if str, ok := field.Tag.Lookup("lines"); ok {
							lines, _ = strconv.Atoi(str)
						}

						if lines == 0 {
							lines = 5
						}

						requiresInit := false
						strState := core.DerivedState[string](state, field.Name).Init(func() string {
							src := state.Get()
							slice := reflect.ValueOf(src).FieldByName(field.Name)
							tmp := make([]string, 0, slice.Len())
							for i := 0; i < slice.Len(); i++ {
								tmp = append(tmp, slice.Index(i).String())
							}

							str := strings.Join(tmp, "\n")

							if val := field.Tag.Get("value"); val != "" && str == "" {
								requiresInit = true
								return val
							}

							return str
						})

						strState.Observe(func(newValue string) {
							v := strings.Split(newValue, "\n")
							slice := reflect.MakeSlice(field.Type, 0, len(v))
							for _, strVal := range v {
								newValue := reflect.New(field.Type.Elem()).Elem()
								newValue.SetString(strVal)
								slice = reflect.Append(slice, newValue)
							}

							dst := state.Get()
							dst = setFieldValue(dst, field.Name, slice.Interface()).(T)
							state.Set(dst)
							state.Notify()
						})

						if requiresInit {
							strState.Notify()
						}

						fieldsBuilder.Append(ui.TextField(label, strState.Get()).
							InputValue(strState).
							ID(id).
							SupportingText(field.Tag.Get("supportingText")).
							Lines(lines).
							Disabled(disabled).
							Frame(ui.Frame{}.FullWidth()),
						)
					}

				}
			case reflect.Bool:
				requiresInit := false
				boolState := core.DerivedState[bool](state, field.Name).Init(func() bool {
					src := state.Get()
					v := reflect.ValueOf(src).FieldByName(field.Name).Bool()
					if val := field.Tag.Get("value"); val != "" && v == false {
						p, err := strconv.ParseBool(val)
						if err == nil {
							requiresInit = true
							return p
						}
					}

					return v
				})

				boolState.Observe(func(newValue bool) {
					dst := state.Get()
					dst = setFieldValue(dst, field.Name, newValue).(T)
					state.Set(dst)
					state.Notify()
				})

				if requiresInit {
					boolState.Notify()
				}

				fieldsBuilder.Append(ui.CheckboxField(label, boolState.Get()).
					Disabled(disabled).
					ID(id).
					InputValue(boolState).
					SupportingText(field.Tag.Get("supportingText")).
					Frame(ui.Frame{}.FullWidth()),
				)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch field.Type {
				case reflect.TypeFor[time.Duration]():
					var displayFormat timepicker.PickerFormat
					switch field.Tag.Get("style") {
					case "decomposed":
						displayFormat = timepicker.DecomposedFormat
					case "clock":
						displayFormat = timepicker.ClockFormat
					}

					requiresInit := false
					intState := core.DerivedState[time.Duration](state, field.Name).Init(func() time.Duration {
						src := state.Get()
						v := reflect.ValueOf(src).FieldByName(field.Name).Int()
						if val := field.Tag.Get("value"); val != "" && v == 0 {
							p, err := strconv.ParseInt(val, 10, 64)
							if err == nil {
								requiresInit = true
								return time.Duration(p)
							}
						}

						return time.Duration(v)
					})

					intState.Observe(func(newValue time.Duration) {
						dst := state.Get()
						dst = setFieldValue(dst, field.Name, newValue).(T)
						state.Set(dst)
						state.Notify()
					})

					if requiresInit {
						intState.Notify()
					}

					showDays := true
					if v, ok := field.Tag.Lookup("days"); ok {
						showDays, _ = strconv.ParseBool(v)
					}

					showHours := true
					if v, ok := field.Tag.Lookup("hours"); ok {
						showHours, _ = strconv.ParseBool(v)
					}

					showMinutes := true
					if v, ok := field.Tag.Lookup("minutes"); ok {
						showMinutes, _ = strconv.ParseBool(v)
					}

					showSeconds := true
					if v, ok := field.Tag.Lookup("seconds"); ok {
						showSeconds, _ = strconv.ParseBool(v)
					}

					fieldsBuilder.Append(timepicker.Picker(label, intState).
						Format(displayFormat).
						Days(showDays).
						Hours(showHours).
						Minutes(showMinutes).
						Seconds(showSeconds).
						Disabled(disabled).
						SupportingText(field.Tag.Get("supportingText")).
						Frame(ui.Frame{}.FullWidth()),
					)
				default:
					requiresInit := false
					intState := core.DerivedState[int64](state, field.Name).Init(func() int64 {
						src := state.Get()
						v := reflect.ValueOf(src).FieldByName(field.Name).Int()
						if val := field.Tag.Get("value"); val != "" && v == 0 {
							p, err := strconv.ParseInt(val, 10, 64)
							if err == nil {
								requiresInit = true
								return p
							}
						}

						return v
					})

					intState.Observe(func(newValue int64) {
						dst := state.Get()
						dst = setFieldValue(dst, field.Name, newValue).(T)
						state.Set(dst)
						state.Notify()
					})

					if requiresInit {
						intState.Notify()
					}

					fieldsBuilder.Append(ui.IntField(label, intState.Get(), intState).
						Disabled(disabled).
						SupportingText(field.Tag.Get("supportingText")).
						Frame(ui.Frame{}.FullWidth()),
					)
				}
			case reflect.Struct:
				switch field.Type {
				case reflect.TypeFor[xtime.Date]():

					dateState := core.DerivedState[xtime.Date](state, field.Name).Init(func() xtime.Date {
						src := state.Get()
						v := reflect.ValueOf(src).FieldByName(field.Name).Interface()

						return v.(xtime.Date)
					})

					dateState.Observe(func(newValue xtime.Date) {
						dst := state.Get()
						dst = setFieldValue(dst, field.Name, newValue).(T)
						state.Set(dst)
						state.Notify()
					})

					fieldsBuilder.Append(ui.SingleDatePicker(label, dateState.Get(), dateState).
						Disabled(disabled).
						SupportingText(field.Tag.Get("supportingText")).
						Frame(ui.Frame{}.FullWidth()),
					)

				}
			case reflect.String:
				switch field.Type {
				case reflect.TypeFor[ui.Color]():
					requiresInit := false
					colorState := core.DerivedState[ui.Color](state, field.Name).Init(func() ui.Color {
						src := state.Get()
						str := reflect.ValueOf(src).FieldByName(field.Name).String()
						if val := field.Tag.Get("value"); val != "" && str == "" {
							requiresInit = true
							return ui.Color(val)
						}

						return ui.Color(str)
					})

					colorState.Observe(func(newValue ui.Color) {
						dst := state.Get()
						dst = setFieldValue(dst, field.Name, newValue).(T)
						state.Set(dst)
						state.Notify()
					})

					if requiresInit {
						colorState.Notify()
					}

					fieldsBuilder.Append(colorpicker.PalettePicker(label, colorpicker.DefaultPalette).State(colorState).Value(colorState.Get()))
				case reflect.TypeFor[image.ID]():

					if opts.Window == nil {
						fieldsBuilder.Append(ui.Text("image.ID not rendered: no window available"))
						continue
					}

					requiresInit := false
					imageState := core.DerivedState[image.ID](state, field.Name).Init(func() image.ID {
						src := state.Get()
						str := reflect.ValueOf(src).FieldByName(field.Name).String()
						if val := field.Tag.Get("value"); val != "" && str == "" {
							requiresInit = true
							return image.ID(val)
						}

						return image.ID(str)
					})

					imageState.Observe(func(newValue image.ID) {
						dst := state.Get()
						dst = setFieldValue(dst, field.Name, newValue).(T)
						state.Set(dst)
						state.Notify()
					})

					if requiresInit {
						imageState.Notify()
					}

					if label != "" {
						fieldsBuilder.Append(ui.Text(label).TextAlignment(ui.TextAlignStart).FullWidth())
					}

					if field.Tag.Get("style") == "avatar" {
						fieldsBuilder.Append(AvatarPicker(opts.Window, nil, field.Name, imageState.Get(), imageState, fmt.Sprintf("%v", state.Get())))
					} else {
						fieldsBuilder.Append(SingleImagePicker(opts.Window, nil, nil, nil, field.Name, imageState.Get(), imageState))
					}

				default:
					var lines int
					if str, ok := field.Tag.Lookup("lines"); ok {
						lines, _ = strconv.Atoi(str)
					}

					switch field.Tag.Get("style") {
					case "secret":
						requiresInit := false
						secretState := core.DerivedState[string](state, field.Name).Init(func() string {
							src := state.Get()
							str := reflect.ValueOf(src).FieldByName(field.Name).String()
							if val := field.Tag.Get("value"); val != "" && str == "" {
								requiresInit = true
								return val
							}

							return str
						})

						secretState.Observe(func(newValue string) {
							dst := state.Get()
							dst = setFieldValue(dst, field.Name, newValue).(T)
							state.Set(dst)
							state.Notify()
						})

						if requiresInit {
							secretState.Notify()
						}

						fieldsBuilder.Append(ui.PasswordField(label, secretState.Get()).
							InputValue(secretState).
							ID(id).
							SupportingText(field.Tag.Get("supportingText")).
							Disabled(disabled).
							Frame(ui.Frame{}.FullWidth()),
						)

					default:

						requiresInit := false
						strState := core.DerivedState[string](state, field.Name).Init(func() string {
							src := state.Get()
							str := reflect.ValueOf(src).FieldByName(field.Name).String()
							if val := field.Tag.Get("value"); val != "" && str == "" {
								requiresInit = true
								return val
							}

							return str
						})

						strState.Observe(func(newValue string) {
							dst := state.Get()
							dst = setFieldValue(dst, field.Name, newValue).(T)
							state.Set(dst)
							state.Notify()
						})

						if requiresInit {
							strState.Notify()
						}

						fieldsBuilder.Append(ui.TextField(label, strState.Get()).
							InputValue(strState).
							ID(id).
							SupportingText(field.Tag.Get("supportingText")).
							Lines(lines).
							Disabled(disabled).
							Frame(ui.Frame{}.FullWidth()),
						)

					}
				}
			}
		}

		fields := fieldsBuilder.Collect()
		if group.Name == "" {
			rootViews.Append(fields...)
		} else {
			card := cardlayout.Card(group.Name).Body(ui.VStack(fields...).Gap(ui.L16).FullWidth()).Frame(ui.Frame{}.FullWidth())
			if opts.SectionPadding.IsSome() {
				card = card.Padding(opts.SectionPadding.Unwrap())
			}
			rootViews.Append(card)
		}
	}

	return ui.VStack(rootViews.Collect()...).Gap(ui.L16).FullWidth()
}

func setFieldValue(dst any, fieldName string, val any) any {
	vDst := reflect.ValueOf(dst)

	for vDst.Kind() == reflect.Ptr || vDst.Kind() == reflect.Interface {
		vDst = vDst.Elem()
	}

	cpy := reflect.New(vDst.Type()).Elem()
	cpy.Set(vDst)

	switch t := val.(type) {
	case string:
		cpy.FieldByName(fieldName).SetString(t)
	case int:
		cpy.FieldByName(fieldName).SetInt(int64(t))
	case int64:
		cpy.FieldByName(fieldName).SetInt(t)
	case time.Duration:
		cpy.FieldByName(fieldName).SetInt(int64(t))
	case bool:
		cpy.FieldByName(fieldName).SetBool(t)
	default:
		//slog.Error("cannot set field value for [form.Auto]", "type", reflect.TypeOf(t))
		cpy.FieldByName(fieldName).Set(reflect.ValueOf(t))
	}

	return cpy.Interface()
}
