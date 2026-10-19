// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package vmcode compiles a checked program into register bytecode and executes it, see package engine/vm. It does
// not depend on package gostrict, which creates the program of each bundle with it, so that a bundle carries its
// compiled program instead of the checked syntax. The engine provides the values of the boundary, which only
// package gostrict knows, through a [Host].
package vmcode

import (
	"context"
	"fmt"
	"os"
)

// Host is the environment of a machine, which the engine provides: the functions of header packages and the
// neutral values of the boundary, which belong to package gostrict.
type Host interface {
	// Call calls the function of a header package with the index of [Program.Externs] with neutral values and
	// returns its results.
	Call(idx int, args []any) []any
	// Variant returns the neutral value of a variant of a union.
	Variant(tag string, fields []any) any
	// VariantOf returns the name and fields of the neutral value of a variant.
	VariantOf(v any) (tag string, fields []any)
	// Token returns the token of a state of the struct with the full name.
	Token(name string, state any) any
	// TokenValue returns the state of a token, or false, if it belongs to another engine.
	TokenValue(v any) (any, bool)
	// Callback returns the neutral value of a function of the program.
	Callback(fn *Closure) any
	// ProgramError returns the error, which the program created at the position with the index pos of the table
	// of the bundle, see [Source.ErrorPos].
	ProgramError(msg string, pos int) error
}

// debugPanics lets the panics of the machine through with their stack, e.g. GOSTRICT_VM_PANIC=1 go test.
var debugPanics = os.Getenv("GOSTRICT_VM_PANIC") != ""

// PanicError is a panic of a host function, which a call of the machine recovered, e.g. an abort.
type PanicError struct {
	Value any
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("vm: %v", p.Value)
}

// Export returns the exported function with the name, or nil.
func (p *Program) Export(name string) *Closure {
	if f := p.exports[name]; f != nil {
		return f.closure
	}
	return nil
}

// Externs returns the full names of the functions of header packages, which the host implements, by their index.
func (p *Program) Externs() []string {
	names := make([]string, len(p.externs))
	for i, x := range p.externs {
		names[i] = x.name
	}
	return names
}

// NewMachine returns a machine, which executes the program. The package variables of the program are initialized
// once, by the first machine.
func NewMachine(ctx context.Context, p *Program, h Host, maxDepth int) (*Machine, error) {
	if maxDepth == 0 {
		maxDepth = 10000
	}
	m := &Machine{prog: p, host: h, maxDepth: maxDepth, fuel: fuelChunk}
	p.once.Do(func() {
		p.gn, p.gs, p.gr = make([]uint64, p.globals.n), make([]string, p.globals.s), make([]any, p.globals.r)
		_, p.initErr = m.Invoke(ctx, &Closure{fn: p.init}, nil)
	})
	if p.initErr != nil {
		return nil, p.initErr
	}
	if p.instInit != nil {
		// the variables, which hold values of the host, belong to the instance, like in a browser
		m.xn, m.xs, m.xr = make([]uint64, p.instGlobals.n), make([]string, p.instGlobals.s), make([]any, p.instGlobals.r)
		if _, err := m.Invoke(ctx, &Closure{fn: p.instInit}, nil); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Context returns the context of the current call, e.g. for a callback, which the host calls during the call.
func (m *Machine) Context() context.Context {
	return m.ctx
}

// saved is the state of the machine, which a call restores, also after a failure.
type saved struct {
	hi    layout
	depth int
	ctx   context.Context
}

// Invoke calls fn with neutral arguments above the frames in use, e.g. within a host function, which calls a
// callback, and returns its neutral results. A failure of the engine, e.g. an ended context, is returned as error,
// a panic of a host function as [PanicError].
func (m *Machine) Invoke(ctx context.Context, cl *Closure, args []any) (res []any, err error) {
	fn := cl.fn
	s := saved{hi: m.hi, depth: m.depth, ctx: m.ctx}
	m.ctx = ctx
	defer m.restore(s, &err)
	base := m.hi
	params := fn.params
	m.ensure(int(base.n+params.l.n), int(base.s+params.l.s), int(base.r+params.l.r))
	for i, a := range args {
		at := base.add(params.offs[i])
		m.write(a, params.elems[i], m.n[at.n:], m.s[at.s:], m.r[at.r:])
	}
	m.runClosure(cl, int(base.n), int(base.s), int(base.r))
	t := fn.results
	if t.kind != kTuple {
		return []any{m.neutral(t, m.n[base.n:], m.s[base.s:], m.r[base.r:])}, nil
	}
	res = make([]any, len(t.elems))
	for i, et := range t.elems {
		at := base.add(t.offs[i])
		res[i] = m.neutral(et, m.n[at.n:], m.s[at.s:], m.r[at.r:])
	}
	return res, nil
}

// restore restores the machine after a call and recovers from its failure.
func (m *Machine) restore(s saved, err *error) {
	m.reset(s)
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

// reset restores the state of the machine after a call.
func (m *Machine) reset(s saved) {
	m.hi, m.depth, m.ctx = s.hi, s.depth, s.ctx
	if s.hi == (layout{}) {
		// the outermost call releases the values of its frames
		clear(m.s[:m.peak.s])
		clear(m.r[:m.peak.r])
		m.peak = layout{}
	}
}

// callHost calls a function of the host with the arguments at base and writes its results over them.
func (m *Machine) callHost(x *extern, base layout) {
	args := make([]any, len(x.params))
	off := base
	for i, t := range x.params {
		args[i] = m.neutral(t, m.n[off.n:], m.s[off.s:], m.r[off.r:])
		off = off.add(t.l)
	}
	res := m.host.Call(x.idx, args)
	t := x.results
	if t.kind != kTuple {
		m.write(res[0], t, m.n[base.n:], m.s[base.s:], m.r[base.r:])
		return
	}
	for i, et := range t.elems {
		at := base.add(t.offs[i])
		m.write(res[i], et, m.n[at.n:], m.s[at.s:], m.r[at.r:])
	}
}

// compiler helpers for the constants of the bytecode

func (c *compiler) nconst(v uint64) int32 {
	if i, ok := c.nk[v]; ok {
		return i
	}
	c.out.nconst = append(c.out.nconst, v)
	c.nk[v] = int32(len(c.out.nconst) - 1)
	return c.nk[v]
}

func (c *compiler) sconst(v string) int32 {
	if i, ok := c.sk[v]; ok {
		return i
	}
	c.out.sconst = append(c.out.sconst, v)
	c.sk[v] = int32(len(c.out.sconst) - 1)
	return c.sk[v]
}

func (c *compiler) rconst(v any) int32 {
	for i, x := range c.out.rconst {
		if x == v {
			return int32(i)
		}
	}
	c.out.rconst = append(c.out.rconst, v)
	return int32(len(c.out.rconst) - 1)
}

func (c *compiler) typeIndex(t *tinfo) int32 {
	if i, ok := c.typeIdx[t]; ok {
		return i
	}
	c.out.types = append(c.out.types, t)
	c.typeIdx[t] = int32(len(c.out.types) - 1)
	return c.typeIdx[t]
}
