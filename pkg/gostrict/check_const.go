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
	"slices"
)

// checkConstArgs checks, that the arguments of the parameters, which the constDirective of the called function
// names, are constants, e.g. a literal, a named constant or an expression of them.
func (c *checker) checkConstArgs(n *ast.CallExpr) {
	fn := staticCallee(c.info, n)
	if fn == nil {
		return
	}
	fn = fn.Origin()
	src := c.decls[fn]
	if len(src.consts) == 0 {
		return
	}
	params := fn.Signature().Params()
	for i := range params.Len() {
		p := params.At(i)
		if !slices.Contains(src.consts, p.Name()) || !constType(p.Type()) {
			continue
		}
		// a call of a function with several results passes them all as one argument
		var arg ast.Expr = n
		if len(n.Args) == params.Len() {
			arg = n.Args[i]
			if tv, ok := c.info.Types[arg]; ok && tv.Value != nil {
				continue
			}
		}
		name := fn.Pkg().Name() + "." + fn.Name()
		c.report(arg, CodeConstArg, "the argument "+p.Name()+" of "+name+" must be a constant", "the host reads it from the source at compile time, without running the program, see "+constDirective+" of "+name, "write a literal or a constant, e.g. "+name+"("+constExample(params, i)+")")
	}
}

// constExample returns the arguments of a call, whose i-th argument is a literal of its type.
func constExample(params *types.Tuple, i int) string {
	res := ""
	for j := range params.Len() {
		if j > 0 {
			res += ", "
		}
		if j != i {
			res += "…"
			continue
		}
		b, _ := params.At(j).Type().Underlying().(*types.Basic)
		switch {
		case b == nil:
			res += "…"
		case b.Info()&types.IsString != 0:
			res += `"…"`
		case b.Info()&types.IsBoolean != 0:
			res += "true"
		default:
			res += "1"
		}
	}
	return res
}

// checkConstValue reports a function with the constDirective, which is not called but used as a value, e.g.
// f := process.Activity, because a call of f would bypass the check of its arguments.
func (c *checker) checkConstValue(n *ast.SelectorExpr) {
	fn, _ := c.info.Uses[n.Sel].(*types.Func)
	if fn == nil || len(c.decls[fn.Origin()].consts) == 0 || c.isCalled(n) {
		return
	}
	name := fn.Pkg().Name() + "." + fn.Name()
	c.report(n, CodeConstValue, name+" can only be called", "its arguments for "+constDirective+" must be constants at each call, which a call of a function value would bypass", "call "+name+"(…) directly, or wrap the call in a function literal with constant arguments")
}

// isCalled tells, whether the expression on top of the stack, maybe instantiated or in parentheses, is the function
// of a call.
func (c *checker) isCalled(n ast.Expr) bool {
	var child ast.Node = n
	for i := len(c.stack) - 2; i >= 0; i-- {
		switch p := c.stack[i].(type) {
		case *ast.ParenExpr:
		case *ast.IndexExpr:
			if p.X != child {
				return false
			}
		case *ast.IndexListExpr:
			if p.X != child {
				return false
			}
		case *ast.CallExpr:
			return p.Fun == child
		default:
			return false
		}
		child = c.stack[i]
	}
	return false
}

// checkConstDirective reports the constDirective on a function of the program, because only the functions of
// header packages are read by the host.
func (c *checker) checkConstDirective(n *ast.FuncDecl) {
	if _, pos := constNames(n); pos != token.NoPos {
		c.d.report(pos, token.NoPos, CodeHeader, constDirective+" is only allowed in header packages", "the host reads the constant arguments of the functions, which its header packages declare", "remove the directive")
	}
}
