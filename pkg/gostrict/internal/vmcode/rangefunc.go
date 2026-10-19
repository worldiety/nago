// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"go/ast"
	"go/token"
	"go/types"
)

// A range over an iterator calls the iterator with its body as the function yield, like the JavaScript emitter. In
// the body, a continue of the loop returns true and a break false. A return of the enclosing function and a branch
// to a label outside of the body store, what is to do, in cells of the enclosing function and stop the iteration:
// $rr and the results for a return, $rb with the code of the branch. After the call of the iterator, the enclosing
// function performs it, or stops its own iteration, if it is the body of another range.

// rangeCtx is the body of a range over an iterator, which a fnLower lowers as a function.
type rangeCtx struct {
	label string
	// pending are the branches to labels outside of the body, which the enclosing function performs
	pending []branchCode
}

type branchCode struct {
	tok   token.Token
	label string
	code  int32
}

// realFunc returns the function, whose return a return statement ends: the nearest, which is no body of a range.
func (l *fnLower) realFunc() *fnLower {
	for l.rng != nil {
		l = l.parent
	}
	return l
}

// rangeVar returns a variable of the real function for the body of a range, which lives in a cell, created by the
// prologue of the real function.
func (l *fnLower) rangeVar(name string, t *tinfo, slot **types.Var) *local {
	r := l.realFunc()
	if *slot == nil {
		*slot = types.NewVar(token.NoPos, nil, name, nil)
		r.cells[*slot] = true
		v := r.newVar(*slot, name, t)
		r.prologue = append(r.prologue, r.initVar(v, r.c.zeroOrNil(t)))
	}
	v, _ := l.lookup(*slot)
	return v
}

// rangeResults returns the cells of the results of the real function, which a return in the body of a range sets.
func (l *fnLower) rangeResults() []*local {
	r := l.realFunc()
	res := r.f.results
	elems := res.elems
	if res.kind != kTuple {
		elems = []*tinfo{res}
	}
	for len(r.rv) < len(elems) {
		r.rv = append(r.rv, nil)
	}
	var vs []*local
	for i, t := range elems {
		vs = append(vs, l.rangeVar("$rv", t, &r.rv[i]))
	}
	return vs
}

func (l *fnLower) rangeFunc(s *ast.RangeStmt, label string) {
	sig := l.sub(l.info.TypeOf(s.X)).Underlying().(*types.Signature)
	yield := sig.Params().At(0).Type().Underlying().(*types.Signature)
	iter := l.expr(s.X)
	bt := l.c.boolType()

	inner := l.child(yield, "range")
	inner.rng = &rangeCtx{label: label}
	vars := []ast.Expr{s.Key, s.Value}
	var assigns []stmt
	for i := range yield.Params().Len() {
		pt := inner.f.params.elems[i]
		var x ast.Expr
		if i < len(vars) {
			x = vars[i]
		}
		id, _ := x.(*ast.Ident)
		switch {
		case x == nil || (id != nil && id.Name == "_"):
			inner.newLocal("_", pt)
		case s.Tok == token.DEFINE:
			inner.param(l.info.Defs[id], id.Name, pt)
		default:
			p := inner.newLocal("$p", pt)
			assigns = append(assigns, &sAssign{lhs: []lval{inner.lvalue(x, false)}, rhs: []expr{&eLocal{t: pt, v: p}}})
		}
	}
	inner.body = append(inner.body, assigns...)
	inner.stmts(s.Body.List)
	inner.emit(&sReturn{vals: []expr{&eConst{t: bt, n: 1}}})
	fn := inner.finish(l.ty(yield))

	l.emit(&sExpr{x: &eCallValue{t: tuple(), f: iter, args: []expr{fn}}})

	// a return in the body
	r := l.realFunc()
	if r.rr != nil {
		rr := l.rangeVarOf(r.rr)
		var then []stmt
		if l.rng != nil {
			then = []stmt{&sReturn{vals: []expr{&eConst{t: bt}}}}
		} else {
			var vals []expr
			for _, v := range l.rangeResults() {
				vals = append(vals, varExpr(v))
			}
			then = []stmt{&sReturn{vals: vals}}
		}
		l.emit(&sIf{cond: varExpr(rr), then: then})
	}
	// the branches to labels outside of the body
	i32 := l.ty(types.Typ[types.Int32])
	for _, b := range inner.rng.pending {
		rb := l.rangeVarOf(r.rb)
		cond := &eBinary{t: bt, op: opEqI, x: varExpr(rb), y: &eConst{t: i32, n: uint64(b.code)}}
		saved := l.body
		l.body = nil
		l.emit(&sAssign{lhs: []lval{varLval(rb)}, rhs: []expr{&eConst{t: i32}}})
		l.branchTo(b.tok, b.label, s.Pos())
		then := l.body
		l.body = saved
		l.emit(&sIf{cond: cond, then: then})
	}
}

// rangeVarOf returns the variable of a cell of the real function in l.
func (l *fnLower) rangeVarOf(obj *types.Var) *local {
	v, _ := l.lookup(obj)
	return v
}

// branchTo emits a break or continue to the label or to the innermost loop or switch, also out of the body of a
// range over an iterator.
func (l *fnLower) branchTo(tok token.Token, label string, pos token.Pos) {
	if label != "" {
		if t, ok := l.labels[label]; ok {
			l.emitBranch(tok, t)
			return
		}
	} else if tok == token.BREAK && len(l.breaks) > 0 {
		l.emit(&sBreak{target: l.breaks[len(l.breaks)-1]})
		return
	} else if tok == token.CONTINUE && len(l.conts) > 0 {
		l.emit(&sContinue{target: l.conts[len(l.conts)-1]})
		return
	}
	if l.rng == nil {
		l.fail(pos, "no target for %s %s", tok, label)
	}
	bt := l.c.boolType()
	if label == "" || label == l.rng.label {
		l.emit(&sReturn{vals: []expr{&eConst{t: bt, n: fromBool(tok == token.CONTINUE)}}})
		return
	}
	// a label of an enclosing function: the code tells it after the iteration
	l.c.branches++
	code := l.c.branches
	l.rng.pending = append(l.rng.pending, branchCode{tok: tok, label: label, code: code})
	i32 := l.ty(types.Typ[types.Int32])
	rb := l.rangeVar("$rb", i32, &l.realFunc().rb)
	l.emit(&sAssign{lhs: []lval{varLval(rb)}, rhs: []expr{&eConst{t: i32, n: uint64(code)}}})
	l.emit(&sReturn{vals: []expr{&eConst{t: bt}}})
}

func (l *fnLower) emitBranch(tok token.Token, target stmt) {
	if tok == token.BREAK {
		l.emit(&sBreak{target: target})
		return
	}
	l.emit(&sContinue{target: target.(*sLoop)})
}

// rangeReturn lowers a return in the body of a range: it stores the results and stops the iteration.
func (l *fnLower) rangeReturn(s *ast.ReturnStmt) {
	r := l.realFunc()
	bt := l.c.boolType()
	cells := l.rangeResults()
	var lhs []lval
	for _, v := range cells {
		lhs = append(lhs, varLval(v))
	}
	switch {
	case len(s.Results) == 0 && len(r.named) > 0:
		var vals []expr
		for _, obj := range r.namedObjs {
			nv, _ := l.lookup(obj)
			vals = append(vals, varExpr(nv))
		}
		l.emit(&sAssign{lhs: lhs, rhs: vals})
	case len(s.Results) == 1 && len(cells) > 1:
		l.emit(&sAssign{lhs: lhs, rhs: []expr{l.expr(s.Results[0])}})
	case len(s.Results) > 0:
		var vals []expr
		for i, x := range s.Results {
			vals = append(vals, l.val(x, cells[i].cell))
		}
		l.emit(&sAssign{lhs: lhs, rhs: vals})
	}
	rr := l.rangeVar("$rr", bt, &r.rr)
	l.emit(&sAssign{lhs: []lval{varLval(rr)}, rhs: []expr{&eConst{t: bt, n: 1}}})
	l.emit(&sReturn{vals: []expr{&eConst{t: bt}}})
}
