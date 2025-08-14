// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package core

import (
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"strconv"

	"go.wdy.de/nago/presentation/proto"
)

// A NavigationPath defines a unique declaration to a root view.
// When initializing the window to render the associated root view, additional [Values] can be submitted, but these
// are never part of the navigation path.
// Technically, this may be serialized data from e.g. an Android Intent or just get-parameters from a Web-URL.
// Reserved paths:
//   - Main Activity resp. index page: .
//   - Fallback resp. not found page: _
//
// There is also a wildcard support to capture multiple pages at once, independently of many suffix elements are
// appended. A wildcard can only occur at the end of a path.
//
// Valid examples:
//   - a
//   - a/b/c
//   - a/b/c/*
//
// Invalid examples:
//
//	-
//	- a?b=c&d=e
//	- \a b \c
//	- a/*/b
type NavigationPath string

func intoStrSlice[A, B ~string](in []A) []B {
	tmp := make([]B, len(in))
	for i, v := range in {
		tmp[i] = B(v)
	}

	return tmp
}

// Values contains string serialized key-value pairs.
// See also UnmarshalValues.
type Values map[string]string

func newValuesFromProto(v proto.RootViewParameters) Values {
	tmp := make(Values, len(v))
	for k, v := range v {
		decV, err := url.QueryUnescape(string(v))
		if err != nil {
			slog.Error("failed to unescape values from proto", "key", k, "value", v, "error", err)
			tmp[string(k)] = string(v)
			continue
		}

		tmp[string(k)] = decV
	}

	return tmp
}

func (v Values) proto() proto.RootViewParameters {
	if v == nil {
		return nil
	}

	tmp := make(proto.RootViewParameters, len(v))
	for k, v := range v {
		tmp[proto.Str(k)] = proto.Str(url.QueryEscape(v))
	}

	return tmp
}

func (v Values) URLEncode() string {
	tmp := url.Values{}
	for key, value := range v {
		tmp.Add(key, value)
	}

	return tmp.Encode()
}

func (v Values) Clone() Values {
	tmp := make(Values, len(v))
	for k, v := range v {
		tmp[k] = v
	}

	return tmp
}

// Put clones the current values and updates the given key-value-pair.
func (v Values) Put(key string, value string) Values {
	tmp := v.Clone()
	tmp[key] = value
	return tmp
}

// UnmarshalValues takes a Values type and tries to deserialize the fields. Supported fields with underlying field types are
//   - string
//   - int
//   - int64
//   - uint64
//   - float64
//   - bool
//
// Alternate names can be unmarshalled using a name field tag.
func UnmarshalValues[Dst any](src Values) (Dst, error) {
	var params Dst
	t := reflect.TypeOf(params)
	v := reflect.ValueOf(&params).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if n, ok := f.Tag.Lookup("name"); ok {
			name = n
		}

		value, ok := src[name]
		if !ok {
			continue
		}

		switch f.Type.Kind() {
		case reflect.String:
			v.Field(i).SetString(value)
		case reflect.Int:
			x, err := strconv.Atoi(value)
			if err != nil {
				slog.Default().Error("cannot parse integer value variable", slog.Any("err", err))
			}

			v.Field(i).SetInt(int64(x))
		case reflect.Int64:
			x, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				slog.Default().Error("cannot parse integer value variable", slog.Any("err", err))
			}

			v.Field(i).SetInt(x)

		case reflect.Uint64:
			x, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				slog.Default().Error("cannot parse integer value variable", slog.Any("err", err))
			}

			v.Field(i).SetUint(x)
		case reflect.Float64:
			x, err := strconv.ParseFloat(value, 64)
			if err != nil {
				slog.Default().Error("cannot parse float value variable", slog.Any("err", err))
			}
			v.Field(i).SetFloat(x)
		case reflect.Bool:
			x, err := strconv.ParseBool(value)
			if err != nil {
				slog.Default().Error("cannot parse bool value variable", slog.Any("err", err))
			}
			v.Field(i).SetBool(x)
		default:
			return params, fmt.Errorf("cannot parse '%s' into %T.%s", value, params, f.Name)
		}

	}

	return params, nil
}
