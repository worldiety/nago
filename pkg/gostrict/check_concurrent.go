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
	"strconv"
)

const concurrentWhy = "the function runs on another goroutine, while the program runs further, so anything, which it shares with the program, would be a data race"

const concurrentFix = "pass only values, which can not change, and keep the rest in the database or in the result: copy what the function needs into a new variable, e.g. id := in.ID"

// checkConcurrentArgs checks the function arguments of a call of a function with //gostrict:calls concurrent. Such an
// argument runs as a call of its own in another instance, see [Detached], so it must not share memory with the program:
// it is a function literal, whose captured variables never change and hold only values, which can not change, or a
// function of a package of the program, which captures nothing. Everything else is denied.
func (c *checker) checkConcurrentArgs(n *ast.CallExpr) {
	fn := staticCallee(c.info, n)
	if fn == nil || c.decls[fn.Origin()].calls != CallsConcurrent {
		return
	}
	sig := fn.Origin().Signature()
	if len(n.Args) != sig.Params().Len() {
		// the results of one call as all arguments: none of them is a literal or a named function
		c.report(n, CodeConcurrent, "the function arguments of "+fn.Pkg().Name()+"."+fn.Name()+" must be function literals or functions of a package", concurrentWhy, concurrentFix)
		return
	}
	for _, i := range funcParams(sig) {
		c.checkConcurrentArg(fn, ast.Unparen(n.Args[i]))
	}
}

func (c *checker) checkConcurrentArg(callee *types.Func, arg ast.Expr) {
	name := callee.Pkg().Name() + "." + callee.Name()
	if lit, ok := arg.(*ast.FuncLit); ok {
		c.checkCaptures(name, lit)
		return
	}
	if fn, iface, ok := funcRef(c.info, arg); ok && !iface && fn.Signature().Recv() == nil && fn.Pkg() != nil {
		if p := c.prog.Package(fn.Pkg().Path()); p != nil && p.Kind == KindUser {
			return
		}
	}
	c.report(arg, CodeConcurrent, "the function argument of "+name+" must be a function literal or a function of a package", concurrentWhy+", and only then the compiler sees, what it refers to", "write the function literal directly into the call or name a function of the package")
}

// checkCaptures reports the variables, which the literal captures and which may share memory with the program.
func (c *checker) checkCaptures(callee string, lit *ast.FuncLit) {
	seen := map[types.Object]bool{}
	ast.Inspect(lit.Body, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		v, ok := c.info.Uses[id].(*types.Var)
		if !ok || seen[v] || v.IsField() || v.Pkg() == nil || v.Parent() == v.Pkg().Scope() || (v.Pos() >= lit.Pos() && v.Pos() < lit.End()) {
			return true
		}
		seen[v] = true
		if msg := concurrentType(v.Type(), map[types.Type]bool{}); msg != "" {
			c.report(id, CodeConcurrent, "the function literal of "+callee+" captures "+v.Name()+", which holds "+msg, concurrentWhy, concurrentFix)
			return true
		}
		if c.isNamedResult(v) {
			c.report(id, CodeConcurrent, "the function literal of "+callee+" captures the named result "+v.Name(), concurrentWhy+", and a return assigns it", "copy the value into a new variable, before the function literal captures it")
			return true
		}
		if pos, ok := c.concurrentWrites()[v]; ok {
			c.report(id, CodeConcurrent, "the function literal of "+callee+" captures "+v.Name()+", which is assigned in line "+strconv.Itoa(c.d.fset.Position(pos).Line), concurrentWhy, "copy the value into a new variable, which nothing assigns, before the function literal captures it")
		}
		return true
	})
}

// isNamedResult tells, whether v is a named result of a function around the node on top of the stack.
func (c *checker) isNamedResult(v *types.Var) bool {
	for _, n := range c.stack {
		var results *ast.FieldList
		switch n := n.(type) {
		case *ast.FuncDecl:
			results = n.Type.Results
		case *ast.FuncLit:
			results = n.Type.Results
		}
		if results == nil {
			continue
		}
		for _, f := range results.List {
			for _, name := range f.Names {
				if c.info.Defs[name] == types.Object(v) {
					return true
				}
			}
		}
	}
	return false
}

// concurrentWrites returns the first position, which writes a local variable or a field of it after its declaration,
// by the variables of the package: an assignment, ++ or --, the range clause of an assignment or a method with a
// pointer receiver, which changes the variable.
func (c *checker) concurrentWrites() map[types.Object]token.Pos {
	if c.writes != nil {
		return c.writes
	}
	c.writes = map[types.Object]token.Pos{}
	mark := func(x ast.Expr) {
		if x == nil {
			return
		}
		obj := rootObject(c.info, x)
		if obj == nil {
			return
		}
		if _, ok := c.writes[obj]; !ok {
			c.writes[obj] = x.Pos()
		}
	}
	for _, f := range c.pkg.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.AssignStmt:
				for _, l := range n.Lhs {
					mark(l)
				}
			case *ast.IncDecStmt:
				mark(n.X)
			case *ast.RangeStmt:
				if n.Tok == token.ASSIGN {
					mark(n.Key)
					mark(n.Value)
				}
			case *ast.CallExpr:
				if se, ok := ast.Unparen(n.Fun).(*ast.SelectorExpr); ok {
					if sel := c.info.Selections[se]; sel != nil && sel.Kind() == types.MethodVal {
						if m, ok := sel.Obj().(*types.Func); ok {
							if _, ptr := m.Origin().Signature().Recv().Type().(*types.Pointer); ptr {
								mark(se.X)
							}
						}
					}
				}
			}
			return true
		})
	}
	return c.writes
}

// concurrentType describes, why a value of t may share memory with the program, or returns "". Only a whitelist of
// types passes: basic values and their enums, big.Int, bin.Bytes, decimal.Dec and structs, opt.Option and list.Frozen of them.
func concurrentType(t types.Type, seen map[types.Type]bool) string {
	if _, ok := t.(*types.TypeParam); ok {
		return "a value of a type parameter, which may be anything"
	}
	if isValueHandle(t) {
		return ""
	}
	if seen[t] {
		return ""
	}
	seen[t] = true

	switch {
	case isError(t):
		return "an error, which may be a pointer"
	case isHandle(t), isHandleUnion(t):
		return "a handle of the host, which belongs to the instance of the program"
	case isOpaque(t):
		n := types.Unalias(t).(*types.Named)
		if n.Obj().Name() == "Frozen" || isOption(t) {
			return concurrentType(n.TypeArgs().At(0), seen)
		}
		return "a " + n.Obj().Pkg().Name() + "." + n.Obj().Name() + ", which can change"
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		return ""
	case *types.Struct:
		for f := range u.Fields() {
			if msg := concurrentType(f.Type(), seen); msg != "" {
				return msg + " in the field " + f.Name()
			}
		}
		return ""
	case *types.Signature:
		return "a function, which may capture anything"
	case *types.Pointer:
		return "a pointer, through which its value can change"
	case *types.Interface:
		return "an interface, whose value may be a pointer"
	}
	return "a value, which can change"
}

// checkConcurrentValue reports a function with //gostrict:calls concurrent, which is not called but used as a value,
// because a call of the value would bypass the check of its function arguments. A function with the constDirective
// is already reported with GS044.
func (c *checker) checkConcurrentValue(n *ast.SelectorExpr) {
	fn, _ := c.info.Uses[n.Sel].(*types.Func)
	if fn == nil || c.decls[fn.Origin()].calls != CallsConcurrent || len(c.decls[fn.Origin()].consts) > 0 || c.isCalled(n) {
		return
	}
	name := fn.Pkg().Name() + "." + fn.Name()
	c.report(n, CodeConcurrent, name+" can only be called", concurrentWhy+", and only at a direct call the compiler sees its function arguments", "call "+name+"(…) directly")
}
