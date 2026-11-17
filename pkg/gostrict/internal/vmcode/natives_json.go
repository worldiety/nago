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
	"go/token"
	"go/types"
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/json/core"
)

const (
	jsonPath    = stdPrefix + "json"
	decimalPath = stdPrefix + "decimal"
)

func init() {
	genericNatives[jsonPath+".Decode"] = jsonDecode
	genericNatives[jsonPath+".Encode"] = jsonEncode
}

// jtype is a type, which json.Decode reads, with the keys of the fields of a struct.
type jtype struct {
	t      *tinfo
	elem   *jtype
	fields []jfield
	// byName are the indices of the exported fields by their keys
	byName map[string]int
	// keys are the exported fields with a key in their order, which json.Encode writes
	keys []jkey
	// dec is set for decimal.Dec, which JSON reads and writes as a decimal number, whose integer is the field at m and
	// whose scale is the field at s
	dec  bool
	m, s layout
}

// isDecimal tells, whether t is decimal.Dec.
func isDecimal(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	return ok && n.Obj().Pkg() != nil && n.Obj().Pkg().Path() == decimalPath && n.Obj().Name() == "Dec"
}

type jkey struct {
	name  string
	index int
	omit  bool
}

type jfield struct {
	t   *jtype
	off layout
}

// jsonType describes t for json.Decode. A recursive type refers to itself through a list, a map or an option.
func jsonType(t *tinfo, memo map[*tinfo]*jtype) *jtype {
	if jt, ok := memo[t]; ok {
		return jt
	}
	jt := &jtype{t: t}
	memo[t] = jt
	switch t.kind {
	case kList, kOpt, kMap:
		jt.elem = jsonType(t.elem, memo)
	case kStruct, kShared:
		st := t.t.Underlying().(*types.Struct)
		jt.byName = map[string]int{}
		jt.dec = isDecimal(t.t)
		for i, f := range t.fields {
			jt.fields = append(jt.fields, jfield{t: jsonType(f.t, memo), off: f.off})
			sf := st.Field(i)
			switch {
			case jt.dec && sf.Name() == "M":
				jt.m = f.off
			case jt.dec && sf.Name() == "S":
				jt.s = f.off
			}
			if !sf.Exported() {
				continue
			}
			if key, ok := core.Key(sf.Name(), st.Tag(i)); ok {
				if _, dup := jt.byName[key]; !dup {
					jt.byName[key] = i
					jt.keys = append(jt.keys, jkey{name: key, index: i, omit: core.OmitEmpty(st.Tag(i))})
				}
			}
		}
	}
	return jt
}

// jsonDecode is json.Decode, which writes the value directly into the slots of the result without a value in between.
func jsonDecode(c *compiler, targs []*tinfo, pos token.Pos) *native {
	jt := jsonType(targs[0], map[*tinfo]*jtype{})
	l := targs[0].l
	return &native{name: "json.Decode", fn: func(m *Machine, n []uint64, s []string, r []any) {
		// the decoder keeps the data, before the result overwrites the slot of the argument
		d := core.New(bytesOf(r[0]))
		jsonValue(d, jt, n, s, r)
		d.End()
		if d.Failed() {
			jsonEmpty(jt, n, s, r)
			r[l.r] = errors.New(d.Err())
			return
		}
		r[l.r] = nil
	}}
}

func jsonValue(d *core.Decoder, jt *jtype, n []uint64, s []string, r []any) {
	if d.Null() {
		jsonEmpty(jt, n, s, r)
		return
	}
	t := jt.t
	switch t.kind {
	case kBool:
		n[0] = fromBool(d.Bool())
	case kI32:
		n[0] = fromI32(d.Int32())
	case kU32:
		n[0] = uint64(d.Uint32())
	case kI64:
		n[0] = uint64(d.Int64())
	case kU64:
		n[0] = d.Uint64()
	case kF64:
		n[0] = math.Float64bits(d.Float64())
	case kStr:
		s[0] = d.String()
	case kBytes:
		r[0] = d.Base64()
	case kBig:
		r[0] = bigParse(d.BigInt())
	case kOpt:
		el := jt.elem
		if optDirect(el.t) {
			jsonValue(d, el, nil, nil, r)
			return
		}
		o := newObject(el.t, el.t.l)
		jsonValue(d, el, o.n, o.s, o.r)
		r[0] = o
	case kList:
		el := jt.elem
		res := &list{}
		if d.Array() {
			en, es, er := make([]uint64, el.t.l.n), make([]string, el.t.l.s), make([]any, el.t.l.r)
			for i := 0; d.Next(i, ']'); i++ {
				jsonValue(d, el, en, es, er)
				res.add(el.t.l, en, es, er)
			}
		}
		r[0] = res
	case kMap:
		el := jt.elem
		res := newDict(t.key)
		if d.Object() {
			ks := make([]string, 1)
			vn, vs, vr := make([]uint64, el.t.l.n), make([]string, el.t.l.s), make([]any, el.t.l.r)
			for i := 0; d.Next(i, '}'); i++ {
				k, ok := d.Key()
				if !ok {
					break
				}
				ks[0] = string(k)
				jsonValue(d, el, vn, vs, vr)
				res.set(t.key, el.t.l, nil, ks, vn, vs, vr)
			}
		}
		r[0] = res
	case kStruct:
		jsonRecord(d, jt, n, s, r)
	case kShared:
		o := newObject(t, t.obj)
		jsonRecord(d, jt, o.n, o.s, o.r)
		r[0] = o
	default:
		// the checker allows no other type
		panic(&vmError{err: fmt.Errorf("vm: json.Decode of %s is not supported", t.name)})
	}
}

func jsonRecord(d *core.Decoder, jt *jtype, n []uint64, s []string, r []any) {
	if jt.dec {
		digits, scale := d.Decimal()
		r[jt.m.r] = bigParse(digits)
		n[jt.s.n] = fromI32(scale)
		return
	}
	jsonEmptyFields(jt, n, s, r)
	if !d.Object() {
		return
	}
	for i := 0; d.Next(i, '}'); i++ {
		k, ok := d.Key()
		if !ok {
			return
		}
		idx, ok := jt.byName[string(k)]
		if !ok {
			d.Skip()
			continue
		}
		f := jt.fields[idx]
		jsonValue(d, f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
	}
}

// jsonEmpty writes the zero value, which a program sees: an empty list, map or bin.Bytes, 0 of a big.Int and None of
// an option.
func jsonEmpty(jt *jtype, n []uint64, s []string, r []any) {
	t := jt.t
	switch t.kind {
	case kBool, kI32, kU32, kI64, kU64, kF64:
		n[0] = 0
	case kStr:
		s[0] = ""
	case kBytes:
		r[0] = []byte(nil)
	case kBig:
		r[0] = bigZero
	case kOpt:
		r[0] = nil
	case kList:
		r[0] = &list{}
	case kMap:
		r[0] = newDict(t.key)
	case kStruct:
		jsonEmptyFields(jt, n, s, r)
	case kShared:
		o := newObject(t, t.obj)
		jsonEmptyFields(jt, o.n, o.s, o.r)
		r[0] = o
	}
}

func jsonEmptyFields(jt *jtype, n []uint64, s []string, r []any) {
	for _, f := range jt.fields {
		jsonEmpty(f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
	}
}

// jsonEncode is json.Encode, which reads the value directly from the slots of the argument.
func jsonEncode(c *compiler, targs []*tinfo, pos token.Pos) *native {
	jt := jsonType(targs[0], map[*tinfo]*jtype{})
	return &native{name: "json.Encode", fn: func(m *Machine, n []uint64, s []string, r []any) {
		var e core.Encoder
		jsonWrite(&e, jt, n, s, r)
		// the results overwrite the slots of the argument only now
		b, err := e.Bytes()
		if err != "" {
			r[0], r[1] = []byte(nil), errors.New(err)
			return
		}
		r[0], r[1] = b, nil
	}}
}

func jsonWrite(e *core.Encoder, jt *jtype, n []uint64, s []string, r []any) {
	t := jt.t
	switch t.kind {
	case kBool:
		e.Bool(n[0] != 0)
	case kI32:
		e.Int(int64(int32(n[0])))
	case kU32:
		e.Uint(uint64(uint32(n[0])))
	case kI64:
		e.Int(int64(n[0]))
	case kU64:
		e.Uint(n[0])
	case kF64:
		e.Float(math.Float64frombits(n[0]))
	case kStr:
		e.String(s[0])
	case kBytes:
		e.Base64(bytesOf(r[0]))
	case kBig:
		e.Digits(bigString(r[0]))
	case kOpt:
		el := jt.elem
		switch {
		case r[0] == nil:
			e.Null()
		case optDirect(el.t):
			jsonWrite(e, el, nil, nil, r)
		default:
			o := r[0].(*object)
			jsonWrite(e, el, o.n, o.s, o.r)
		}
	case kList:
		el, l := jt.elem, r[0].(*list)
		e.Byte('[')
		for i := range int(l.len) {
			if i > 0 {
				e.Byte(',')
			}
			jsonWrite(e, el, l.n[i*int(el.t.l.n):], l.s[i*int(el.t.l.s):], l.r[i*int(el.t.l.r):])
		}
		e.Byte(']')
	case kMap:
		el, d := jt.elem, r[0].(*dict)
		ks := int(t.key.l.s)
		e.Byte('{')
		for i := range int(d.keys.len) {
			if i > 0 {
				e.Byte(',')
			}
			e.Key(d.keys.s[i*ks])
			jsonWrite(e, el, d.vals.n[i*int(el.t.l.n):], d.vals.s[i*int(el.t.l.s):], d.vals.r[i*int(el.t.l.r):])
		}
		e.Byte('}')
	case kStruct:
		jsonWriteFields(e, jt, n, s, r)
	case kShared:
		o := r[0].(*object)
		jsonWriteFields(e, jt, o.n, o.s, o.r)
	default:
		// the checker allows no other type
		panic(&vmError{err: fmt.Errorf("vm: json.Encode of %s is not supported", t.name)})
	}
}

func jsonWriteFields(e *core.Encoder, jt *jtype, n []uint64, s []string, r []any) {
	if jt.dec {
		e.Decimal(bigString(r[jt.m.r]), int32(n[jt.s.n]))
		return
	}
	e.Byte('{')
	first := true
	for _, k := range jt.keys {
		mark := e.Len()
		if !first {
			e.Byte(',')
		}
		e.Key(k.name)
		start := e.Len()
		f := jt.fields[k.index]
		jsonWrite(e, f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
		if k.omit && e.Empty(start) {
			e.Truncate(mark)
			continue
		}
		first = false
	}
	e.Byte('}')
}
