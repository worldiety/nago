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

// Interfaces: a value of an interface is a box, an object with the dynamic type and the slots of the value, e.g.
// the pointer of a pointer variant. Each type, which the program boxes, has a table of the methods of the
// interfaces, see [tinfo.methods]. A call of a method of an interface copies the receiver from the box in front of
// the arguments, because its slots depend on the variant. A Go error of errors.New is an error as it is.

// methodRef is a method in the table of a type: the function and whether it takes the value of a pointer.
type methodRef struct {
	fn *Func
	// deref is set, if the type is a pointer and the method takes the value, which it points to
	deref bool
}

// methodsFor adds the methods of the interface to the table of the type v, whose values the program boxes.
func (c *compiler) methodsFor(v *tinfo, iface types.Type, pos token.Pos) {
	it, ok := iface.Underlying().(*types.Interface)
	if !ok || v.t == nil {
		return
	}
	for m := range it.Methods() {
		if _, ok := v.methods[m.Name()]; ok {
			continue
		}
		obj, _, _ := types.LookupFieldOrMethod(v.t, true, m.Pkg(), m.Name())
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		if p, ok := c.pkgs[fn.Pkg().Path()]; !ok || !p.User {
			// a method of a handle, e.g. the marker of ui.View, which only the host calls
			continue
		}
		var recvArgs []types.Type
		if n, ok := types.Unalias(derefType(v.t)).(*types.Named); ok {
			for t := range n.TypeArgs().Types() {
				recvArgs = append(recvArgs, t)
			}
		}
		_, ptrRecv := fn.Signature().Recv().Type().(*types.Pointer)
		if v.methods == nil {
			v.methods = map[string]*methodRef{}
		}
		v.methods[m.Name()] = &methodRef{fn: c.funcFor(fn, recvArgs, nil), deref: v.kind == kPtr && !ptrRecv}
	}
}

// box boxes the value e of a concrete type into the interface to.
func (l *fnLower) box(pos token.Pos, e expr, to *tinfo) expr {
	from := e.typ()
	l.c.methodsFor(from, to.t, pos)
	return &eBox{t: to, x: e, v: from}
}

// ifaceCall calls the method of the interface value x.
func (l *fnLower) ifaceCall(se *ast.SelectorExpr, call *ast.CallExpr) expr {
	sig := l.sub(l.info.TypeOf(se)).(*types.Signature)
	return &eCallIface{t: l.tupleOf(sig.Results()), x: l.expr(se.X), method: se.Sel.Name, args: l.args(call, sig, nil)}
}

// typeSwitch lowers a type switch over a sealed interface into a chain of comparisons of the dynamic type.
func (l *fnLower) typeSwitch(s *ast.TypeSwitchStmt, label string) {
	if s.Init != nil {
		l.stmt(s.Init, "")
	}
	var bind *ast.Ident
	var subject ast.Expr
	switch a := s.Assign.(type) {
	case *ast.AssignStmt:
		bind = a.Lhs[0].(*ast.Ident)
		subject = a.Rhs[0].(*ast.TypeAssertExpr).X
	case *ast.ExprStmt:
		subject = a.X.(*ast.TypeAssertExpr).X
	}
	x := l.expr(subject)
	xv := l.newLocal("$x", x.typ())
	l.emit(&sAssign{lhs: []lval{&lLocal{t: x.typ(), v: xv}}, rhs: []expr{x}})
	val := &eLocal{t: x.typ(), v: xv}

	block := &sBlock{}
	if label != "" {
		l.labels[label] = block
	}
	l.breaks = append(l.breaks, block)
	defer func() { l.breaks = l.breaks[:len(l.breaks)-1] }()

	bt := l.c.boolType()
	var ifs []*sIf
	var dflt []stmt
	hasDefault := false
	for _, cs := range s.Body.List {
		cc := cs.(*ast.CaseClause)
		var cond expr
		var single *tinfo
		for _, te := range cc.List {
			var c expr
			if tv := l.info.Types[te]; tv.IsNil() {
				c = &eUnary{t: bt, op: opIsNil, x: val}
			} else {
				t := l.typeOf(te)
				c = &eTypeIs{t: bt, x: val, v: t}
				if len(cc.List) == 1 {
					single = t
				}
			}
			if cond == nil {
				cond = c
			} else {
				cond = &eAndOr{t: bt, x: cond, y: c}
			}
		}
		saved := l.body
		l.body = nil
		if bind != nil {
			if obj := l.info.Implicits[cc]; obj != nil {
				t := l.ty(obj.Type())
				var v expr = retype(val, t)
				if single != nil && t.kind != kIface && t.kind != kError {
					v = &eLoad{t: t, p: val}
				}
				nv := l.newVar(obj, bind.Name, t)
				l.emit(l.initVar(nv, v))
			}
		}
		l.stmts(cc.Body)
		body := l.body
		l.body = saved
		if cc.List == nil {
			dflt, hasDefault = body, true
			continue
		}
		ifs = append(ifs, &sIf{cond: cond, then: body})
	}
	for i := len(ifs) - 1; i >= 0; i-- {
		if i == len(ifs)-1 {
			if hasDefault {
				ifs[i].els = dflt
			}
		} else {
			ifs[i].els = []stmt{ifs[i+1]}
		}
	}
	switch {
	case len(ifs) > 0:
		block.body = []stmt{ifs[0]}
	case hasDefault:
		block.body = dflt
	}
	l.emit(block)
}

// addressed returns the variables, whose address a call of a method with a pointer receiver takes implicitly, e.g.
// b.Inc() or p.A.Inc() of variables b and p of struct types. They live in cells, which are their addresses.
func addressed(info *types.Info, body *ast.BlockStmt, res map[types.Object]bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		se, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		sel := info.Selections[se]
		if sel == nil || sel.Kind() != types.MethodVal {
			return true
		}
		fn := sel.Obj().(*types.Func)
		if _, ptr := fn.Signature().Recv().Type().(*types.Pointer); !ptr {
			return true
		}
		if _, ok := info.TypeOf(se.X).Underlying().(*types.Pointer); ok {
			return true
		}
		// the root of the fields, unless a pointer is on the way
		x := ast.Unparen(se.X)
		for {
			s, ok := x.(*ast.SelectorExpr)
			if !ok {
				break
			}
			if fs := info.Selections[s]; fs == nil || fs.Kind() != types.FieldVal || fs.Indirect() {
				return true
			}
			if _, ok := info.TypeOf(s.X).Underlying().(*types.Pointer); ok {
				return true
			}
			x = ast.Unparen(s.X)
		}
		if id, ok := x.(*ast.Ident); ok {
			if v, ok := info.Uses[id].(*types.Var); ok && v.Parent() != v.Pkg().Scope() {
				res[v] = true
			}
		}
		return true
	})
}

// addr returns the address of an addressable value, e.g. the receiver of a method with a pointer receiver: the cell
// of a variable, or a view into the object of a struct, whose field it is.
func (l *fnLower) addr(x ast.Expr) expr {
	ptrT := l.ty(types.NewPointer(l.info.TypeOf(x)))
	switch x := ast.Unparen(x).(type) {
	case *ast.Ident:
		if v, ok := l.lookup(l.info.Uses[x]); ok && v.cell != nil {
			return retype(&eLocal{t: v.t, v: v}, ptrT)
		}
	case *ast.SelectorExpr:
		if sel := l.info.Selections[x]; sel != nil && sel.Kind() == types.FieldVal {
			bt := l.typeOf(x.X)
			if bt.kind == kPtr {
				return l.view(l.expr(x.X), bt.fieldOf(x.Sel.Name).off, ptrT)
			}
			if bt.kind == kStruct {
				return l.view(l.addr(x.X), bt.fieldOf(x.Sel.Name).off, ptrT)
			}
		}
	case *ast.StarExpr:
		return l.expr(x.X)
	}
	// a value, which nothing else refers to, e.g. a package variable, which never changes
	return &eNew{t: ptrT, x: l.expr(x)}
}

func (l *fnLower) view(p expr, off layout, ptrT *tinfo) expr {
	if v, ok := p.(*eView); ok {
		p, off = v.p, v.off.add(off)
	}
	if off == (layout{}) {
		return retype(p, ptrT)
	}
	return &eView{t: ptrT, p: p, off: off}
}

// methodValue returns a method bound to its receiver, e.g. f := a.Get, as a function literal, which captures the
// receiver.
func (l *fnLower) methodValue(se *ast.SelectorExpr, sel *types.Selection) expr {
	sig := l.sub(l.info.TypeOf(se)).(*types.Signature)
	fn := sel.Obj().(*types.Func)
	recvT := l.typeOf(se.X)
	var recv expr
	_, ptrRecv := fn.Signature().Recv().Type().(*types.Pointer)
	switch {
	case recvT.kind == kIface || recvT.kind == kError:
		recv = l.expr(se.X)
	case ptrRecv && recvT.kind != kPtr:
		recv = l.addr(se.X)
	case !ptrRecv && recvT.kind == kPtr:
		recv = l.deref(l.expr(se.X))
	default:
		recv = l.expr(se.X)
	}
	// the receiver is evaluated with the method value
	rv := l.newLocal("$recv", recv.typ())
	l.emit(&sAssign{lhs: []lval{&lLocal{t: rv.t, v: rv}}, rhs: []expr{recv}})

	inner := l.child(sig, "bound")
	var args []expr
	for i, pt := range inner.f.params.elems {
		p := inner.newLocal("p", pt)
		_ = i
		args = append(args, &eLocal{t: pt, v: p})
	}
	cp := inner.newLocal("$recv", rv.t)
	inner.caps = append(inner.caps, capture{src: rv, dst: cp})
	res := inner.f.results
	var call expr
	if recvT.kind == kIface || recvT.kind == kError {
		call = &eCallIface{t: res, x: &eLocal{t: cp.t, v: cp}, method: se.Sel.Name, args: args}
	} else {
		var recvArgs []types.Type
		if n, ok := types.Unalias(derefType(recvT.t)).(*types.Named); ok {
			for t := range n.TypeArgs().Types() {
				recvArgs = append(recvArgs, t)
			}
		}
		call = &eCall{t: res, fn: l.c.funcFor(fn, recvArgs, nil), args: append([]expr{&eLocal{t: cp.t, v: cp}}, args...)}
	}
	inner.returnCall(call)
	return inner.finish(l.typeOf(se))
}

// deref returns the value, which the pointer p points to.
func (l *fnLower) deref(p expr) expr {
	t := p.typ().elem
	if t.kind == kShared {
		// the pointer refers to the shared object itself
		return &eField{t: t, x: p}
	}
	return &eLoad{t: t, p: p}
}
