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

// A package variable is bound once: it is initialized, when the program is loaded, and never changes. The state of
// a program, which changes, belongs into a value, which an exported function returns and the host passes back,
// see [ValueState]. So the instances of a bundle are interchangeable, a call depends only on its arguments, a new
// version of a program can take over the state of the old one, and an interpreter could share the program
// between parallel calls.

const globalWhy = "a package variable is initialized once and never changes, so that a call depends only on its arguments and the instances of a program are interchangeable"

const globalFix = "keep the changing values in a state, e.g. type State struct{ ... }, return a *State from an exported function and pass it to the functions, which change it"

// checkGlobals checks the types of the package variables of the file.
func (c *checker) checkGlobals(f *ast.File) {
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.VAR {
			continue
		}
		for _, spec := range g.Specs {
			for _, n := range spec.(*ast.ValueSpec).Names {
				v, ok := c.info.Defs[n].(*types.Var)
				if !ok || n.Name == "_" {
					continue
				}
				if msg := mutableType(v.Type()); msg != "" {
					c.report(n, CodeGlobal, "the package variable "+n.Name+" holds "+msg, globalWhy, "use list.Freeze(...) for a table of values, or "+globalFix)
				}
			}
		}
	}
}

// mutableType describes, why values of t can change, or returns "".
func mutableType(t types.Type) string {
	return mutableTypeSeen(t, map[types.Type]bool{})
}

func mutableTypeSeen(t types.Type, seen map[types.Type]bool) string {
	if seen[t] {
		return ""
	}
	seen[t] = true

	switch {
	case isHandle(t), isError(t):
		return ""
	case isOpaque(t):
		n := types.Unalias(t).(*types.Named)
		if n.Obj().Name() == "Frozen" {
			return mutableTypeSeen(n.TypeArgs().At(0), seen)
		}
		return "a " + n.Obj().Pkg().Name() + "." + n.Obj().Name() + ", which can change"
	}

	switch u := t.Underlying().(type) {
	case *types.Basic, *types.Signature:
		return ""
	case *types.Pointer:
		return "a pointer, through which its value can change"
	case *types.Interface:
		return "an interface, whose value may be a pointer"
	case *types.Struct:
		for f := range u.Fields() {
			if msg := mutableTypeSeen(f.Type(), seen); msg != "" {
				return msg + " in the field " + f.Name()
			}
		}
		return ""
	}
	return "a value, which can change"
}

// globalRoot returns the package variable, which the expression refers to, directly or through the fields of its
// struct value, or nil.
func (c *checker) globalRoot(e ast.Expr) *types.Var {
	switch e := ast.Unparen(e).(type) {
	case *ast.Ident:
		if v, ok := c.info.Uses[e].(*types.Var); ok && v.Pkg() != nil && v.Parent() == v.Pkg().Scope() {
			return v
		}
	case *ast.SelectorExpr:
		if sel := c.info.Selections[e]; sel != nil {
			if sel.Kind() == types.FieldVal && !isPointer(c.info.TypeOf(e.X)) {
				return c.globalRoot(e.X)
			}
			return nil
		}
		// a qualified package variable
		if v, ok := c.info.Uses[e.Sel].(*types.Var); ok && v.Pkg() != nil && v.Parent() == v.Pkg().Scope() {
			return v
		}
	}
	return nil
}

func (c *checker) checkGlobalWrite(lhs ast.Expr) {
	if v := c.globalRoot(lhs); v != nil {
		c.report(lhs, CodeGlobal, "the package variable "+v.Name()+" can not change after its initialization", globalWhy, globalFix)
	}
}

// checkGlobalMethod rejects a method with a pointer receiver on a package variable, which could change it.
func (c *checker) checkGlobalMethod(call *ast.CallExpr) {
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return
	}
	sel := c.info.Selections[se]
	if sel == nil || sel.Kind() != types.MethodVal || isPointer(sel.Recv()) {
		return
	}
	fn, ok := sel.Obj().(*types.Func)
	if !ok || !isPointer(fn.Signature().Recv().Type()) {
		return
	}
	if v := c.globalRoot(se.X); v != nil {
		c.report(se, CodeGlobal, "the method "+se.Sel.Name+" could change the package variable "+v.Name(), globalWhy, globalFix)
	}
}
