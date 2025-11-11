// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type VariantOption struct {
	T    reflect.Type
	Name string
}

func Variant[T any](name string) VariantOption {
	return VariantOption{
		T:    reflect.TypeFor[T](),
		Name: name,
	}
}

func UnmarshalInternally(tag string, b []byte, variants ...VariantOption) (any, error) {
	type p0 struct {
		Type string `json:"type"`
	}

	isSlice := len(b) > 0 && b[0] == '['

	var slice []json.RawMessage
	if !isSlice {
		slice = []json.RawMessage{b}
	} else {
		if err := json.Unmarshal(b, &slice); err != nil {
			return nil, err
		}
	}

	var tmp []any

	for _, b := range slice {
		var vName string
		switch tag {
		case "type":
			var peek p0
			if err := json.Unmarshal(b, &peek); err != nil {
				return nil, err
			}

			vName = peek.Type
		default:
			var peek map[string]any
			if err := json.Unmarshal(b, &peek); err != nil {
				return nil, err
			}

			if v, ok := peek[tag]; ok {
				if v, ok := v.(string); ok {
					vName = v
				}
			}
		}

		if vName == "" {
			return nil, fmt.Errorf("no type discrimator in JSON for tag '%s'", tag)
		}

		decoded := false
		for _, variant := range variants {
			if variant.Name == vName {
				obj := reflect.New(variant.T).Interface()
				if err := json.Unmarshal(b, obj); err != nil {
					return nil, err
				}

				tmp = append(tmp, reflect.ValueOf(obj).Elem().Interface())
				decoded = true
				break
			}
		}

		if !decoded {
			return nil, fmt.Errorf("cannot unmarshal unknown variant type '%s'", vName)
		}
	}

	if isSlice {
		return tmp, nil
	}

	return tmp[0], nil
}

func MarshalInternally(tag string, v any, variants ...VariantOption) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	vT := reflect.TypeOf(v)
	switch vT.Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(v)
		tmp := make([]json.RawMessage, slice.Len())
		for i := 0; i < len(tmp); i++ {
			buf, err := marshalObj(tag, slice.Index(i).Interface(), variants...)
			if err != nil {
				return nil, err
			}

			tmp[i] = buf
		}

		return json.Marshal(tmp)
	case reflect.Struct:
		return marshalObj(tag, v, variants...)
	default:
		return nil, fmt.Errorf("cannot marshal JSON object for type '%T'", v)
	}

}

func marshalObj(tag string, v any, variants ...VariantOption) ([]byte, error) {
	obj, err := intoObject(v)
	if err != nil {
		return nil, err
	}

	vT := reflect.TypeOf(v)

	var variant VariantOption
	for _, v := range variants {
		if v.T == vT {
			variant = v
			break
		}
	}

	if variant.Name == "" {
		return nil, fmt.Errorf("cannot marshal undeclared variant type '%T'", v)
	}

	obj[tag] = variant.Name

	return json.Marshal(obj)
}

func intoObject(a any) (map[string]any, error) {
	t := reflect.TypeOf(a)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%T type must be struct", a)
	}

	obj := map[string]any{}
	for _, field := range reflect.VisibleFields(t) {
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		name := field.Name
		omitEmpty := false
		if tag != "" {
			nameParam := strings.Split(tag, ",")
			switch len(nameParam) {
			case 1:
				name = nameParam[0]
			default:
				name = nameParam[0]
				omitEmpty = nameParam[1] == "omitempty"
			}
		}

		value := reflect.ValueOf(a).FieldByIndex(field.Index)
		if value.IsZero() && omitEmpty {
			continue
		}

		obj[name] = value.Interface()
	}

	return obj, nil
}
