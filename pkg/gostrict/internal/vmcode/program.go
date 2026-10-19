// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"sync"
	"sync/atomic"
)

// Program is a compiled program. It never changes after its compilation and its package variables never change
// after their initialization, so all instances share it.
type Program struct {
	funcs   []*Func
	externs []*extern
	exports map[string]*Func
	// the constants of the bytecode
	nconst []uint64
	sconst []string
	rconst []any
	// the package variables, which init writes once
	globals layout
	init    *Func
	once    sync.Once
	initErr error
	gn      []uint64
	gs      []string
	gr      []any
	// types are the types, which instructions refer to, e.g. of new objects
	types []*tinfo
	// entries are the hooks of the entry points of a whole program
	entries []*entryInfo
	// byName are the records and variants of the boundary by their full name
	byName map[string]*tinfo
}

// Func is a compiled function.
type Func struct {
	name    string
	params  *tinfo
	results *tinfo
	// frame are the slots of the frame: the parameters, the variables and the temporaries
	frame layout
	code  []instr
	// lines are the positions of the instructions, see [Func.pos]
	calls    []callSite
	aux      []aux
	closures []closureSite
	// body is the IR, which the backends translate
	body   []stmt
	locals layout
	// closure is the value of a function, which captures nothing, e.g. a callback
	closure *Closure
	// caps are the slots of the captured variables of a function literal, see [Closure]
	caps []capSlot
}

// capSlot is the place of a captured variable in the frame of a function literal.
type capSlot struct {
	at layout
	l  layout
}

// extern is a function of a header package, which the host implements.
type extern struct {
	name    string
	params  []*tinfo
	results *tinfo
	idx     int
	// variable is set for a variable of a header package, whose value the function returns
	variable bool
}

// callSite is the target of a call and the offset of its arguments and results in the frame of the caller.
type callSite struct {
	fn   *Func
	nat  *native
	ext  *extern
	base layout
	// method is the name of the method of an interface, whose arguments are in the slots at args
	method string
	args   layout
	argsL  layout
}

// closureSite creates a closure of fn with the captured slots at caps into r[dst].
type closureSite struct {
	fn   *Func
	dst  int32
	caps []layout
}

// aux are the operands of an instruction, which do not fit into it, e.g. the slots of a struct.
type aux struct {
	t   *tinfo
	dst layout
	src layout
	off layout
}

// instr is an instruction: the operation and up to three operands, which are slots of the frame, constants,
// targets of jumps or indices of call sites.
type instr struct {
	op      op
	a, b, c int32
}

// object is a struct, which a pointer, a shared struct or an interface refers to.
type object struct {
	t *tinfo
	n []uint64
	s []string
	r []any
	// cache is the neutral value of a shared struct, which never changes, e.g. a style, which many cells of a
	// table refer to
	cache atomic.Pointer[any]
	// kept is the Go value of a shared struct, which a reader of a fast call keeps, see [Reader.Keep]
	kept atomic.Pointer[any]
}

func newObject(t *tinfo, l layout) *object {
	o := &object{t: t}
	if l.n > 0 {
		o.n = make([]uint64, l.n)
	}
	if l.s > 0 {
		o.s = make([]string, l.s)
	}
	if l.r > 0 {
		o.r = make([]any, l.r)
	}
	return o
}

// list is a list.List, list.Frozen or msg.Out. Element i occupies the slots from i times the layout of an
// element.
type list struct {
	n   []uint64
	s   []string
	r   []any
	len int32
}

// Closure is a function value: the function and the slots of the variables, which it captured, in the order of
// [Func.caps].
type Closure struct {
	fn *Func
	n  []uint64
	s  []string
	r  []any
}

// Name returns the full name of the function.
func (c *Closure) Name() string {
	return c.fn.name
}

type op uint8

const (
	opNop op = iota
	opMovN
	opMovS
	opMovR
	opLdiN // n[a] = int64(b)
	opLdkN // n[a] = nconst[b]
	opLdkS
	opLdkR
	opLdgN // n[a] = gn[b]
	opLdgS
	opLdgR
	opStgN // gn[a] = n[b]
	opStgS
	opStgR
	opLdoN // n[a] = r[b].n[c]
	opLdoS
	opLdoR
	opStoN // r[a].n[b] = n[c]
	opStoS
	opStoR
	opNew // r[a] = new object of types[b]

	opAddI
	opAddIK // n[a] = n[b] + c
	opMulIK // n[a] = n[b] * c
	opSubI
	opMulI
	opDivI
	opRemI
	opDivU
	opRemU
	opAndI
	opOrI
	opXorI
	opAndNotI
	opShlI // n[a] = n[b] << c
	opShrI // signed
	opShrU // unsigned
	opNegI
	opNotI // ^x
	opEqI
	opNeI
	opLtI
	opLeI
	opLtU
	opLeU

	opAddF
	opSubF
	opMulF
	opDivF
	opNegF
	opEqF
	opNeF
	opLtF
	opLeF
	opI2F
	opU2F

	opNotB
	opIsNil // n[a] = r[b] == nil
	opEqR   // n[a] = r[b] == r[c], e.g. of pointers
	opNeR
	opEqV // aux[a]: n[dst.n] = the values of type t at src and off are equal
	opConcat
	opEqS
	opNeS
	opLtS
	opLeS

	opJmp  // pc = a
	opLoop // pc = a, a backward jump, which burns fuel
	opJmpT // if n[a] != 0 { pc = b }
	opJmpF
	// fused comparisons, which jump, if the comparison of b and c fails, to a
	opJNltI
	opJNleI
	opJNeqI
	opJNneI
	opJNltF
	opJNleF
	opJEqIK // if n[b] == c { pc = a }
	opJNeIK // if n[b] != c { pc = a }

	opCall       // calls[a]
	opCallNative // calls[a]
	opCallHost   // calls[a]
	opCallValue  // calls[a] with the closure in r[b]
	opRet

	opLLen   // n[a] = len(r[b])
	opLGetN  // n[a] = r[b].n[n[c]] for an element of one scalar
	opLGetS  // s[a] = r[b].s[n[c]]
	opLGetR  // r[a] = r[b].r[n[c]]
	opLGet   // aux[a]: dst = element n[c] of r[b], without the check of the range
	opLGetOk // aux[a]: dst, ok = element n[c] of r[b]
	opLAppN  // append n[b] to r[a]
	opLAppS
	opLAppR
	opLApp      // aux[a]: append the element at src to r[b]
	opLNew      // r[a] = new list
	opLSet      // aux[a]: n[dst.n + el.n] = set element n[c] of r[b] to src
	opBox       // aux[a]: r[dst.r] = box of the value at src of type t
	opZeroR     // r[a] = nil
	opClosure   // closures[a]: r[dst] = a closure, which captures the slots
	opCallIface // calls[a]: the method of the interface value in r[b]
	opTypeIs    // n[a] = the box in r[b] holds a value of types[c]
	opView      // aux[a]: r[dst.r] = a pointer into the object of r[src.r] at off
	opEntry     // calls the entry point of entries[a] within the aspect
)
