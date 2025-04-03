// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xreflect

import (
	"reflect"
	"sync"
)

type tagName = string
type tagValue = string
type fieldName = string

var typeFields = map[reflect.Type]map[fieldName]map[tagName]tagValue{}
var typeFieldsMutex sync.RWMutex

// SetFieldTagFor records in a thread safe way in a global map the given triple.
func SetFieldTagFor[T any](fieldname string, tagname string, tagvalue string) {
	typ := reflect.TypeFor[T]()

	typeFieldsMutex.Lock()
	defer typeFieldsMutex.Unlock()

	if _, ok := typeFields[typ]; !ok {
		typeFields[typ] = map[fieldName]map[tagName]tagValue{}
	}

	if typeFields[typ][fieldname] == nil {
		typeFields[typ][fieldname] = map[tagName]tagValue{}
	}

	typeFields[typ][fieldname][tagname] = tagvalue
}

// FieldTagFor returns in a thread safe way the tag value for the given type, field and tag. If no such
// value is found, lookups the original tag value from the field and otherwise returns the empty string.
func FieldTagFor[T any](fieldname string, tagname string) string {
	typ := reflect.TypeFor[T]()

	typeFieldsMutex.RLock()
	defer typeFieldsMutex.RUnlock()

	if v, ok := typeFields[typ]; ok {
		if v, ok := v[fieldname]; ok {
			if v, ok := v[tagname]; ok {
				return v
			}
		}
	}

	if f, ok := typ.FieldByName(fieldname); ok {
		if v, ok := f.Tag.Lookup(tagname); ok {
			return v
		}
	}

	return ""
}
