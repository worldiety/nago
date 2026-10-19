// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"context"
	"errors"
	"fmt"
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/js"
)

// The fast path of a callback writes the arguments into the slots of the frame and reads the results from them,
// one value after another, without neutral values, see gostrict.FastCallback. A reader and a writer walk the values
// like the neutral conversion: tuples, records in the order of the boundary, lists, maps and variants.

const (
	lvTuple uint8 = iota
	lvRecord
	lvList
	lvMap
	lvSingle
)

// level is a value, whose parts a reader or writer walks.
type level struct {
	kind uint8
	// t is the tuple, record or single value, or the element of a list
	t       *tinfo
	n       []uint64
	s       []string
	r       []any
	l       *list
	d       *dict
	key, el *tinfo
	i       int
	// obj is the object of a shared record, whose Go value the reader keeps
	obj *object
	// key slots of a map, which the writer fills before the value
	kn []uint64
	ks []string
}

// next returns the type and slots of the next part.
func (lv *level) next() (*tinfo, []uint64, []string, []any) {
	i := lv.i
	lv.i++
	switch lv.kind {
	case lvTuple:
		t, o := lv.t.elems[i], lv.t.offs[i]
		return t, lv.n[o.n:], lv.s[o.s:], lv.r[o.r:]
	case lvRecord:
		f := &lv.t.recFields[i]
		o := f.off
		return f.t, lv.n[o.n:], lv.s[o.s:], lv.r[o.r:]
	case lvList:
		el := lv.t.l
		return lv.t, lv.l.n[i*int(el.n):], lv.l.s[i*int(el.s):], lv.l.r[i*int(el.r):]
	case lvMap:
		j := i / 2
		if i%2 == 0 {
			kl := lv.key.l
			return lv.key, lv.d.keys.n[j*int(kl.n):], lv.d.keys.s[j*int(kl.s):], nil
		}
		vl := lv.el.l
		return lv.el, lv.d.vals.n[j*int(vl.n):], lv.d.vals.s[j*int(vl.s):], lv.d.vals.r[j*int(vl.r):]
	}
	return lv.t, lv.n, lv.s, lv.r
}

// Reader reads the results of a fast call.
type Reader struct {
	m     *Machine
	stack []level
	// top is the last level of the stack
	top *level
	// last is the shared record, which has just ended
	last *object
}

func (rd *Reader) push(lv level) {
	rd.stack = append(rd.stack, lv)
	rd.top = &rd.stack[len(rd.stack)-1]
}

// field returns the offset of the next field of a record, or false for another level.
func (rd *Reader) field() (layout, bool) {
	lv := rd.top
	if lv.kind != lvRecord {
		return layout{}, false
	}
	f := &lv.t.recFields[lv.i]
	lv.i++
	return f.off, true
}

func (rd *Reader) reset(t *tinfo, n []uint64, s []string, r []any) {
	rd.stack = rd.stack[:0]
	kind := lvSingle
	if t.kind == kTuple {
		kind = lvTuple
	}
	rd.push(level{kind: kind, t: t, n: n, s: s, r: r})
	rd.last = nil
}

func (rd *Reader) next() (*tinfo, []uint64, []string, []any) {
	return rd.top.next()
}

func (rd *Reader) Bool() bool {
	if o, ok := rd.field(); ok {
		return rd.top.n[o.n] != 0
	}
	_, n, _, _ := rd.next()
	return n[0] != 0
}

func (rd *Reader) Int32() int32 {
	if o, ok := rd.field(); ok {
		return int32(rd.top.n[o.n])
	}
	_, n, _, _ := rd.next()
	return int32(n[0])
}

func (rd *Reader) Uint32() uint32 {
	if o, ok := rd.field(); ok {
		return uint32(rd.top.n[o.n])
	}
	_, n, _, _ := rd.next()
	return uint32(n[0])
}

func (rd *Reader) Float64() float64 {
	if o, ok := rd.field(); ok {
		return math.Float64frombits(rd.top.n[o.n])
	}
	_, n, _, _ := rd.next()
	return math.Float64frombits(n[0])
}

func (rd *Reader) String() string {
	if o, ok := rd.field(); ok {
		return rd.top.s[o.s]
	}
	_, _, s, _ := rd.next()
	return s[0]
}

func (rd *Reader) Error() error {
	_, _, _, r := rd.next()
	switch v := r[0].(type) {
	case nil:
		return nil
	case *object:
		return errors.New(rd.m.errorMessage(v))
	}
	return r[0].(error)
}

func (rd *Reader) Handle() js.Ref {
	_, _, _, r := rd.next()
	return js.NewRef(r[0])
}

func (rd *Reader) Big() string {
	_, _, _, r := rd.next()
	return bigString(r[0])
}

// StateValue returns the token of a state, see [Host.Token].
func (rd *Reader) StateValue() any {
	t, _, _, r := rd.next()
	return rd.m.host.Token(t.elem.name, r[0])
}

// FuncValue returns the callback of a function, see [Host.Callback].
func (rd *Reader) FuncValue() any {
	_, _, _, r := rd.next()
	return rd.m.host.Callback(r[0].(*Closure))
}

func (rd *Reader) Record() (any, bool) {
	t, n, s, r := rd.next()
	if t.kind == kShared {
		o := r[0].(*object)
		if k := o.kept.Load(); k != nil {
			return *k, true
		}
		rd.push(level{kind: lvRecord, t: t, n: o.n, s: o.s, r: o.r, obj: o})
		return nil, false
	}
	rd.push(level{kind: lvRecord, t: t, n: n, s: s, r: r})
	return nil, false
}

func (rd *Reader) List() int {
	t, _, _, r := rd.next()
	l := r[0].(*list)
	rd.push(level{kind: lvList, t: t.elem, l: l})
	return int(l.len)
}

func (rd *Reader) Map() int {
	t, _, _, r := rd.next()
	d := r[0].(*dict)
	rd.push(level{kind: lvMap, key: t.key, el: t.elem, d: d})
	return int(d.keys.len)
}

func (rd *Reader) Variant() string {
	t, _, _, r := rd.next()
	o := r[0].(*object)
	vt := o.t
	switch {
	case t.handles || vt.kind == kHandle:
		rd.push(level{kind: lvSingle, t: vt, n: o.n, s: o.s, r: o.r})
	case vt.kind == kShared:
		inner := o.r[0].(*object)
		rd.push(level{kind: lvRecord, t: vt, n: inner.n, s: inner.s, r: inner.r})
	default:
		rd.push(level{kind: lvRecord, t: vt, n: o.n, s: o.s, r: o.r})
	}
	return vt.name
}

func (rd *Reader) End() bool {
	obj := rd.top.obj
	rd.stack = rd.stack[:len(rd.stack)-1]
	rd.top = &rd.stack[len(rd.stack)-1]
	rd.last = obj
	return obj != nil
}

func (rd *Reader) Keep(v any) {
	if rd.last != nil {
		rd.last.kept.Store(&v)
	}
}

// Writer writes the arguments of a fast call.
type Writer struct {
	m     *Machine
	stack []level
}

func (w *Writer) reset(t *tinfo, n []uint64, s []string, r []any) {
	w.stack = append(w.stack[:0], level{kind: lvTuple, t: t, n: n, s: s, r: r})
}

func (w *Writer) next() (*tinfo, []uint64, []string, []any) {
	lv := &w.stack[len(w.stack)-1]
	if lv.kind == lvMap {
		// a key and then its value, which sets the entry
		if lv.i%2 == 0 {
			lv.i++
			return lv.key, lv.kn, lv.ks, nil
		}
		lv.i++
		vn, vs, vr := make([]uint64, lv.el.l.n), make([]string, lv.el.l.s), make([]any, lv.el.l.r)
		lv.d.set(lv.key, lv.el.l, lv.kn, lv.ks, vn, vs, vr)
		i, _ := lv.d.find(lv.key, lv.kn, lv.ks)
		vl := lv.el.l
		return lv.el, lv.d.vals.n[int(i)*int(vl.n):], lv.d.vals.s[int(i)*int(vl.s):], lv.d.vals.r[int(i)*int(vl.r):]
	}
	return lv.next()
}

func (w *Writer) Bool(v bool) {
	_, n, _, _ := w.next()
	n[0] = fromBool(v)
}

func (w *Writer) Int32(v int32) {
	_, n, _, _ := w.next()
	n[0] = fromI32(v)
}

func (w *Writer) Uint32(v uint32) {
	_, n, _, _ := w.next()
	n[0] = uint64(v)
}

func (w *Writer) Float64(v float64) {
	_, n, _, _ := w.next()
	n[0] = math.Float64bits(v)
}

func (w *Writer) String(v string) {
	_, _, s, _ := w.next()
	s[0] = v
}

func (w *Writer) Error(err error) {
	_, _, _, r := w.next()
	if err == nil {
		r[0] = nil
		return
	}
	r[0] = err
}

func (w *Writer) Handle(ref js.Ref) {
	_, _, _, r := w.next()
	r[0] = ref.Value()
}

func (w *Writer) Big(decimal string) {
	_, _, _, r := w.next()
	r[0] = bigParse(decimal)
}

// StateValue writes the state of a token, see [Host.TokenValue].
func (w *Writer) StateValue(tok any) {
	t, _, _, r := w.next()
	o, ok := w.m.host.TokenValue(tok)
	if _, isObj := o.(*object); !ok || !isObj {
		panic(&vmError{err: fmt.Errorf("the state of %s belongs to another engine", t.elem.name)})
	}
	r[0] = o
}

func (w *Writer) Record() {
	t, n, s, r := w.next()
	if t.kind == kShared {
		o := newObject(t, t.obj)
		r[0] = o
		w.stack = append(w.stack, level{kind: lvRecord, t: t, n: o.n, s: o.s, r: o.r})
		return
	}
	w.stack = append(w.stack, level{kind: lvRecord, t: t, n: n, s: s, r: r})
}

func (w *Writer) List(n int) {
	t, _, _, r := w.next()
	el := t.elem
	l := &list{len: int32(n), n: make([]uint64, n*int(el.l.n)), s: make([]string, n*int(el.l.s)), r: make([]any, n*int(el.l.r))}
	r[0] = l
	w.stack = append(w.stack, level{kind: lvList, t: el, l: l})
}

func (w *Writer) Map(int) {
	t, _, _, r := w.next()
	d := newDict(t.key)
	r[0] = d
	w.stack = append(w.stack, level{kind: lvMap, key: t.key, el: t.elem, d: d, kn: make([]uint64, t.key.l.n), ks: make([]string, t.key.l.s)})
}

func (w *Writer) Variant(tag string) {
	t, _, _, r := w.next()
	vt := w.m.prog.byName[tag]
	if vt == nil {
		panic(&vmError{err: fmt.Errorf("%s is no variant of %s", tag, t.name)})
	}
	o := newObject(vt, vt.l)
	r[0] = o
	switch {
	case t.handles || vt.kind == kHandle:
		w.stack = append(w.stack, level{kind: lvSingle, t: vt, n: o.n, s: o.s, r: o.r})
	case vt.kind == kShared:
		inner := newObject(vt, vt.obj)
		o.r[0] = inner
		w.stack = append(w.stack, level{kind: lvRecord, t: vt, n: inner.n, s: inner.s, r: inner.r})
	default:
		w.stack = append(w.stack, level{kind: lvRecord, t: vt, n: o.n, s: o.s, r: o.r})
	}
}

func (w *Writer) End() {
	w.stack = w.stack[:len(w.stack)-1]
}

// fastCall is a fast call of a closure, which may be nested in another one, e.g. a callback, which a host function
// calls during a call.
type fastCall struct {
	cl    *Closure
	base  layout
	saved saved
	w     Writer
	r     Reader
}

// BeginFast begins a fast call of the closure above the frames in use and returns the writer of its arguments.
func (m *Machine) BeginFast(ctx context.Context, cl *Closure) *Writer {
	if m.fast == len(m.calls) {
		m.calls = append(m.calls, &fastCall{w: Writer{m: m}, r: Reader{m: m}})
	}
	c := m.calls[m.fast]
	m.fast++
	c.cl = cl
	c.saved = saved{hi: m.hi, depth: m.depth, ctx: m.ctx}
	m.ctx = ctx
	base := m.hi
	fn := cl.fn
	m.ensure(int(base.n+fn.frame.n), int(base.s+fn.frame.s), int(base.r+fn.frame.r))
	// the arguments belong to the frames in use, so that a nested call does not overwrite them
	m.hi = base.add(fn.params.l)
	c.base = base
	c.w.reset(fn.params, m.n[base.n:], m.s[base.s:], m.r[base.r:])
	return &c.w
}

// InvokeFast runs the closure of the fast call and returns the reader of its results.
func (m *Machine) InvokeFast() (rd *Reader, err error) {
	c := m.calls[m.fast-1]
	defer m.recoverFast(&err)
	base := c.base
	m.hi = c.saved.hi
	m.runClosure(c.cl, int(base.n), int(base.s), int(base.r))
	// the results stay in the frames in use until End
	m.hi = base.add(c.cl.fn.results.l)
	c.r.reset(c.cl.fn.results, m.n[base.n:], m.s[base.s:], m.r[base.r:])
	return &c.r, nil
}

func (m *Machine) recoverFast(err *error) {
	v := recover()
	if v == nil {
		return
	}
	if e, ok := v.(*vmError); ok {
		*err = e.err
		return
	}
	if debugPanics {
		panic(v)
	}
	*err = &PanicError{Value: v}
}

// EndFast ends the fast call.
func (m *Machine) EndFast() {
	m.fast--
	c := m.calls[m.fast]
	m.reset(c.saved)
}
