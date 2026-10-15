// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/token"
	"go/types"
)

// A struct is passed by value, so a call copies it. The copy is not needed, if the function only reads the
// parameter and the caller passes a variable, which nothing else can change during the call. The translated code
// then passes the struct itself.

// readonlyInfo tells, which parameters and whether the receiver of a function are only read: the function does not
// assign them or their fields, does not call a method with a pointer receiver on them and does not refer to them
// in a function literal, which could outlive the call. Wherever such a parameter is assigned to another variable,
// stored or returned, the translated code copies it anyway.
type readonlyInfo struct {
	recv   bool
	params []bool
}

func (e *emitter) readonly(fn *types.Func) readonlyInfo {
	fn = fn.Origin()
	if e.readonlyCache == nil {
		e.readonlyCache = map[*types.Func]readonlyInfo{}
	}
	if info, ok := e.readonlyCache[fn]; ok {
		return info
	}
	info := readonlyInfo{}
	e.readonlyCache[fn] = info

	src, ok := e.decls[fn]
	if !ok || src.pkg.Kind != KindUser || src.decl.Body == nil {
		return info
	}
	defs := src.pkg.Info.Defs
	objs := map[types.Object]bool{}
	var recv types.Object
	if src.decl.Recv != nil {
		for _, f := range src.decl.Recv.List {
			for _, n := range f.Names {
				recv = defs[n]
				objs[recv] = true
			}
		}
	}
	var params []types.Object
	for _, f := range src.decl.Type.Params.List {
		if len(f.Names) == 0 {
			params = append(params, nil)
		}
		for _, n := range f.Names {
			params = append(params, defs[n])
			objs[defs[n]] = true
		}
	}

	written := writtenObjects(src.pkg.Info, src.decl.Body, objs)
	if recv != nil {
		info.recv = !written[recv]
	} else {
		// a receiver without a name is never used
		info.recv = src.decl.Recv != nil
	}
	for _, p := range params {
		info.params = append(info.params, p == nil || !written[p])
	}
	e.readonlyCache[fn] = info
	return info
}

// writtenObjects returns the objects of objs, which the body assigns, calls a method with a pointer receiver on or
// refers to in a function literal.
func writtenObjects(info *types.Info, body *ast.BlockStmt, objs map[types.Object]bool) map[types.Object]bool {
	res := map[types.Object]bool{}
	mark := func(x ast.Expr) {
		if obj := rootObject(info, x); obj != nil && objs[obj] {
			res[obj] = true
		}
	}

	var visit func(n ast.Node, inLit bool) bool
	visit = func(n ast.Node, inLit bool) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			ast.Inspect(n.Body, func(m ast.Node) bool { return visit(m, true) })
			return false
		case *ast.Ident:
			if obj := info.Uses[n]; inLit && obj != nil && objs[obj] {
				res[obj] = true
			}
		case *ast.AssignStmt:
			for _, l := range n.Lhs {
				mark(l)
			}
		case *ast.IncDecStmt:
			mark(n.X)
		case *ast.RangeStmt:
			if n.Tok == token.ASSIGN {
				if n.Key != nil {
					mark(n.Key)
				}
				if n.Value != nil {
					mark(n.Value)
				}
			}
		case *ast.CallExpr:
			if se, ok := ast.Unparen(n.Fun).(*ast.SelectorExpr); ok {
				if sel := info.Selections[se]; sel != nil && sel.Kind() == types.MethodVal {
					if fn, ok := sel.Obj().(*types.Func); ok {
						if _, ptr := fn.Origin().Signature().Recv().Type().(*types.Pointer); ptr {
							mark(se.X)
						}
					}
				}
			}
		}
		return true
	}
	ast.Inspect(body, func(n ast.Node) bool { return visit(n, false) })
	return res
}

// rootObject returns the variable, whose value or field x is, e.g. p of p.A.B.
func rootObject(info *types.Info, x ast.Expr) types.Object {
	for {
		switch e := ast.Unparen(x).(type) {
		case *ast.Ident:
			return info.Uses[e]
		case *ast.SelectorExpr:
			x = e.X
		case *ast.StarExpr:
			x = e.X
		default:
			return nil
		}
	}
}

// unaliased tells, whether nothing can change the struct x during a call: it is a local variable or a field of a
// local struct, which no function literal of the function refers to, because a variable has no address in the
// subset. A field of a pointer may change through another pointer.
func (c *fnCtx) unaliased(x ast.Expr) bool {
	if c.captured == nil {
		return false
	}
	for {
		switch e := ast.Unparen(x).(type) {
		case *ast.Ident:
			v, ok := c.info.Uses[e].(*types.Var)
			if !ok || v.Parent() == nil || v.Pkg() == nil || v.Parent() == v.Pkg().Scope() {
				return false
			}
			return !c.captured[v]
		case *ast.SelectorExpr:
			sel := c.info.Selections[e]
			if sel == nil || sel.Kind() != types.FieldVal || !isStruct(c.typeOf(e.X)) {
				return false
			}
			x = e.X
		default:
			return false
		}
	}
}

// capturedObjects returns the objects, which a function literal in the body refers to.
func capturedObjects(info *types.Info, body *ast.BlockStmt) map[types.Object]bool {
	res := map[types.Object]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		lit, ok := n.(*ast.FuncLit)
		if !ok {
			return true
		}
		ast.Inspect(lit.Body, func(m ast.Node) bool {
			if id, ok := m.(*ast.Ident); ok {
				if obj := info.Uses[id]; obj != nil {
					res[obj] = true
				}
			}
			return true
		})
		return false
	})
	return res
}

// argVal translates an argument for the parameter i of fn, which needs no copy, if fn only reads it.
func (c *fnCtx) argVal(fn *types.Func, i int, x ast.Expr, to types.Type) string {
	if fn != nil && isStruct(c.typeOf(x)) && !isInterface(c.sub(to)) && c.unaliased(x) {
		if ro := c.e.readonly(fn); i < len(ro.params) && ro.params[i] {
			return c.expr(x)
		}
	}
	return c.val(x, to)
}
