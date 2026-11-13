// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build !gostrict

package json

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"go.wdy.de/nago/pkg/gostrict/std/big"
	"go.wdy.de/nago/pkg/gostrict/std/bin"
	"go.wdy.de/nago/pkg/gostrict/std/decimal"
	"go.wdy.de/nago/pkg/gostrict/std/json/core"
)

// decimalType is decimal.Dec, which JSON reads and writes as a decimal number, not as its fields.
var decimalType = reflect.TypeFor[decimal.Dec]()

// decode is Decode in native Go. The register machine and JavaScript have their own, see package core.
func decode[T any](data bin.Bytes) (T, error) {
	var v T
	d := core.New(data.HostBytes())
	value(d, reflect.ValueOf(&v).Elem())
	d.End()
	if d.Failed() {
		var zero T
		fill(reflect.ValueOf(&zero).Elem())
		return zero, errors.New(d.Err())
	}
	return v, nil
}

// The types of std know how to create their values from values of Go, see e.g. list.List.HostFromValues.
type (
	optionHost interface {
		HostFromOption(v any, ok bool) any
		HostElem() reflect.Type
	}
	listHost interface {
		HostFromValues(values []any) any
		HostElem() reflect.Type
	}
	mapHost interface {
		HostFromEntries(entries [][2]any) any
		HostKey() reflect.Type
		HostElem() reflect.Type
	}
	bytesHost interface {
		HostFromBytes(b []byte) any
	}
	bigHost interface {
		HostFromString(s string) any
	}
)

func hostOf[I any](t reflect.Type) (I, bool) {
	h, ok := reflect.Zero(t).Interface().(I)
	return h, ok
}

func value(d *core.Decoder, v reflect.Value) {
	if d.Null() {
		fill(v)
		return
	}
	t := v.Type()
	if h, ok := hostOf[optionHost](t); ok {
		e := reflect.New(h.HostElem()).Elem()
		value(d, e)
		v.Set(reflect.ValueOf(h.HostFromOption(e.Interface(), true)))
		return
	}
	if h, ok := hostOf[listHost](t); ok {
		values := []any{}
		if d.Array() {
			for i := 0; d.Next(i, ']'); i++ {
				e := reflect.New(h.HostElem()).Elem()
				value(d, e)
				values = append(values, e.Interface())
			}
		}
		v.Set(reflect.ValueOf(h.HostFromValues(values)))
		return
	}
	if h, ok := hostOf[mapHost](t); ok {
		var entries [][2]any
		if d.Object() {
			for i := 0; d.Next(i, '}'); i++ {
				k, ok := d.Key()
				if !ok {
					break
				}
				key := reflect.ValueOf(string(k)).Convert(h.HostKey()).Interface()
				e := reflect.New(h.HostElem()).Elem()
				value(d, e)
				entries = append(entries, [2]any{key, e.Interface()})
			}
		}
		v.Set(reflect.ValueOf(h.HostFromEntries(entries)))
		return
	}
	if h, ok := hostOf[bytesHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromBytes(d.Base64())))
		return
	}
	if h, ok := hostOf[bigHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromString(d.BigInt())))
		return
	}
	if t == decimalType {
		digits, scale := d.Decimal()
		v.Set(reflect.ValueOf(decimal.Dec{M: big.Int{}.HostFromString(digits).(big.Int), S: scale}))
		return
	}

	switch t.Kind() {
	case reflect.Bool:
		v.SetBool(d.Bool())
	case reflect.Int32:
		v.SetInt(int64(d.Int32()))
	case reflect.Uint32:
		v.SetUint(uint64(d.Uint32()))
	case reflect.Int64:
		v.SetInt(d.Int64())
	case reflect.Uint64:
		v.SetUint(d.Uint64())
	case reflect.Float64:
		v.SetFloat(d.Float64())
	case reflect.String:
		v.SetString(d.String())
	case reflect.Struct:
		record(d, v)
	default:
		// the checker allows no other type, see CodeJSON
		panic(fmt.Errorf("json: %s is not supported", t))
	}
}

func record(d *core.Decoder, v reflect.Value) {
	fill(v)
	if !d.Object() {
		return
	}
	byName := fieldsOf(v.Type()).byName
	for i := 0; d.Next(i, '}'); i++ {
		k, ok := d.Key()
		if !ok {
			return
		}
		if idx, ok := byName[string(k)]; ok {
			value(d, v.Field(idx))
		} else {
			d.Skip()
		}
	}
}

var fieldCache sync.Map // reflect.Type -> *structInfo

// structInfo has the exported fields of a struct, which have a key, in their order and by their keys. The first field
// of a key wins.
type structInfo struct {
	byName map[string]int
	fields []fieldInfo
}

type fieldInfo struct {
	key   string
	index int
	omit  bool
}

func fieldsOf(t reflect.Type) *structInfo {
	if m, ok := fieldCache.Load(t); ok {
		return m.(*structInfo)
	}
	info := &structInfo{byName: map[string]int{}}
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		key, ok := core.Key(f.Name, string(f.Tag))
		if !ok {
			continue
		}
		if _, dup := info.byName[key]; !dup {
			info.byName[key] = i
			info.fields = append(info.fields, fieldInfo{key: key, index: i, omit: core.OmitEmpty(string(f.Tag))})
		}
	}
	fieldCache.Store(t, info)
	return info
}

// fill sets the zero value, which a program sees: an empty list, map or bin.Bytes, 0 of a big.Int and None of an
// option, also in unexported fields.
func fill(v reflect.Value) {
	t := v.Type()
	if !v.CanSet() {
		v = reflect.NewAt(t, unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	if _, ok := hostOf[optionHost](t); ok {
		v.SetZero()
		return
	}
	if h, ok := hostOf[listHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromValues([]any{})))
		return
	}
	if h, ok := hostOf[mapHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromEntries(nil)))
		return
	}
	if h, ok := hostOf[bytesHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromBytes(nil)))
		return
	}
	if h, ok := hostOf[bigHost](t); ok {
		v.Set(reflect.ValueOf(h.HostFromString("0")))
		return
	}
	if t.Kind() == reflect.Struct {
		for i := range t.NumField() {
			fill(v.Field(i))
		}
		return
	}
	v.SetZero()
}

// encode is Encode in native Go, see package core.
func encode[T any](v T) (bin.Bytes, error) {
	var e core.Encoder
	write(&e, reflect.ValueOf(&v).Elem())
	b, err := e.Bytes()
	if err != "" {
		return bin.Empty(), errors.New(err)
	}
	return bin.OfSlice(b), nil
}

// The types of std know how to return their values for a host in Go, see e.g. list.List.HostValues.
type (
	optionValue interface{ HostOption() (any, bool) }
	listValues  interface{ HostValues() []any }
	mapEntries  interface{ HostEntries() [][2]any }
	bytesValue  interface{ HostBytes() []byte }
	bigValue    interface{ HostString() string }
)

func write(e *core.Encoder, v reflect.Value) {
	switch h := v.Interface().(type) {
	case optionValue:
		if x, ok := h.HostOption(); ok {
			write(e, reflect.ValueOf(x))
		} else {
			e.Null()
		}
		return
	case listValues:
		e.Byte('[')
		for i, x := range h.HostValues() {
			if i > 0 {
				e.Byte(',')
			}
			write(e, reflect.ValueOf(x))
		}
		e.Byte(']')
		return
	case mapEntries:
		e.Byte('{')
		for i, x := range h.HostEntries() {
			if i > 0 {
				e.Byte(',')
			}
			e.Key(reflect.ValueOf(x[0]).String())
			write(e, reflect.ValueOf(x[1]))
		}
		e.Byte('}')
		return
	case bytesValue:
		e.Base64(h.HostBytes())
		return
	case bigValue:
		e.Digits(h.HostString())
		return
	case decimal.Dec:
		e.Decimal(h.M.HostString(), h.S)
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		e.Bool(v.Bool())
	case reflect.Int32, reflect.Int64:
		e.Int(v.Int())
	case reflect.Uint32, reflect.Uint64:
		e.Uint(v.Uint())
	case reflect.Float64:
		e.Float(v.Float())
	case reflect.String:
		e.String(v.String())
	case reflect.Struct:
		e.Byte('{')
		first := true
		for _, f := range fieldsOf(v.Type()).fields {
			mark := e.Len()
			if !first {
				e.Byte(',')
			}
			e.Key(f.key)
			start := e.Len()
			write(e, v.Field(f.index))
			if f.omit && e.Empty(start) {
				e.Truncate(mark)
				continue
			}
			first = false
		}
		e.Byte('}')
	default:
		// the checker allows no other type, see CodeJSON
		panic(fmt.Errorf("json: %s is not supported", v.Type()))
	}
}
