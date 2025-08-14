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

	"go.wdy.de/nago/pkg/std/concurrent"
)

var envRegistry concurrent.RWMap[Typename, reflect.Type]

type Typename string

func NewTypename(r reflect.Type) Typename {
	if r == nil {
		return ""
	}

	return Typename(fmt.Sprintf("%s.%s", r.PkgPath(), r.Name()))
}

func RegisterFor[T any]() {
	rtype := reflect.TypeFor[T]()
	Register(rtype)
}

func Register(rtype reflect.Type) {
	envRegistry.Put(NewTypename(rtype), rtype)
}

// AdjacentEnvelope encodes the type in an envelope as t and the content as c like the haskel default. This default
// is hardcoded.
type AdjacentEnvelope struct {
	Type  Typename `json:"t"`
	Value any      `json:"c"`
}

func NewAdjacentEnvelope(v any) AdjacentEnvelope {
	t := NewTypename(reflect.TypeOf(v))
	if _, ok := envRegistry.Get(t); !ok {
		envRegistry.Put(t, reflect.TypeOf(v))
	}

	return AdjacentEnvelope{
		Type:  t,
		Value: v,
	}
}

func (e *AdjacentEnvelope) UnmarshalJSON(bytes []byte) error {
	type discriminator struct {
		Type    Typename        `json:"t"`
		Content json.RawMessage `json:"c"`
	}

	var tmp discriminator
	if err := json.Unmarshal(bytes, &tmp); err != nil {
		return err
	}

	if tmp.Type == "" {
		e.Value = nil
		e.Type = ""
		return nil
	}

	rtype, ok := envRegistry.Get(tmp.Type)
	if !ok {
		return fmt.Errorf("unknown type: %s", tmp.Type)
	}

	rval := reflect.New(rtype)
	if err := json.Unmarshal(tmp.Content, rval.Interface()); err != nil {
		return err
	}

	e.Type = tmp.Type
	e.Value = rval.Elem().Interface()
	return nil
}
