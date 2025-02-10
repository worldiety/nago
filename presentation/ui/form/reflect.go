package form

import (
	"encoding/json"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/cardlayout"
	"log/slog"
	"reflect"
	"strconv"
)

type AutoOptions struct {
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
	for _, group := range GroupsOf(structType) {
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

			label := field.Name
			if name, ok := field.Tag.Lookup("label"); ok {
				label = name
			}

			if field.Name == "_" && label != "_" {
				fieldsBuilder.Append(ui.Text(label))
				continue
			} else if label == "_" {
				continue
			}

			var values []string
			if array, ok := field.Tag.Lookup("values"); ok {
				err := json.Unmarshal([]byte(array), &values)
				if err != nil {
					slog.Error("cannot parse values from struct field", "type", structType.String(), "field", field.Name, "literal", array, "err", err)
				}
			}

			switch field.Type.Kind() {
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
					InputValue(boolState).
					SupportingText(field.Tag.Get("supportingText")).
					Frame(ui.Frame{}.FullWidth()),
				)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch field.Type {
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
			case reflect.String:
				switch field.Type {
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
			rootViews.Append(cardlayout.Card(group.Name).Body(ui.VStack(fields...).Gap(ui.L16)).Frame(ui.Frame{}.FullWidth()))
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
	case bool:
		cpy.FieldByName(fieldName).SetBool(t)
	default:
		slog.Error("cannot set field value for [form.Auto]", "type", reflect.TypeOf(t))
	}

	return cpy.Interface()
}
