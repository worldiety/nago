// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import "go/token"

// dict is a dict.Map: the keys and values in the order of their insertion, like std/dict, and the index of each
// key. A key is a single scalar or string.
type dict struct {
	keys, vals list
	idxN       map[uint64]int32
	idxS       map[string]int32
}

func newDict(kt *tinfo) *dict {
	if kt.l.s > 0 {
		return &dict{idxS: map[string]int32{}}
	}
	return &dict{idxN: map[uint64]int32{}}
}

// norm returns the scalar of a key in the width of its type, e.g. a uint32 without the sign of an addition.
func norm(kt *tinfo, v uint64) uint64 {
	switch kt.kind {
	case kI32:
		return fromI32(int32(v))
	case kU32:
		return uint64(uint32(v))
	}
	return v
}

func (d *dict) find(kt *tinfo, kn []uint64, ks []string) (int32, bool) {
	if d.idxS != nil {
		i, ok := d.idxS[ks[0]]
		return i, ok
	}
	i, ok := d.idxN[norm(kt, kn[0])]
	return i, ok
}

func (d *dict) set(kt *tinfo, vl layout, kn []uint64, ks []string, vn []uint64, vs []string, vr []any) {
	if i, ok := d.find(kt, kn, ks); ok {
		d.vals.set(int(i), vl, vn, vs, vr)
		return
	}
	i := d.keys.len
	if d.idxS != nil {
		d.idxS[ks[0]] = i
		d.keys.add(kt.l, nil, ks, nil)
	} else {
		k := norm(kt, kn[0])
		d.idxN[k] = i
		d.keys.add(kt.l, []uint64{k}, nil, nil)
	}
	d.vals.add(vl, vn, vs, vr)
}

func (d *dict) delete(kt *tinfo, vl layout, kn []uint64, ks []string) {
	i, ok := d.find(kt, kn, ks)
	if !ok {
		return
	}
	if d.idxS != nil {
		delete(d.idxS, ks[0])
	} else {
		delete(d.idxN, norm(kt, kn[0]))
	}
	d.keys.delete(int(i), kt.l)
	d.vals.delete(int(i), vl)
	for j := i; j < d.keys.len; j++ {
		if d.idxS != nil {
			d.idxS[d.keys.s[j]] = j
		} else {
			d.idxN[d.keys.n[j]] = j
		}
	}
}

func dictNatives(c *compiler, name string, targs []*tinfo, pos token.Pos) *native {
	kt, vt := targs[0], targs[1]
	vl := vt.l
	zero := c.zeroSlots(vt, pos)
	switch name {
	case "New":
		return &native{name: "dict.New", fn: func(m *Machine, n []uint64, s []string, r []any) { r[0] = newDict(kt) }}
	case "Len":
		return &native{name: "Map.Len", fn: func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromI32(r[0].(*dict).keys.len) }}
	case "Get":
		return &native{name: "Map.Get", fn: func(m *Machine, n []uint64, s []string, r []any) {
			d := r[0].(*dict)
			i, ok := d.find(kt, n, s)
			if !ok {
				zero.copyTo(n, s, r)
				n[vl.n] = 0
				return
			}
			d.vals.get(int(i), vl, n, s, r)
			n[vl.n] = 1
		}}
	case "Has":
		return &native{name: "Map.Has", fn: func(m *Machine, n []uint64, s []string, r []any) {
			_, ok := r[0].(*dict).find(kt, n, s)
			n[0] = fromBool(ok)
		}}
	case "Set":
		return &native{name: "Map.Set", fn: func(m *Machine, n []uint64, s []string, r []any) {
			r[0].(*dict).set(kt, vl, n, s, n[kt.l.n:], s[kt.l.s:], r[1:])
		}}
	case "Delete":
		return &native{name: "Map.Delete", fn: func(m *Machine, n []uint64, s []string, r []any) {
			r[0].(*dict).delete(kt, vl, n, s)
		}}
	case "$keys":
		// the keys, which exist, when an iteration starts
		return &native{name: "Map.$keys", fn: func(m *Machine, n []uint64, s []string, r []any) {
			d := r[0].(*dict)
			r[0] = &list{n: cloneSlice(d.keys.n), s: cloneSlice(d.keys.s), len: d.keys.len}
		}}
	}
	return nil
}

// optDirect tells, whether an option of t holds the value itself, which is never nil, instead of a box.
func optDirect(t *tinfo) bool {
	return t.l == layout{r: 1} && t.kind != kError
}

func optNatives(c *compiler, name string, targs []*tinfo, pos token.Pos) *native {
	t := targs[0]
	l := t.l
	direct := optDirect(t)
	zero := c.zeroSlots(t, pos)
	switch name {
	case "Some":
		return &native{name: "opt.Some", fn: func(m *Machine, n []uint64, s []string, r []any) {
			if direct {
				return
			}
			o := newObject(t, l)
			copy(o.n, n)
			copy(o.s, s)
			copy(o.r, r)
			r[0] = o
		}}
	case "Get":
		return &native{name: "Option.Get", fn: func(m *Machine, n []uint64, s []string, r []any) {
			v := r[0]
			switch {
			case v == nil:
				zero.copyTo(n, s, r)
				n[l.n] = 0
			case direct:
				n[0] = 1
			default:
				o := v.(*object)
				copy(n, o.n)
				copy(s, o.s)
				copy(r, o.r)
				n[l.n] = 1
			}
		}}
	case "Or":
		// the receiver is r[0], the default follows it
		return &native{name: "Option.Or", fn: func(m *Machine, n []uint64, s []string, r []any) {
			v := r[0]
			switch {
			case v == nil:
				copy(r, r[1:1+l.r])
			case direct:
			default:
				o := v.(*object)
				copy(n, o.n)
				copy(s, o.s)
				copy(r, o.r)
			}
		}}
	case "IsSome":
		return &native{name: "Option.IsSome", fn: func(m *Machine, n []uint64, s []string, r []any) { n[0] = fromBool(r[0] != nil) }}
	}
	return nil
}
