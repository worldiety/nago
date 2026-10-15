// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// The functions of this file convert the values of Go into the neutral values of an engine and back by reflection,
// see [ValueKind], so that a host in Go calls a program and exchanges messages with it without generated code.

// The containers of the library, which convert themselves for a host.
type (
	hostSeq interface {
		HostValues() []any
		HostFromValues(values []any) any
		HostElem() reflect.Type
	}
	hostMap interface {
		HostEntries() [][2]any
		HostFromEntries(entries [][2]any) any
		HostKey() reflect.Type
		HostElem() reflect.Type
	}
)

// typedState is a [State] of any type.
type typedState interface {
	token() Token
}

var (
	goErrorType    = reflect.TypeFor[error]()
	refType        = reflect.TypeFor[js.Ref]()
	tokenType      = reflect.TypeFor[Token]()
	hostSeqType    = reflect.TypeFor[hostSeq]()
	hostMapType    = reflect.TypeFor[hostMap]()
	typedStateType = reflect.TypeFor[typedState]()
)

// TypeName returns the full name of a named type, e.g. "example.com/app.Hit", as the compiler names records,
// unions and states, see [Record.Name].
func TypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.PkgPath() + "." + t.Name()
}

var variants sync.Map // reflect.Type of the interface -> map[string]reflect.Type

// Variants registers the variants of the sealed interface T, so that a host can decode its values, e.g. in a message
// of a program. The compiler checks, that each variant implements T.
//
//	gostrict.Variants[Shape](Circle{}, Rect{})
func Variants[T any](vs ...T) {
	iface := reflect.TypeFor[T]()
	m := map[string]reflect.Type{}
	if old, ok := variants.Load(iface); ok {
		for k, v := range old.(map[string]reflect.Type) {
			m[k] = v
		}
	}
	for _, v := range vs {
		t := reflect.TypeOf(any(v))
		m[TypeName(t)] = t
	}
	variants.Store(iface, m)
}

func variantType(iface reflect.Type, tag string) (reflect.Type, error) {
	if m, ok := variants.Load(iface); ok {
		if t, ok := m.(map[string]reflect.Type)[tag]; ok {
			return t, nil
		}
	}
	return nil, fmt.Errorf("%s is no registered variant of %s, see gostrict.Variants", tag, iface)
}

// ToNeutral converts a value of Go into the neutral value of an engine: records into the []any of their fields,
// lists and outboxes into the []any of their elements, maps into the []any of their entries, values of sealed
// interfaces into a [Variant] and pointers to states into a [Token].
func ToNeutral(v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	return toNeutral(reflect.ValueOf(v))
}

// ToNeutralValue is like [ToNeutral], but keeps the static type of v, e.g. of a sealed interface or an error.
func ToNeutralValue(v reflect.Value) (any, error) {
	return toNeutral(v)
}

// FromNeutralType is like [FromNeutral] and returns the value of type t.
func FromNeutralType(n any, t reflect.Type) (reflect.Value, error) {
	return fromNeutral(n, t)
}

func toNeutral(v reflect.Value) (any, error) {
	t := v.Type()
	switch {
	case t == refType, t == tokenType:
		return v.Interface(), nil
	case t.Implements(typedStateType):
		return v.Interface().(typedState).token(), nil
	case t.Implements(hostSeqType):
		seq := v.Interface().(hostSeq)
		elems := seq.HostValues()
		for i, e := range elems {
			n, err := toNeutral(typed(e, seq.HostElem()))
			if err != nil {
				return nil, err
			}
			elems[i] = n
		}
		return elems, nil
	case t.Implements(hostMapType):
		m := v.Interface().(hostMap)
		entries := m.HostEntries()
		res := make([]any, len(entries))
		for i, e := range entries {
			k, err := toNeutral(typed(e[0], m.HostKey()))
			if err != nil {
				return nil, err
			}
			val, err := toNeutral(typed(e[1], m.HostElem()))
			if err != nil {
				return nil, err
			}
			res[i] = []any{k, val}
		}
		return res, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int32:
		return int32(v.Int()), nil
	case reflect.Uint32:
		return uint32(v.Uint()), nil
	case reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Struct:
		fields := make([]any, t.NumField())
		for i := range fields {
			f, err := toNeutral(v.Field(i))
			if err != nil {
				return nil, fmt.Errorf("field %s of %s: %w", t.Field(i).Name, t, err)
			}
			fields[i] = f
		}
		return fields, nil
	case reflect.Pointer:
		if t.Elem().Kind() == reflect.Struct {
			return NewToken(TypeName(t), v.Interface()), nil
		}
	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		if t == goErrorType {
			return v.Interface(), nil
		}
		dyn := v.Elem()
		fields, err := toNeutral(dyn)
		if err != nil {
			return nil, err
		}
		return Variant{Tag: TypeName(dyn.Type()), Fields: fields.([]any)}, nil
	}
	return nil, fmt.Errorf("%s can not cross the boundary", t)
}

// typed returns v as a value of t, which keeps an interface, e.g. of a sealed interface, as the static type.
func typed(v any, t reflect.Type) reflect.Value {
	res := reflect.New(t).Elem()
	if v != nil {
		res.Set(reflect.ValueOf(v))
	}
	return res
}

// FromNeutral converts the neutral value n into the value of Go, to which ptr points.
func FromNeutral(n any, ptr any) error {
	p := reflect.ValueOf(ptr)
	if p.Kind() != reflect.Pointer || p.IsNil() {
		return errors.New("FromNeutral needs a pointer")
	}
	v, err := fromNeutral(n, p.Type().Elem())
	if err != nil {
		return err
	}
	p.Elem().Set(v)
	return nil
}

func fromNeutral(n any, t reflect.Type) (reflect.Value, error) {
	zero := reflect.Zero(t)
	switch {
	case t == refType, t == tokenType:
		if n == nil {
			return zero, nil
		}
		return reflect.ValueOf(n), nil
	case t.Implements(typedStateType):
		tok, ok := n.(Token)
		if !ok {
			return zero, fmt.Errorf("%T is no state", n)
		}
		res := reflect.New(t).Elem()
		res.Field(0).Set(reflect.ValueOf(tok))
		return res, nil
	case t.Implements(hostSeqType):
		elems, _ := n.([]any)
		seq := zero.Interface().(hostSeq)
		values := make([]any, len(elems))
		for i, e := range elems {
			v, err := fromNeutral(e, seq.HostElem())
			if err != nil {
				return zero, err
			}
			values[i] = v.Interface()
		}
		return reflect.ValueOf(seq.HostFromValues(values)), nil
	case t.Implements(hostMapType):
		entries, _ := n.([]any)
		m := zero.Interface().(hostMap)
		res := make([][2]any, len(entries))
		for i, e := range entries {
			kv, ok := e.([]any)
			if !ok || len(kv) != 2 {
				return zero, fmt.Errorf("entry %d of %s is no pair", i, t)
			}
			k, err := fromNeutral(kv[0], m.HostKey())
			if err != nil {
				return zero, err
			}
			val, err := fromNeutral(kv[1], m.HostElem())
			if err != nil {
				return zero, err
			}
			res[i] = [2]any{k.Interface(), val.Interface()}
		}
		return reflect.ValueOf(m.HostFromEntries(res)), nil
	}

	switch t.Kind() {
	case reflect.Bool, reflect.Int32, reflect.Uint32, reflect.Float64, reflect.String:
		if n == nil {
			return zero, nil
		}
		v := reflect.ValueOf(n)
		if !v.CanConvert(t) || v.Kind() != t.Kind() {
			return zero, fmt.Errorf("%T is no %s", n, t)
		}
		return v.Convert(t), nil
	case reflect.Struct:
		fields, _ := n.([]any)
		if len(fields) != t.NumField() {
			return zero, fmt.Errorf("%s has %d fields, got %d values", t, t.NumField(), len(fields))
		}
		res := reflect.New(t).Elem()
		for i, f := range fields {
			v, err := fromNeutral(f, t.Field(i).Type)
			if err != nil {
				return zero, fmt.Errorf("field %s of %s: %w", t.Field(i).Name, t, err)
			}
			res.Field(i).Set(v)
		}
		return res, nil
	case reflect.Pointer:
		tok, ok := n.(Token)
		if !ok {
			return zero, fmt.Errorf("%T is no state of %s", n, t)
		}
		v := reflect.ValueOf(tok.Value())
		if !v.IsValid() || v.Type() != t {
			return zero, fmt.Errorf("the state of %s belongs to another engine", tok.Type())
		}
		return v, nil
	case reflect.Interface:
		if n == nil {
			return zero, nil
		}
		if t == goErrorType {
			err, ok := n.(error)
			if !ok {
				return zero, fmt.Errorf("%T is no error", n)
			}
			return reflect.ValueOf(&err).Elem(), nil
		}
		x, ok := n.(Variant)
		if !ok {
			return zero, fmt.Errorf("%T is no variant of %s", n, t)
		}
		vt, err := variantType(t, x.Tag)
		if err != nil {
			return zero, err
		}
		v, err := fromNeutral(x.Fields, vt)
		if err != nil {
			return zero, err
		}
		res := reflect.New(t).Elem()
		res.Set(v)
		return res, nil
	}
	return zero, fmt.Errorf("%s can not cross the boundary", t)
}

// Call calls an exported function in any engine with values of Go and stores its results in the pointers. An
// instance of the engines of this module knows its bundle, so a variant of a sealed interface may be passed as is.
// For example:
//
//	var x, y float64
//	err := gostrict.Call(ctx, in, "Position", []any{state, 1.5}, &x, &y)
//
// A state is passed as the [Token], which a former call stored, or as a [State].
func Call(ctx context.Context, in Instance, name string, args []any, results ...any) error {
	var params []Type
	if b, ok := in.(interface{ Bundle() *Bundle }); ok {
		params = b.Bundle().Exports[name].Params
	}

	neutral := make([]any, len(args))
	for i, a := range args {
		n, err := ToNeutral(a)
		if err != nil {
			return fmt.Errorf("value %d of %s: %w", i, name, err)
		}
		// a variant, which is passed as any, has lost its sealed interface
		if fields, ok := n.([]any); ok && i < len(params) && params[i].Kind == ValueUnion {
			n = Variant{Tag: TypeName(reflect.TypeOf(a)), Fields: fields}
		}
		neutral[i] = n
	}

	res, err := in.Call(ctx, name, neutral...)
	if err != nil {
		return err
	}
	if len(results) != len(res) {
		return fmt.Errorf("%s returns %d values, got %d pointers", name, len(res), len(results))
	}
	for i, r := range results {
		if err := FromNeutral(res[i], r); err != nil {
			return fmt.Errorf("result %d of %s: %w", i, name, err)
		}
	}
	return nil
}

func (s State[T]) token() Token {
	return s.Token
}
