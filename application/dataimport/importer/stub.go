// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package importer

import (
	"github.com/worldiety/jsonptr"
	"reflect"
	"strings"
)

// Stub returns an "empty" object with all struct field values set to null.
func Stub(t reflect.Type) *jsonptr.Obj {
	obj := &jsonptr.Obj{}
	stubObj(obj, t)
	return obj
}

func stubObj(dst *jsonptr.Obj, t reflect.Type) {
	if dst == nil || t.Kind() != reflect.Struct {
		return
	}

	for _, field := range reflect.VisibleFields(t) {
		if !field.IsExported() {
			continue
		}

		name := field.Name
		values := strings.Split(field.Tag.Get("json"), ",")
		if values[0] != "" {
			name = values[0]
		}

		if field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Struct {
			obj := &jsonptr.Obj{}
			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}

			stubObj(obj, field.Type)
			dst.Put(name, obj)
		} else {
			dst.Put(name, jsonptr.Null{})
		}

	}
}
