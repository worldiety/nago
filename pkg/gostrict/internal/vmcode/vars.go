// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// Variables and closures: a function literal captures the variables of the enclosing functions, which it uses. A
// variable, which is captured and assigned after its declaration, lives in a cell, an object, which the function
// and its literals share; any other captured variable is copied, when the literal is evaluated. A cell is created by
// the declaration of its variable, so that each iteration of a loop has its own variable, like in Go 1.22.

// cellsOf returns the variables of a function, which need a cell: those, which a function literal or the body of a
// range over an iterator uses and which are assigned after their declaration.
func cellsOf(info *types.Info, body *ast.BlockStmt) map[types.Object]bool {
	captured := map[types.Object]bool{}
	assigned := map[types.Object]bool{}
	var inner func(n ast.Node, from, to token.Pos)
	// inner marks the variables, which n uses and which are declared outside of [from, to)
	inner = func(n ast.Node, from, to token.Pos) {
		ast.Inspect(n, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				if v, ok := info.Uses[id].(*types.Var); ok && !v.IsField() && v.Parent() != v.Pkg().Scope() && (v.Pos() < from || v.Pos() >= to) {
					captured[v] = true
				}
			}
			return true
		})
	}
	target := func(x ast.Expr) {
		if id, ok := ast.Unparen(x).(*ast.Ident); ok {
			if obj := info.Uses[id]; obj != nil {
				assigned[obj] = true
			}
		}
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			inner(n.Body, n.Pos(), n.End())
		case *ast.RangeStmt:
			if rangeOverFunc(info, n) {
				inner(n.Body, n.Pos(), n.End())
			}
			if n.Tok == token.ASSIGN {
				target(n.Key)
				if n.Value != nil {
					target(n.Value)
				}
			}
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				// := assigns the variables, which it does not declare
				target(lhs)
			}
		case *ast.IncDecStmt:
			target(n.X)
		}
		return true
	})
	cells := map[types.Object]bool{}
	for v := range captured {
		if assigned[v] {
			cells[v] = true
		}
	}
	return cells
}

// rangeOverFunc tells, whether the range calls an iterator with its body, unlike a range over the methods of
// lists, maps and outboxes, which the compiler turns into loops.
func rangeOverFunc(info *types.Info, s *ast.RangeStmt) bool {
	if _, ok := info.TypeOf(s.X).Underlying().(*types.Signature); !ok {
		return false
	}
	if call, ok := ast.Unparen(s.X).(*ast.CallExpr); ok {
		if se, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sel := info.Selections[se]; sel != nil && sel.Kind() == types.MethodVal {
				if n, ok := types.Unalias(derefType(sel.Recv())).(*types.Named); ok && n.Obj().Pkg() != nil {
					switch n.Obj().Pkg().Path() + "." + n.Obj().Name() {
					case listPath + ".List", listPath + ".Frozen", msgPath + ".Out", dictPath + ".Map":
						return false
					}
				}
			}
		}
	}
	return true
}

// cellOf returns the type of a pointer to a cell of a variable of type t, whose object holds the slots of t.
func (c *compiler) cellOf(t *tinfo) *tinfo {
	key := "$cell " + t.name + fmt.Sprintf(" %p", t)
	if ti, ok := c.types[key]; ok {
		return ti
	}
	cell := &tinfo{kind: kStruct, l: t.l, obj: t.l, name: "$cell", fields: []field{{name: "v", t: t}}}
	ti := &tinfo{kind: kPtr, l: layout{r: 1}, elem: cell, obj: t.l, fields: cell.fields, name: "*$cell"}
	c.types[key] = ti
	return ti
}

// newVar declares a variable of the object, in a cell, if it needs one.
func (l *fnLower) newVar(obj types.Object, name string, t *tinfo) *local {
	if l.cells[obj] {
		v := l.newLocal(name, l.c.cellOf(t))
		v.cell = t
		if obj != nil {
			l.vars[obj] = v
		}
		return v
	}
	v := l.newLocal(name, t)
	if obj != nil {
		l.vars[obj] = v
	}
	return v
}

// initVar returns the statement, which initializes a new variable: a variable in a cell gets a new cell.
func (l *fnLower) initVar(v *local, val expr) stmt {
	if v.cell != nil {
		return &sAssign{lhs: []lval{&lLocal{t: v.t, v: v}}, rhs: []expr{&eNew{t: v.t, x: val}}}
	}
	return &sAssign{lhs: []lval{&lLocal{t: v.t, v: v}}, rhs: []expr{val}}
}

// varExpr returns the value of a variable.
func varExpr(v *local) expr {
	if v.cell != nil {
		return &eLoad{t: v.cell, p: &eLocal{t: v.t, v: v}}
	}
	return &eLocal{t: v.t, v: v}
}

// varLval returns the target of an assignment to a variable.
func varLval(v *local) lval {
	if v.cell != nil {
		return &lStore{t: v.cell, p: &eLocal{t: v.t, v: v}}
	}
	return &lLocal{t: v.t, v: v}
}

// capture is a variable of an enclosing function, which a function literal copies into its frame, when it is
// called: the value of the variable or the pointer to its cell.
type capture struct {
	src, dst *local
}

// lookup returns the variable of the object, which may be captured from the enclosing functions.
func (l *fnLower) lookup(obj types.Object) (*local, bool) {
	if v, ok := l.vars[obj]; ok {
		return v, true
	}
	if l.parent == nil {
		return nil, false
	}
	src, ok := l.parent.lookup(obj)
	if !ok {
		return nil, false
	}
	dst := l.newLocal(src.name, src.t)
	dst.cell = src.cell
	l.caps = append(l.caps, capture{src: src, dst: dst})
	l.vars[obj] = dst
	return dst, true
}

// funcLit lowers a function literal into a function of its own, whose value captures the variables.
func (l *fnLower) funcLit(lit *ast.FuncLit) expr {
	sig := l.sub(l.info.TypeOf(lit)).(*types.Signature)
	inner := l.child(sig, "func")
	names := fieldNames(lit.Type.Params)
	for i, id := range names {
		inner.param(l.info.Defs[id], id.Name, inner.f.params.elems[i])
	}
	inner.namedResults(lit.Type.Results)
	inner.stmts(lit.Body.List)
	return inner.finish(l.typeOf(lit))
}

// child returns the lowering of a function, which the function of l encloses, e.g. a function literal or the
// body of a range over an iterator, which captures the variables of l.
func (l *fnLower) child(sig *types.Signature, kind string) *fnLower {
	l.c.lits++
	f := &Func{name: fmt.Sprintf("%s.%s%d", l.f.name, kind, l.c.lits), params: l.paramTuple(nil, sig), results: l.tupleOf(sig.Results())}
	l.c.out.funcs = append(l.c.out.funcs, f)
	inner := l.c.newFnLower(f, l.info)
	inner.parent, inner.subst, inner.cells, inner.pkg = l, l.subst, l.cells, l.pkg
	return inner
}

// finish completes a function literal and returns its value, which captures the variables.
func (inner *fnLower) finish(t *tinfo) expr {
	f := inner.f
	f.body = append(inner.prologue, inner.body...)
	f.locals = inner.frame
	var caps []expr
	for _, cp := range inner.caps {
		caps = append(caps, &eLocal{t: cp.src.t, v: cp.src})
		f.caps = append(f.caps, capSlot{at: cp.dst.at, l: cp.dst.t.l})
	}
	if len(caps) == 0 {
		f.closure = &Closure{fn: f}
		return &eFunc{t: t, fn: f}
	}
	return &eClosure{t: t, fn: f, caps: caps}
}

// param binds a parameter, which the caller writes into the next slots of the frame: a parameter in a cell is
// copied into a new cell.
func (l *fnLower) param(obj types.Object, name string, t *tinfo) {
	raw := l.newLocal(name, t)
	if obj == nil || name == "_" {
		return
	}
	if !l.cells[obj] {
		l.vars[obj] = raw
		return
	}
	v := l.newVar(obj, name, t)
	l.emit(l.initVar(v, &eLocal{t: t, v: raw}))
}

// namedResults declares the named results with their zero values, which a bare return returns.
func (l *fnLower) namedResults(fl *ast.FieldList) {
	if fl == nil {
		return
	}
	for _, fld := range fl.List {
		for _, id := range fld.Names {
			obj := l.info.Defs[id]
			t := l.ty(obj.Type())
			v := l.newVar(obj, id.Name, t)
			l.named = append(l.named, v)
			l.namedObjs = append(l.namedObjs, obj)
			l.emit(l.initVar(v, l.c.zeroOrNil(t)))
		}
	}
}

// paramTuple returns the parameters of a function as a tuple, the receiver first.
func (l *fnLower) paramTuple(recv *types.Var, sig *types.Signature) *tinfo {
	var params []*tinfo
	if recv != nil {
		params = append(params, l.ty(recv.Type()))
	}
	for v := range sig.Params().Variables() {
		params = append(params, l.ty(v.Type()))
	}
	return &tinfo{kind: kTuple, elems: params, offs: tupleOffsets(params), l: tupleLayout(params)}
}

func tupleOffsets(ts []*tinfo) []layout {
	var offs []layout
	var at layout
	for _, t := range ts {
		offs = append(offs, at)
		at = at.add(t.l)
	}
	return offs
}

func tupleLayout(ts []*tinfo) layout {
	var at layout
	for _, t := range ts {
		at = at.add(t.l)
	}
	return at
}

// sub substitutes the type arguments of the instance, which l lowers.
func (l *fnLower) sub(t types.Type) types.Type {
	if len(l.subst) == 0 || t == nil {
		return t
	}
	return l.c.substitute(t, l.subst)
}

// ty returns the description of a type of the source of l.
func (l *fnLower) ty(t types.Type) *tinfo {
	return l.c.typeOf(l.sub(t))
}

func (l *fnLower) tupleOf(res *types.Tuple) *tinfo {
	return l.c.tupleOf(l.c.substituteTuple(res, l.subst))
}

func (c *compiler) substitute(t types.Type, m map[*types.TypeParam]types.Type) types.Type {
	if len(m) == 0 {
		return t
	}
	switch t := types.Unalias(t).(type) {
	case *types.TypeParam:
		if r, ok := m[t]; ok {
			return r
		}
		return t
	case *types.Pointer:
		return types.NewPointer(c.substitute(t.Elem(), m))
	case *types.Named:
		if t.TypeArgs().Len() == 0 {
			return t
		}
		var args []types.Type
		changed := false
		for a := range t.TypeArgs().Types() {
			r := c.substitute(a, m)
			changed = changed || r != a
			args = append(args, r)
		}
		if !changed {
			return t
		}
		inst, err := types.Instantiate(c.tctx, t.Origin(), args, false)
		if err != nil {
			c.fail(t.Obj().Pos(), "cannot instantiate %s: %v", t, err)
		}
		return inst
	case *types.Signature:
		return types.NewSignatureType(nil, nil, nil, c.substituteTuple(t.Params(), m), c.substituteTuple(t.Results(), m), t.Variadic())
	case *types.Tuple:
		return c.substituteTuple(t, m)
	case *types.Slice:
		return types.NewSlice(c.substitute(t.Elem(), m))
	}
	return t
}

func (c *compiler) substituteTuple(t *types.Tuple, m map[*types.TypeParam]types.Type) *types.Tuple {
	if t == nil || len(m) == 0 {
		return t
	}
	var vars []*types.Var
	for v := range t.Variables() {
		vars = append(vars, types.NewParam(v.Pos(), v.Pkg(), v.Name(), c.substitute(v.Type(), m)))
	}
	return types.NewTuple(vars...)
}

// declare declares a variable, which an assignment initializes: a variable in a cell is assigned as a plain
// variable first and moved into its cell by boxCells, after all values are assigned.
func (l *fnLower) declare(obj types.Object, name string, t *tinfo) (*local, lval) {
	if !l.cells[obj] {
		v := l.newVar(obj, name, t)
		return v, &lLocal{t: t, v: v}
	}
	raw := l.newLocal(name, t)
	v := l.newVar(obj, name, t)
	v.raw = raw
	return v, &lLocal{t: t, v: raw}
}

// boxCells moves the values of declared variables into their new cells.
func (l *fnLower) boxCells(cells []*local) {
	for _, v := range cells {
		l.emit(l.initVar(v, &eLocal{t: v.raw.t, v: v.raw}))
	}
}
