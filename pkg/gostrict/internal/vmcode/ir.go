// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import "go/token"

// The IR is a typed tree, which the lowering builds from the checked syntax: every variable has its slots, every
// operator its type and every call its target. The backends translate it into register bytecode or closures.

type expr interface {
	typ() *tinfo
}

type (
	// eConst is a constant scalar (bits), string or nil reference.
	eConst struct {
		t *tinfo
		n uint64
		s string
		// r is a constant reference, e.g. the shared zero value of a shared struct
		r any
	}
	// eLocal is a variable or a part of it, e.g. a field of a struct in its slots.
	eLocal struct {
		t   *tinfo
		v   *local
		off layout
	}
	// eGlobal is a package variable or a part of it.
	eGlobal struct {
		t   *tinfo
		off layout
	}
	// eField is a field of a struct value, which is not a variable, e.g. the result of a call.
	eField struct {
		t   *tinfo
		x   expr
		off layout
	}
	// eLoad loads a value of the object of a pointer or shared struct p at the offset.
	eLoad struct {
		t   *tinfo
		p   expr
		off layout
	}
	eUnary struct {
		t  *tinfo
		op op
		x  expr
	}
	eBinary struct {
		t    *tinfo
		op   op
		x, y expr
		// swap passes the operands in the opposite order, after they were evaluated in their order, e.g. a > b as
		// b < a
		swap bool
		// imm is set, if y is a constant, which the operation takes as an operand, e.g. the count of a shift
		imm bool
	}
	// eAndOr is && or ||.
	eAndOr struct {
		t    *tinfo
		and  bool
		x, y expr
	}
	eCall struct {
		t    *tinfo
		fn   *Func
		args []expr
	}
	// eCallValue calls a function value.
	eCallValue struct {
		t    *tinfo
		f    expr
		args []expr
	}
	// eNative calls a function of the standard library, which the VM implements.
	eNative struct {
		t      *tinfo
		nat    *native
		params *tinfo
		args   []expr
	}
	// eListAppend appends the elements to the list l.
	eListAppend struct {
		t     *tinfo
		l     expr
		elems []expr
	}
	// eHost calls a function of a header package, which the host implements in Go.
	eHost struct {
		t    *tinfo
		ext  *extern
		args []expr
	}
	// eStruct is a composite literal of a struct with all fields in their order.
	eStruct struct {
		t      *tinfo
		fields []expr
		// order is the order of the evaluation of the fields, which is their order in the source
		order []int
	}
	// eNew allocates an object with the value x, e.g. &T{...}.
	eNew struct {
		t *tinfo
		x expr
	}
	// eFunc is a function as a value.
	eFunc struct {
		t  *tinfo
		fn *Func
	}
	// eClosure is a function literal, which captures the values of caps.
	eClosure struct {
		t    *tinfo
		fn   *Func
		caps []expr
	}
	// eBox converts the value x of the variant v into an interface.
	eBox struct {
		t *tinfo
		x expr
		v *tinfo
	}
	// eListOf creates a list of the elements.
	eListOf struct {
		t     *tinfo
		elems []expr
	}
	// eListGet is l.Get(i) with the element and whether i is in range.
	eListGet struct {
		t    *tinfo
		l, i expr
		// unchecked is set within a loop over the indices of the list, which only reads the element
		unchecked bool
	}
	// eCallIface calls the method of the interface value x.
	eCallIface struct {
		t      *tinfo
		x      expr
		method string
		args   []expr
	}
	// eTypeIs tells, whether the interface value x holds a value of type v.
	eTypeIs struct {
		t *tinfo
		x expr
		v *tinfo
	}
	// eView is a pointer to the part at off of the object of p, e.g. the address of a field.
	eView struct {
		t   *tinfo
		p   expr
		off layout
	}
	// eTupleAt is the element of a tuple.
	eTupleAt struct {
		t *tinfo
		x expr
		i int
	}
)

func (e *eConst) typ() *tinfo      { return e.t }
func (e *eLocal) typ() *tinfo      { return e.t }
func (e *eGlobal) typ() *tinfo     { return e.t }
func (e *eField) typ() *tinfo      { return e.t }
func (e *eLoad) typ() *tinfo       { return e.t }
func (e *eUnary) typ() *tinfo      { return e.t }
func (e *eBinary) typ() *tinfo     { return e.t }
func (e *eAndOr) typ() *tinfo      { return e.t }
func (e *eCall) typ() *tinfo       { return e.t }
func (e *eCallValue) typ() *tinfo  { return e.t }
func (e *eNative) typ() *tinfo     { return e.t }
func (e *eHost) typ() *tinfo       { return e.t }
func (e *eStruct) typ() *tinfo     { return e.t }
func (e *eNew) typ() *tinfo        { return e.t }
func (e *eFunc) typ() *tinfo       { return e.t }
func (e *eClosure) typ() *tinfo    { return e.t }
func (e *eCallIface) typ() *tinfo  { return e.t }
func (e *eTypeIs) typ() *tinfo     { return e.t }
func (e *eView) typ() *tinfo       { return e.t }
func (e *eBox) typ() *tinfo        { return e.t }
func (e *eListOf) typ() *tinfo     { return e.t }
func (e *eListGet) typ() *tinfo    { return e.t }
func (e *eTupleAt) typ() *tinfo    { return e.t }
func (e *eListAppend) typ() *tinfo { return e.t }

// local is a variable of a function with its slots in the frame.
type local struct {
	name string
	t    *tinfo
	at   layout
	// cell is the type of a variable, which lives in a cell, whose pointer the local holds
	cell *tinfo
	// raw is the plain variable, which an assignment writes, before it moves the value into the cell
	raw *local
}

type stmt interface{}

type (
	// sAssign assigns the values to the targets. All values are evaluated before the first assignment. A single
	// value of a tuple assigns its elements.
	sAssign struct {
		lhs []lval
		rhs []expr
	}
	sExpr struct {
		x expr
	}
	sIf struct {
		cond      expr
		then, els []stmt
	}
	// sLoop runs the body, while cond is true or nil, and the post statements after the body and each continue.
	sLoop struct {
		cond       expr
		body, post []stmt
		pos        token.Pos
	}
	// sBlock is the target of a break out of a switch.
	sBlock struct {
		body []stmt
	}
	sBreak struct {
		target stmt
	}
	sContinue struct {
		target *sLoop
	}
	sReturn struct {
		vals []expr
	}
)

// lval is the target of an assignment.
type lval interface {
	ltyp() *tinfo
}

type (
	lLocal struct {
		t   *tinfo
		v   *local
		off layout
	}
	lGlobal struct {
		t   *tinfo
		off layout
	}
	// lStore stores into the object of the pointer p at the offset.
	lStore struct {
		t   *tinfo
		p   expr
		off layout
	}
	lDiscard struct {
		t *tinfo
	}
)

func (l *lLocal) ltyp() *tinfo   { return l.t }
func (l *lGlobal) ltyp() *tinfo  { return l.t }
func (l *lStore) ltyp() *tinfo   { return l.t }
func (l *lDiscard) ltyp() *tinfo { return l.t }
