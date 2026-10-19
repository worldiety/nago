// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

// The inliner replaces a call of a small function, whose body only returns an expression, e.g. a method of a vector,
// by the expression. A parameter becomes its argument, if the argument is a variable or a constant, which no other
// argument can change, or a temporary, which the arguments are evaluated into in their order before the expression,
// see eLet.

// maxInline is the number of nodes of the expression of a function, up to which it is inlined.
const maxInline = 24

type (
	// eLet evaluates the values into the variables and then the expression.
	eLet struct {
		t    *tinfo
		vars []*local
		vals []expr
		x    expr
	}
)

func (e *eLet) typ() *tinfo { return e.t }

// inlineBody returns the expression, which a function returns, if it is small enough to be inlined.
func inlineBody(f *Func) (expr, bool) {
	if len(f.body) != 1 || len(f.caps) > 0 || f.body == nil {
		return nil, false
	}
	ret, ok := f.body[0].(*sReturn)
	if !ok || len(ret.vals) != 1 || f.results.kind == kTuple {
		return nil, false
	}
	size := 0
	inlinable := true
	walk(ret.vals[0], func(e expr) {
		size++
		switch e.(type) {
		case *eClosure, *eLet, *eCallValue, *eCallIface, *eHost:
			// a closure needs its frame, a host function may call the program
			inlinable = false
		case *eCall:
			// the callee is inlined in a later round, if it can
		}
	})
	return ret.vals[0], inlinable && size <= maxInline
}

// inliner inlines the calls of one function.
type inliner struct {
	c *compiler
	f *Func
}

// inline inlines the small functions into the functions of the program, a few levels deep.
func (c *compiler) inline() {
	for range 3 {
		for _, f := range c.out.funcs {
			in := &inliner{c: c, f: f}
			for _, s := range f.body {
				in.stmt(s)
			}
		}
	}
}

func (in *inliner) stmt(s stmt) {
	switch s := s.(type) {
	case *sAssign:
		for i, x := range s.rhs {
			s.rhs[i] = in.expr(x)
		}
		for _, lv := range s.lhs {
			if st, ok := lv.(*lStore); ok {
				st.p = in.expr(st.p)
			}
		}
	case *sExpr:
		s.x = in.expr(s.x)
	case *sIf:
		s.cond = in.expr(s.cond)
		for _, x := range s.then {
			in.stmt(x)
		}
		for _, x := range s.els {
			in.stmt(x)
		}
	case *sLoop:
		if s.cond != nil {
			s.cond = in.expr(s.cond)
		}
		for _, x := range s.body {
			in.stmt(x)
		}
		for _, x := range s.post {
			in.stmt(x)
		}
	case *sBlock:
		for _, x := range s.body {
			in.stmt(x)
		}
	case *sReturn:
		for i, x := range s.vals {
			s.vals[i] = in.expr(x)
		}
	}
}

// expr replaces the inlinable calls in x and its operands.
func (in *inliner) expr(x expr) expr {
	switch e := x.(type) {
	case *eField:
		e.x = in.expr(e.x)
	case *eLoad:
		e.p = in.expr(e.p)
	case *eUnary:
		e.x = in.expr(e.x)
	case *eBinary:
		e.x, e.y = in.expr(e.x), in.expr(e.y)
	case *eAndOr:
		e.x, e.y = in.expr(e.x), in.expr(e.y)
	case *eCallValue:
		e.f = in.expr(e.f)
		in.exprs(e.args)
	case *eNative:
		in.exprs(e.args)
	case *eHost:
		in.exprs(e.args)
	case *eStruct:
		in.exprs(e.fields)
	case *eNew:
		e.x = in.expr(e.x)
	case *eClosure:
		in.exprs(e.caps)
	case *eBox:
		e.x = in.expr(e.x)
	case *eListOf:
		in.exprs(e.elems)
	case *eListGet:
		e.l, e.i = in.expr(e.l), in.expr(e.i)
	case *eListAppend:
		e.l = in.expr(e.l)
		in.exprs(e.elems)
	case *eCallIface:
		e.x = in.expr(e.x)
		in.exprs(e.args)
	case *eTypeIs:
		e.x = in.expr(e.x)
	case *eView:
		e.p = in.expr(e.p)
	case *eTupleAt:
		e.x = in.expr(e.x)
	case *eLet:
		in.exprs(e.vals)
		e.x = in.expr(e.x)
	case *eCall:
		in.exprs(e.args)
		if e.fn != in.f {
			if body, ok := inlineBody(e.fn); ok {
				return in.call(e, body)
			}
		}
	}
	return x
}

func (in *inliner) exprs(xs []expr) {
	for i, x := range xs {
		xs[i] = in.expr(x)
	}
}

// call returns the body of the callee of the call with its parameters replaced by the arguments.
func (in *inliner) call(call *eCall, body expr) expr {
	params := map[*local]expr{}
	let := &eLet{t: call.t}
	callee := call.fn
	// the parameters are the first variables of the callee in the order of its tuple
	byAt := map[layout]int{}
	for i, off := range callee.params.offs {
		byAt[off] = i
	}
	var fail bool
	used := make([]bool, len(call.args))
	walk(body, func(e expr) {
		l, ok := e.(*eLocal)
		if !ok || params[l.v] != nil {
			return
		}
		i, ok := byAt[l.v.at]
		if !ok || i >= len(call.args) || l.v.t != callee.params.elems[i] {
			fail = true
			return
		}
		used[i] = true
		arg := call.args[i]
		switch arg.(type) {
		case *eConst, *eLocal, *eGlobal:
			params[l.v] = arg
		default:
			v := in.newLocal(l.v.t)
			let.vars = append(let.vars, v)
			let.vals = append(let.vals, arg)
			params[l.v] = &eLocal{t: v.t, v: v}
		}
	})
	if fail {
		return call
	}
	if len(call.args) != len(callee.params.elems) {
		// f(g()) passes a tuple
		return call
	}
	// the arguments, which the body does not use, are still evaluated for their effects
	for i, a := range call.args {
		if used[i] {
			continue
		}
		switch a.(type) {
		case *eConst, *eLocal, *eGlobal:
			continue
		}
		v := in.newLocal(a.typ())
		let.vars = append(let.vars, v)
		let.vals = append(let.vals, a)
	}
	// the arguments are evaluated in their order
	orderLet(let, call.args)
	let.x = subst(clone(body), params)
	if len(let.vars) == 0 {
		return let.x
	}
	return let
}

// orderLet sorts the bound values of the let in the order of the arguments.
func orderLet(let *eLet, args []expr) {
	pos := map[expr]int{}
	for i, a := range args {
		pos[a] = i
	}
	for i := 1; i < len(let.vals); i++ {
		for j := i; j > 0 && pos[let.vals[j]] < pos[let.vals[j-1]]; j-- {
			let.vals[j], let.vals[j-1] = let.vals[j-1], let.vals[j]
			let.vars[j], let.vars[j-1] = let.vars[j-1], let.vars[j]
		}
	}
}

func (in *inliner) newLocal(t *tinfo) *local {
	v := &local{name: "$inl", t: t, at: in.f.locals}
	in.f.locals = in.f.locals.add(t.l)
	return v
}

// subst replaces the parameters in a copy of an expression.
func subst(x expr, params map[*local]expr) expr {
	switch e := x.(type) {
	case *eLocal:
		if p, ok := params[e.v]; ok {
			if e.off == (layout{}) && e.t == p.typ() {
				return p
			}
			// a part of a struct parameter
			switch a := p.(type) {
			case *eLocal:
				return &eLocal{t: e.t, v: a.v, off: a.off.add(e.off)}
			case *eGlobal:
				return &eGlobal{t: e.t, off: a.off.add(e.off)}
			}
			return &eField{t: e.t, x: p, off: e.off}
		}
		return e
	case *eField:
		e.x = subst(e.x, params)
		if l, ok := e.x.(*eLocal); ok {
			return &eLocal{t: e.t, v: l.v, off: l.off.add(e.off)}
		}
	case *eLoad:
		e.p = subst(e.p, params)
	case *eUnary:
		e.x = subst(e.x, params)
	case *eBinary:
		e.x, e.y = subst(e.x, params), subst(e.y, params)
	case *eAndOr:
		e.x, e.y = subst(e.x, params), subst(e.y, params)
	case *eCall:
		substAll(e.args, params)
	case *eNative:
		substAll(e.args, params)
	case *eStruct:
		substAll(e.fields, params)
	case *eNew:
		e.x = subst(e.x, params)
	case *eBox:
		e.x = subst(e.x, params)
	case *eListOf:
		substAll(e.elems, params)
	case *eListGet:
		e.l, e.i = subst(e.l, params), subst(e.i, params)
	case *eListAppend:
		e.l = subst(e.l, params)
		substAll(e.elems, params)
	case *eTypeIs:
		e.x = subst(e.x, params)
	case *eView:
		e.p = subst(e.p, params)
	case *eTupleAt:
		e.x = subst(e.x, params)
	}
	return x
}

func substAll(xs []expr, params map[*local]expr) {
	for i, x := range xs {
		xs[i] = subst(x, params)
	}
}

// clone copies an expression, whose nodes subst changes.
func clone(x expr) expr {
	switch e := x.(type) {
	case *eConst:
		c := *e
		return &c
	case *eLocal:
		c := *e
		return &c
	case *eGlobal:
		c := *e
		return &c
	case *eField:
		c := *e
		c.x = clone(e.x)
		return &c
	case *eLoad:
		c := *e
		c.p = clone(e.p)
		return &c
	case *eUnary:
		c := *e
		c.x = clone(e.x)
		return &c
	case *eBinary:
		c := *e
		c.x, c.y = clone(e.x), clone(e.y)
		return &c
	case *eAndOr:
		c := *e
		c.x, c.y = clone(e.x), clone(e.y)
		return &c
	case *eCall:
		c := *e
		c.args = cloneAll(e.args)
		return &c
	case *eNative:
		c := *e
		c.args = cloneAll(e.args)
		return &c
	case *eStruct:
		c := *e
		c.fields = cloneAll(e.fields)
		return &c
	case *eNew:
		c := *e
		c.x = clone(e.x)
		return &c
	case *eFunc:
		c := *e
		return &c
	case *eBox:
		c := *e
		c.x = clone(e.x)
		return &c
	case *eListOf:
		c := *e
		c.elems = cloneAll(e.elems)
		return &c
	case *eListGet:
		c := *e
		c.l, c.i = clone(e.l), clone(e.i)
		return &c
	case *eListAppend:
		c := *e
		c.l = clone(e.l)
		c.elems = cloneAll(e.elems)
		return &c
	case *eTypeIs:
		c := *e
		c.x = clone(e.x)
		return &c
	case *eView:
		c := *e
		c.p = clone(e.p)
		return &c
	case *eTupleAt:
		c := *e
		c.x = clone(e.x)
		return &c
	}
	panic("vm: clone of an expression, which is not inlined")
}

func cloneAll(xs []expr) []expr {
	res := make([]expr, len(xs))
	for i, x := range xs {
		res[i] = clone(x)
	}
	return res
}
