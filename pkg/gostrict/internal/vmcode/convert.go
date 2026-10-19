// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"errors"
	"fmt"
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// The neutral values are those of gostrict.ValueKind: basic values, nil or an error, js.Ref, []any of lists, maps
// and records, the decimal string of a big.Int and the variants, tokens and callbacks of the host.

// neutral converts the value of type t at the slots.
func (m *Machine) neutral(t *tinfo, n []uint64, s []string, r []any) any {
	switch t.kind {
	case kBool:
		return n[0] != 0
	case kI32:
		return int32(n[0])
	case kU32:
		return uint32(n[0])
	case kF64:
		return math.Float64frombits(n[0])
	case kStr:
		return s[0]
	case kError:
		switch v := r[0].(type) {
		case nil:
			return nil
		case *progError:
			return m.host.ProgramError(v.msg, v.pos)
		case *object:
			// an error of the program crosses the boundary with its message
			return errors.New(m.errorMessage(v))
		}
		return r[0].(error)
	case kHandle:
		return js.NewRef(r[0])
	case kBig:
		return bigString(r[0])
	case kPtr:
		return m.host.Token(t.elem.name, r[0])
	case kFunc:
		return m.host.Callback(r[0].(*Closure))
	case kList, kOut:
		return m.neutralList(t.elem, r[0].(*list))
	case kMap:
		d := r[0].(*dict)
		res := make([]any, d.keys.len)
		kt, vt := t.key, t.elem
		for i := range int(d.keys.len) {
			k := m.neutral(kt, d.keys.n[i*int(kt.l.n):], d.keys.s[i*int(kt.l.s):], nil)
			v := m.neutral(vt, d.vals.n[i*int(vt.l.n):], d.vals.s[i*int(vt.l.s):], d.vals.r[i*int(vt.l.r):])
			res[i] = []any{k, v}
		}
		return res
	case kStruct:
		return m.fields(t, n, s, r)
	case kShared:
		o := r[0].(*object)
		if c := o.cache.Load(); c != nil {
			return *c
		}
		// the cache keeps the boxed value, so that returning it needs no allocation
		var res any = m.fields(t, o.n, o.s, o.r)
		o.cache.Store(&res)
		return res
	case kIface:
		o := r[0].(*object)
		if t.handles {
			return m.host.Variant(o.t.name, []any{js.NewRef(o.r[0])})
		}
		vt := o.t
		if vt.kind == kShared {
			inner := o.r[0].(*object)
			return m.host.Variant(vt.name, m.fields(vt, inner.n, inner.s, inner.r))
		}
		return m.host.Variant(vt.name, m.fields(vt, o.n, o.s, o.r))
	}
	panic(&vmError{err: fmt.Errorf("vm: the conversion of %s is not supported", t.name)})
}

func (m *Machine) neutralList(el *tinfo, l *list) []any {
	res := make([]any, l.len)
	for i := range int(l.len) {
		res[i] = m.neutral(el, l.n[i*int(el.l.n):], l.s[i*int(el.l.s):], l.r[i*int(el.l.r):])
	}
	return res
}

// fields returns the fields of a record in the order of the boundary.
func (m *Machine) fields(t *tinfo, n []uint64, s []string, r []any) []any {
	res := make([]any, len(t.rec))
	for i, fi := range t.rec {
		f := t.fields[fi]
		res[i] = m.neutral(f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
	}
	return res
}

// write writes a neutral value of type t into the slots.
func (m *Machine) write(v any, t *tinfo, n []uint64, s []string, r []any) {
	switch t.kind {
	case kBool:
		n[0] = fromBool(v.(bool))
	case kI32:
		n[0] = fromI32(v.(int32))
	case kU32:
		n[0] = uint64(v.(uint32))
	case kF64:
		n[0] = math.Float64bits(v.(float64))
	case kStr:
		s[0] = v.(string)
	case kError:
		if v == nil {
			r[0] = nil
			return
		}
		r[0] = v.(error)
	case kHandle:
		r[0] = v.(js.Ref).Value()
	case kBig:
		r[0] = bigParse(v.(string))
	case kPtr:
		o, ok := m.host.TokenValue(v)
		if _, isObj := o.(*object); !ok || !isObj {
			panic(&vmError{err: fmt.Errorf("the state of %s belongs to another engine", t.elem.name)})
		}
		r[0] = o
	case kList, kOut:
		r[0] = m.newList(t.elem, v.([]any))
	case kMap:
		entries := v.([]any)
		d := newDict(t.key)
		kv := make([]uint64, t.key.l.n)
		ks := make([]string, t.key.l.s)
		vn, vs, vr := make([]uint64, t.elem.l.n), make([]string, t.elem.l.s), make([]any, t.elem.l.r)
		for _, e := range entries {
			pair := e.([]any)
			m.write(pair[0], t.key, kv, ks, nil)
			m.write(pair[1], t.elem, vn, vs, vr)
			d.set(t.key, t.elem.l, kv, ks, vn, vs, vr)
		}
		r[0] = d
	case kStruct:
		m.writeFields(v.([]any), t, n, s, r)
	case kShared:
		o := newObject(t, t.obj)
		m.writeFields(v.([]any), t, o.n, o.s, o.r)
		r[0] = o
	case kIface:
		tag, fields := m.host.VariantOf(v)
		vt := m.prog.byName[tag]
		if vt == nil {
			panic(&vmError{err: fmt.Errorf("%s is no variant of %s", tag, t.name)})
		}
		if t.handles {
			r[0] = &object{t: vt, r: []any{fields[0].(js.Ref).Value()}}
			return
		}
		o := newObject(vt, vt.l)
		m.write(fields, vt, o.n, o.s, o.r)
		r[0] = o
	default:
		panic(&vmError{err: fmt.Errorf("vm: the conversion of %s is not supported", t.name)})
	}
}

func (m *Machine) newList(el *tinfo, vs []any) *list {
	l := &list{len: int32(len(vs))}
	if el.l.n > 0 {
		l.n = make([]uint64, len(vs)*int(el.l.n))
	}
	if el.l.s > 0 {
		l.s = make([]string, len(vs)*int(el.l.s))
	}
	if el.l.r > 0 {
		l.r = make([]any, len(vs)*int(el.l.r))
	}
	for i, e := range vs {
		m.write(e, el, l.n[i*int(el.l.n):], l.s[i*int(el.l.s):], l.r[i*int(el.l.r):])
	}
	return l
}

func (m *Machine) writeFields(fs []any, t *tinfo, n []uint64, s []string, r []any) {
	for i, fi := range t.rec {
		f := t.fields[fi]
		m.write(fs[i], f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
	}
}

// CheckArgs reports, whether the neutral values fit the parameters of the function, e.g. of a callback, which the
// host calls with values of its own.
func (c *Closure) CheckArgs(args []any) error {
	f := c.fn
	if len(args) != len(f.params.elems) {
		return fmt.Errorf("%s needs %d values, got %d", f.name, len(f.params.elems), len(args))
	}
	for i, t := range f.params.elems {
		ok := true
		switch t.kind {
		case kBool:
			_, ok = args[i].(bool)
		case kI32:
			_, ok = args[i].(int32)
		case kU32:
			_, ok = args[i].(uint32)
		case kF64:
			_, ok = args[i].(float64)
		case kStr, kBig:
			_, ok = args[i].(string)
		case kList, kOut, kMap, kStruct, kShared:
			_, ok = args[i].([]any)
		case kHandle:
			_, ok = args[i].(js.Ref)
		}
		if !ok {
			return fmt.Errorf("value %d of %s: %T does not fit", i, f.name, args[i])
		}
	}
	return nil
}
