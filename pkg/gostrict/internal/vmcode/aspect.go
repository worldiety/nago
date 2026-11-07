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
)

// The entry points of a whole program run within the aspect of the host, when the host or another package calls
// them: the hook of an entry point calls $enter with its full name and whether its last result is an error, then the
// entry point, then $exit with its error result. A failure of $enter, e.g. a denied call, aborts the whole call
// without an exit of the denied frame. An error, which $enter returns, e.g. a denial, which the aspect marks as a
// result, is the error result of the entry point with zero values, without its body and without an exit. An abort
// of the entry point, e.g. by a host function, exits its frame with an error, while the abort goes on.

// Names of the functions of the aspect, see gostrict.EnterFunc and gostrict.ExitFunc.
const (
	enterFunc = "$enter"
	exitFunc  = "$exit"
)

// entryInfo is the hook of an entry point.
type entryInfo struct {
	name        string
	hook, inner *Func
	enter, exit *extern
	// errAt is the result, which is an error, or nil
	errT  *tinfo
	errAt layout
}

var errAborted = errors.New("aborted")

// hookFor returns the hook of an entry point, a function with its signature, which calls it within the aspect.
func (c *compiler) hookFor(fn *types.Func) *Func {
	key := "H:" + fn.FullName()
	if f, ok := c.funcs[key]; ok {
		return f
	}
	inner := c.funcFor(fn, nil, nil)
	h := &Func{name: inner.name + "$h", params: inner.params, results: inner.results}
	h.closure = &Closure{fn: h}
	c.funcs[key] = h
	e := &entryInfo{name: c.src.Entries[fn], hook: h, inner: inner, enter: c.aspect(enterFunc), exit: c.aspect(exitFunc)}
	res := inner.results
	switch {
	case res.kind == kError:
		e.errT = res
	case res.kind == kTuple && len(res.elems) > 0 && res.elems[len(res.elems)-1].kind == kError:
		e.errT, e.errAt = res.elems[len(res.elems)-1], res.offs[len(res.elems)-1]
	}
	c.out.entries = append(c.out.entries, e)
	return h
}

// aspect returns the function of the aspect of the host.
func (c *compiler) aspect(name string) *extern {
	if x, ok := c.externs[name]; ok {
		return x
	}
	if !c.src.Hosted[name] {
		c.fail(token.NoPos, "the bundle has no aspect")
	}
	str := c.typeOf(types.Typ[types.String])
	x := &extern{name: name, idx: len(c.out.externs), params: []*tinfo{str}, results: tuple()}
	switch name {
	case enterFunc:
		// whether the last result of the entry point is an error, and the error, which it returns instead of running
		x.params = append(x.params, c.typeOf(types.Typ[types.Bool]))
		x.results = c.typeOf(errorType)
	case exitFunc:
		x.params = append(x.params, c.typeOf(errorType))
	}
	c.externs[name] = x
	c.out.externs = append(c.out.externs, x)
	return x
}

// resultZero writes the zero value of t into the slots, which a checked program may use, for the results of a denied
// entry point: an empty list or map instead of nil, also in the fields of a struct, while a pointer, a function, an
// interface and a handle are nil. A shared struct is a new object, so that no denial shares a list with another.
func resultZero(t *tinfo, n []uint64, s []string, r []any) {
	switch t.kind {
	case kTuple:
		for i, et := range t.elems {
			at := t.offs[i]
			resultZero(et, n[at.n:], s[at.s:], r[at.r:])
		}
	case kStruct:
		for _, f := range t.fields {
			resultZero(f.t, n[f.off.n:], s[f.off.s:], r[f.off.r:])
		}
	case kShared:
		o := newObject(t, t.obj)
		for _, f := range t.fields {
			resultZero(f.t, o.n[f.off.n:], o.s[f.off.s:], o.r[f.off.r:])
		}
		r[0] = o
	case kList, kOut:
		r[0] = &list{}
	case kMap:
		r[0] = newDict(t.key)
	case kBig:
		r[0] = bigZero
	default:
		clear(n[:t.l.n])
		clear(s[:t.l.s])
		clear(r[:t.l.r])
	}
}

// callEntry calls the entry point of a hook within the aspect at the frame, where its arguments are. An error, which
// $enter returns, replaces the arguments with the results of the entry point: zero values, see [resultZero], and the
// error. A function of the host, which ends the entry point with an error (gostrict.ReturnError), unwinds its calls
// and returns the same results after the exit with the error; in an entry point without an error result it aborts
// the whole call instead, so that no entry point around it takes the error as its result.
func (m *Machine) callEntry(e *entryInfo, bn, bs, br int) {
	if res := m.host.Call(e.enter.idx, []any{e.name, e.errT != nil}); len(res) > 0 && res[0] != nil {
		if e.errT == nil {
			// gostrict.Funcs.WithAspect aborts instead; a host, which implements $enter itself, fails loudly
			panic(&vmError{err: fmt.Errorf("vm: $enter returned an error for %s, whose last result is no error", e.name)})
		}
		m.returnErr(e, res[0].(error), bn, bs, br)
		return
	}
	s := saved{hi: m.hi, depth: m.depth, ctx: m.ctx}
	done := false
	defer func() {
		if done {
			return
		}
		v := recover()
		err, returned := m.host.Returned(v)
		if returned && e.errT != nil {
			// a function of the host ended the entry point with an error as its result
			m.reset(s)
			m.returnErr(e, err, bn, bs, br)
			m.host.Call(e.exit.idx, []any{e.name, err})
			return
		}
		m.host.Call(e.exit.idx, []any{e.name, errAborted})
		if returned {
			// an entry point without an error result aborts the whole call, also within one with an error result
			panic(m.host.Abort(err))
		}
		panic(v)
	}()
	m.enter()
	m.run(e.inner, bn, bs, br)
	m.depth--
	done = true
	var err any
	if e.errT != nil {
		at := layout{int32(bn), int32(bs), int32(br)}.add(e.errAt)
		err = m.neutral(e.errT, m.n[at.n:], m.s[at.s:], m.r[at.r:])
	}
	m.host.Call(e.exit.idx, []any{e.name, err})
}

// returnErr writes the results of an entry point, which returns err without its body: zero values, see [resultZero],
// and err.
func (m *Machine) returnErr(e *entryInfo, err error, bn, bs, br int) {
	resultZero(e.inner.results, m.n[bn:], m.s[bs:], m.r[br:])
	at := layout{int32(bn), int32(bs), int32(br)}.add(e.errAt)
	m.write(err, e.errT, m.n[at.n:], m.s[at.s:], m.r[at.r:])
}
