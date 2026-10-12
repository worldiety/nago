// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package secret

import (
	"reflect"
)

// writeOnlyTag marks a string field of a [Credentials] type as write-only, e.g. a password or a token:
//
//	Password string `style:"secret" credentialWriteOnly:"true"`
//
// The secret pages never hand out the stored value of such a field again, see [WithoutWriteOnly], and
// [UpdateMyCredentials] keeps the stored value, if the field is empty, see [keepWriteOnly].
const writeOnlyTag = "credentialWriteOnly"

// WriteOnlyFields returns the names of the write-only fields of the credentials and whether a value is stored in
// them.
func WriteOnlyFields(credentials Credentials) map[string]bool {
	res := map[string]bool{}
	v, ok := structOf(credentials)
	if !ok {
		return res
	}

	for _, idx := range writeOnlyFieldIndices(v.Type()) {
		res[v.Type().Field(idx).Name] = v.Field(idx).String() != ""
	}

	return res
}

// WithoutWriteOnly returns a copy of the credentials, whose write-only fields are empty. Use it, before credentials
// are shown to a user.
func WithoutWriteOnly(credentials Credentials) Credentials {
	v, ok := structOf(credentials)
	if !ok {
		return credentials
	}

	indices := writeOnlyFieldIndices(v.Type())
	if len(indices) == 0 {
		return credentials
	}

	cpy := reflect.New(v.Type()).Elem()
	cpy.Set(v)
	for _, idx := range indices {
		cpy.Field(idx).SetString("")
	}

	return cpy.Interface().(Credentials)
}

// keepWriteOnly returns the updated credentials, whose empty write-only fields have the values of the stored
// credentials. Thus, a form which has never received the stored values does not delete them.
func keepWriteOnly(stored, updated Credentials) Credentials {
	sv, ok := structOf(stored)
	if !ok {
		return updated
	}

	uv, ok := structOf(updated)
	if !ok || sv.Type() != uv.Type() {
		return updated
	}

	indices := writeOnlyFieldIndices(uv.Type())
	if len(indices) == 0 {
		return updated
	}

	cpy := reflect.New(uv.Type()).Elem()
	cpy.Set(uv)
	for _, idx := range indices {
		if cpy.Field(idx).String() == "" {
			cpy.Field(idx).SetString(sv.Field(idx).String())
		}
	}

	return cpy.Interface().(Credentials)
}

func structOf(credentials Credentials) (reflect.Value, bool) {
	if credentials == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(credentials)
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	return v, true
}

func writeOnlyFieldIndices(t reflect.Type) []int {
	var res []int
	for i := range t.NumField() {
		f := t.Field(i)
		if f.IsExported() && f.Type.Kind() == reflect.String && f.Tag.Get(writeOnlyTag) == "true" {
			res = append(res, i)
		}
	}

	return res
}
