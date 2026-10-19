// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"errors"
	"go/token"
	"go/types"
)

// The entry points of a whole program run within the aspect of the host, when the host or another package calls
// them: the hook of an entry point calls $enter with its full name, then the entry point, then $exit with its error
// result. A failure of $enter, e.g. a denied call, aborts the whole call without an exit of the denied frame. An
// abort of the entry point, e.g. by a host function, exits its frame with an error, while the abort goes on.

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
	if name == exitFunc {
		x.params = append(x.params, c.typeOf(errorType))
	}
	c.externs[name] = x
	c.out.externs = append(c.out.externs, x)
	return x
}

// callEntry calls the entry point of a hook within the aspect at the frame, where its arguments are.
func (m *Machine) callEntry(e *entryInfo, bn, bs, br int) {
	m.host.Call(e.enter.idx, []any{e.name})
	done := false
	defer func() {
		if !done {
			m.host.Call(e.exit.idx, []any{e.name, errAborted})
		}
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
