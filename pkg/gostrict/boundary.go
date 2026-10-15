// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/types"
)

// valueKind returns the kind of a basic value, an error or a handle, or ValueInvalid.
func valueKind(t types.Type) ValueKind {
	if isError(t) {
		return ValueError
	}
	if isHandle(t) {
		return ValueHandle
	}
	switch basicKind(t) {
	case types.Bool:
		return ValueBool
	case types.Int32:
		return ValueInt32
	case types.Uint32:
		return ValueUint32
	case types.Float64:
		return ValueFloat64
	case types.String:
		return ValueString
	}
	return ValueInvalid
}

// signature returns the types of a function of a header package or false, if a parameter or result is not a basic
// value or a handle, or an error as a result. A generic function must be instantiated first.
func signature(name string, sig *types.Signature) (Func, bool) {
	f := Func{Name: name}
	if sig.Variadic() || sig.TypeParams().Len() > 0 || sig.Recv() != nil {
		return f, false
	}
	for v := range sig.Params().Variables() {
		k := valueKind(v.Type())
		if k == ValueInvalid || k == ValueError {
			return f, false
		}
		f.Params = append(f.Params, Type{Kind: k})
	}
	for v := range sig.Results().Variables() {
		k := valueKind(v.Type())
		if k == ValueInvalid {
			return f, false
		}
		f.Results = append(f.Results, Type{Kind: k})
	}
	return f, true
}

// boundary maps the types of exported functions to the types of the boundary and collects their records.
type boundary struct {
	records map[string]Record
	unions  map[string]Union
	// structs are the types of the records and states by their full name
	structs map[string]types.Type
}

func newBoundary() *boundary {
	return &boundary{records: map[string]Record{}, unions: map[string]Union{}, structs: map[string]types.Type{}}
}

// exportSignature returns the types of an exported function of the main package or false, if a parameter or
// result can not cross the boundary. Exports take records, lists and states but no handles, which only a header
// package creates.
func (b *boundary) exportSignature(name string, sig *types.Signature) (Func, bool) {
	f := Func{Name: name}
	if sig.Variadic() || sig.TypeParams().Len() > 0 || sig.Recv() != nil {
		return f, false
	}
	for v := range sig.Params().Variables() {
		t, ok := b.typeOf(v.Type())
		if !ok || t.Kind == ValueError {
			return f, false
		}
		f.Params = append(f.Params, t)
	}
	for v := range sig.Results().Variables() {
		t, ok := b.typeOf(v.Type())
		if !ok {
			return f, false
		}
		f.Results = append(f.Results, t)
	}
	return f, true
}

func (b *boundary) typeOf(t types.Type) (Type, bool) {
	if k := valueKind(t); k != ValueInvalid {
		return Type{Kind: k}, k != ValueHandle
	}

	if isOpaque(t) {
		n := types.Unalias(t).(*types.Named)
		if n.Obj().Name() == "Out" {
			elem, ok := b.typeOf(n.TypeArgs().At(0))
			if !ok || (elem.Kind != ValueRecord && elem.Kind != ValueUnion) {
				return Type{}, false
			}
			return Type{Kind: ValueOut, Elem: &elem}, true
		}
		if n.Obj().Name() == "Map" {
			key, ok := b.typeOf(n.TypeArgs().At(0))
			if !ok || key.Kind < ValueBool || key.Kind > ValueString || key.Kind == ValueFloat64 {
				return Type{}, false
			}
			elem, ok := b.typeOf(n.TypeArgs().At(1))
			if !ok || elem.Kind == ValueError || elem.Kind == ValueState || elem.Kind == ValueOut {
				return Type{}, false
			}
			return Type{Kind: ValueMap, Key: &key, Elem: &elem}, true
		}
		if n.Obj().Name() != "List" {
			return Type{}, false
		}
		elem, ok := b.typeOf(n.TypeArgs().At(0))
		if !ok || elem.Kind == ValueError || elem.Kind == ValueState {
			return Type{}, false
		}
		return Type{Kind: ValueList, Elem: &elem}, true
	}

	if p, ok := t.Underlying().(*types.Pointer); ok {
		elem := types.Unalias(p.Elem())
		if !isStruct(elem) {
			return Type{}, false
		}
		name := typeKey(elem)
		b.structs[name] = elem
		return Type{Kind: ValueState, Name: name}, true
	}

	if variants, ok := sealedVariants(t); ok {
		name := typeKey(t)
		if _, ok := b.unions[name]; ok {
			return Type{Kind: ValueUnion, Name: name}, true
		}
		u := Union{Name: name}
		b.unions[name] = u
		for _, v := range variants {
			vt, ok := b.typeOf(v)
			if !ok || vt.Kind != ValueRecord {
				delete(b.unions, name)
				return Type{}, false
			}
			u.Variants = append(u.Variants, vt.Name)
		}
		if len(u.Variants) == 0 {
			delete(b.unions, name)
			return Type{}, false
		}
		b.unions[name] = u
		b.structs[name] = t
		return Type{Kind: ValueUnion, Name: name}, true
	}

	if !isStruct(t) {
		return Type{}, false
	}
	name := typeKey(t)
	if _, ok := b.records[name]; ok {
		return Type{Kind: ValueRecord, Name: name}, true
	}

	st := t.Underlying().(*types.Struct)
	// register the record before its fields, so that a record may contain a list of itself
	b.records[name] = Record{Name: name}
	rec := Record{Name: name}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		ft, ok := b.typeOf(f.Type())
		if !f.Exported() || !ok || ft.Kind == ValueError || ft.Kind == ValueState {
			delete(b.records, name)
			return Type{}, false
		}
		rec.Fields = append(rec.Fields, Field{Name: f.Name(), JS: fieldName(f, i), Type: ft})
	}
	b.records[name] = rec
	b.structs[name] = t
	return Type{Kind: ValueRecord, Name: name}, true
}
