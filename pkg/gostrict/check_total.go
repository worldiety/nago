// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"
)

// The rules of this file make a checked program total: it can not fail at runtime. There is no nil except for
// error, an integer division proves its divisor and every function of the library is defined for all inputs.

var errorType = types.Universe.Lookup("error").Type()

func isError(t types.Type) bool {
	return t != nil && types.Identical(t, errorType)
}

// nilable tells, whether the zero value of t contains nil, apart from an error, which may be nil.
func nilable(t types.Type) bool {
	return nilableSeen(t, map[types.Type]bool{})
}

func nilableSeen(t types.Type, seen map[types.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	if isOption(t) {
		// the zero value holds nothing
		return false
	}
	if isOpaque(t) {
		return true
	}
	if tp, ok := types.Unalias(t).(*types.TypeParam); ok {
		return typeParamNilable(tp, seen)
	}
	switch u := t.Underlying().(type) {
	case *types.Pointer, *types.Signature:
		return true
	case *types.Interface:
		return !isError(t)
	case *types.Struct:
		for f := range u.Fields() {
			if nilableSeen(f.Type(), seen) {
				return true
			}
		}
	}
	return false
}

func isFloat(t types.Type) bool {
	return basicKind(t) == types.Float64
}

func isInteger(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsInteger != 0
}

func (c *checker) isLocal(obj types.Object) bool {
	v, ok := obj.(*types.Var)
	return ok && v.Pkg() != nil && v.Parent() != v.Pkg().Scope()
}

func (c *checker) checkShift(n ast.Node, count ast.Expr) {
	tv, ok := c.info.Types[count]
	if ok && tv.Value != nil {
		if v, exact := constant.Int64Val(constant.ToInt(tv.Value)); exact && v >= 0 && v < 32 {
			return
		}
	}
	c.report(n, CodeBits, "the shift count must be a constant from 0 to 31", "Go and JavaScript treat counts from 32 on and negative counts differently", "shift by a constant, e.g. x << 3, or use a loop for variable counts")
}

func (c *checker) checkDivision(n ast.Node, x, y ast.Expr) {
	if !isInteger(c.info.TypeOf(x)) {
		return
	}
	tv, ok := c.info.Types[y]
	if ok && tv.Value != nil {
		// a constant zero is already a type error
		return
	}

	const why = "an integer division by zero fails at runtime in Go"
	if c.positiveMax(y) {
		return
	}
	id, ok := ast.Unparen(y).(*ast.Ident)
	obj := c.info.Uses[id]
	if !ok || !c.isLocal(obj) {
		c.report(y, CodeDivision, "the divisor may be zero", why, "assign the divisor to a local variable d, check it with if d == 0 { return ... } and divide by d")
		return
	}

	_, from, guarded := c.guardFor(c.nonZero(obj))
	switch {
	case !guarded:
		c.report(y, CodeDivision, "the divisor "+id.Name+" may be zero", why, "check it before the division, e.g. if "+id.Name+" == 0 { return ... } or if "+id.Name+" != 0 { ... }")
	case !c.stable(obj, from):
		c.report(y, CodeDivision, "the divisor "+id.Name+" may be zero, because it is assigned after the check", why, "copy it into a new variable after the last assignment and check that one")
	}
}

// headerCall returns the function of a header package, which the call calls, or nil.
func (c *checker) headerCall(call *ast.CallExpr) *types.Func {
	var fn *types.Func
	switch f := ast.Unparen(call.Fun).(type) {
	case *ast.SelectorExpr:
		if sel := c.info.Selections[f]; sel != nil {
			fn, _ = sel.Obj().(*types.Func)
		} else {
			fn, _ = c.info.Uses[f.Sel].(*types.Func)
		}
	case *ast.Ident:
		fn, _ = c.info.Uses[f].(*types.Func)
	case *ast.IndexExpr:
		return c.headerCall(&ast.CallExpr{Fun: f.X})
	case *ast.IndexListExpr:
		return c.headerCall(&ast.CallExpr{Fun: f.X})
	}
	if fn == nil || fn.Pkg() == nil {
		return nil
	}
	if p := c.prog.Package(fn.Pkg().Path()); p != nil && p.Kind == KindHeader {
		return fn
	}
	return nil
}

// maybeInvalid tells, whether the call returns (T, bool), where T is a zero value with nil, if the bool is false.
func (c *checker) maybeInvalid(x ast.Expr) bool {
	call, ok := ast.Unparen(x).(*ast.CallExpr)
	if !ok || c.headerCall(call) == nil {
		return false
	}
	t, ok := c.info.TypeOf(call).(*types.Tuple)
	if !ok || t.Len() != 2 || basicKind(t.At(1).Type()) != types.Bool {
		return false
	}
	return nilable(t.At(0).Type())
}

// collectPairs finds v, ok := l.Get(i) and similar, whose v may contain nil, unless ok is true.
func (c *checker) collectPairs(f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		var lhs []ast.Expr
		var rhs []ast.Expr
		define := true
		switch n := n.(type) {
		case *ast.AssignStmt:
			lhs, rhs, define = n.Lhs, n.Rhs, n.Tok == token.DEFINE
		case *ast.ValueSpec:
			for _, id := range n.Names {
				lhs = append(lhs, id)
			}
			rhs = n.Values
		default:
			return true
		}
		if len(lhs) != 2 || len(rhs) != 1 || !c.maybeInvalid(rhs[0]) {
			return true
		}

		// = may assign a field or an element, := and var only new variables
		v, isVar := lhs[0].(*ast.Ident)
		ok, isOk := lhs[1].(*ast.Ident)
		switch {
		case !define || !isVar || !isOk:
			fix := "write v, ok := ... and use v only where ok is true"
			if !isVar {
				fix = "write v, ok := ... and assign v inside if ok { ... }, e.g. if v, ok := l.Get(i); ok { r.F = v }"
			}
			c.report(n, CodeGuard, "assign the result with := to new variables", "the value may be nil, unless ok is true, and the compiler tracks this only for new variables", fix)
		case v.Name == "_":
		case ok.Name == "_":
			c.report(ok, CodeGuard, "ok must be checked", "the value is nil, if ok is false", "write v, ok := ... and use v inside if ok { ... } or after if !ok { return ... }")
		default:
			okObj := c.info.Defs[ok]
			if okObj == nil {
				okObj = c.info.Uses[ok]
			}
			c.pairs[c.info.Defs[v]] = okObj
		}
		return true
	})
}

// checkIdent checks the uses of values, which may be nil, and the nil literal.
func (c *checker) checkIdent(id *ast.Ident) {
	obj := c.info.Uses[id]
	if obj == nil {
		return
	}

	if _, ok := obj.(*types.Nil); ok {
		c.checkNil(id)
		return
	}

	okObj, paired := c.pairs[obj]
	if !paired {
		return
	}
	if len(c.assigns[obj]) > 0 {
		c.report(id, CodeGuard, id.Name+" is assigned again, so the compiler can not prove, that it is not nil", "the value may be nil, unless ok is true", "use a new variable for the new value")
		return
	}
	if _, _, ok := c.guardFor(c.truthy(okObj)); !ok || !c.stable(okObj, obj.Pos()) {
		c.report(id, CodeGuard, id.Name+" may be nil, because "+okObj.Name()+" is not checked", "the value is the zero value, if "+okObj.Name()+" is false", "use "+id.Name+" inside if "+okObj.Name()+" { ... } or after if !"+okObj.Name()+" { return ... }")
	}
}

func (c *checker) checkNil(id *ast.Ident) {
	if len(c.stack) >= 2 {
		if b, ok := c.stack[len(c.stack)-2].(*ast.BinaryExpr); ok && (b.Op == token.EQL || b.Op == token.NEQ) {
			other := b.X
			if other == ast.Expr(id) {
				other = b.Y
			}
			if t := c.info.TypeOf(other); !isError(t) {
				c.report(b, CodeNil, "values of "+t.String()+" are never nil", "the subset has no nil except for errors", "remove the comparison")
			}
			return
		}
	}

	if t := c.nilTarget(id); !isError(t) {
		c.report(id, CodeNil, "nil is only allowed for errors", "a checked program never dereferences nil, so only an error may be nil", "create a value, e.g. &T{...} or list.New[T](), or use a bool result to tell, that there is no value")
	}
}

// checkErrorCall proves, that the error, whose method is called, is not nil.
func (c *checker) checkErrorCall(call *ast.CallExpr) {
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return
	}
	sel := c.info.Selections[se]
	if sel == nil || sel.Kind() != types.MethodVal || !isError(sel.Recv()) {
		return
	}

	id, ok := ast.Unparen(se.X).(*ast.Ident)
	obj := c.info.Uses[id]
	if !ok || !c.isLocal(obj) {
		c.report(se.X, CodeGuard, "the error may be nil", "a method call on a nil error fails at runtime", "assign the error to a local variable err and use it inside if err != nil { ... }")
		return
	}
	_, from, guarded := c.guardFor(c.nonNil(obj))
	if !guarded || !c.stable(obj, from) {
		c.report(se.X, CodeGuard, id.Name+" may be nil", "a method call on a nil error fails at runtime", "call "+id.Name+"."+se.Sel.Name+"() inside if "+id.Name+" != nil { ... } and do not assign "+id.Name+" in between")
	}
}

// checkMake rejects list.Make for elements, whose zero value contains nil.
func (c *checker) checkMake(call *ast.CallExpr) {
	fn := c.headerCall(call)
	if fn == nil || fn.Pkg().Path() != listPath || fn.Name() != "Make" {
		return
	}
	if t, ok := c.info.TypeOf(call).(*types.Named); ok && t.TypeArgs().Len() == 1 && nilable(t.TypeArgs().At(0)) {
		c.report(call, CodeNil, "list.Make of "+t.TypeArgs().At(0).String()+" would contain nil", "the zero value of the element contains nil", "create the list with list.New and append the elements")
	}
}

// checkValueSpec rejects variables without a value, whose zero value contains nil.
func (c *checker) checkValueSpec(vs *ast.ValueSpec) {
	if len(vs.Values) > 0 {
		return
	}
	for _, n := range vs.Names {
		if obj := c.info.Defs[n]; obj != nil && nilable(obj.Type()) {
			c.report(n, CodeNil, n.Name+" has no value and would be nil", "the zero value of "+obj.Type().String()+" contains nil", "initialize it, e.g. "+n.Name+" := "+example(obj.Type()))
		}
	}
}

// checkNamedResults rejects named results, whose zero value contains nil.
func (c *checker) checkNamedResults(ft *ast.FuncType) {
	if ft.Results == nil {
		return
	}
	for _, f := range ft.Results.List {
		for _, n := range f.Names {
			if obj := c.info.Defs[n]; obj != nil && nilable(obj.Type()) {
				c.report(n, CodeNil, "the named result "+n.Name+" would be nil", "the zero value of "+obj.Type().String()+" contains nil", "remove the name and return the value explicitly")
			}
		}
	}
}

// checkMissingFields rejects composite literals, which leave fields with nil.
func (c *checker) checkMissingFields(lit *ast.CompositeLit) {
	t := c.info.TypeOf(lit)
	st, ok := t.Underlying().(*types.Struct)
	if !ok || isOpaque(t) {
		return
	}
	set := map[string]bool{}
	for i, el := range lit.Elts {
		if kv, ok := el.(*ast.KeyValueExpr); ok {
			set[kv.Key.(*ast.Ident).Name] = true
		} else if i < st.NumFields() {
			set[st.Field(i).Name()] = true
		}
	}

	var missing []string
	for f := range st.Fields() {
		if !set[f.Name()] && nilable(f.Type()) {
			missing = append(missing, f.Name())
		}
	}
	if len(missing) > 0 {
		c.report(lit, CodeNil, "the fields "+strings.Join(missing, ", ")+" would be nil", "their zero value contains nil", "set them in the literal, e.g. "+missing[0]+": list.New[T]()")
	}
}

// typeParamNilable tells, whether a type argument of tp may contain nil. A constraint without types, e.g. any,
// allows any type, a union only its terms.
func typeParamNilable(tp *types.TypeParam, seen map[types.Type]bool) bool {
	iface, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok || iface.IsMethodSet() {
		return true
	}
	for i := 0; i < iface.NumEmbeddeds(); i++ {
		switch e := iface.EmbeddedType(i).Underlying().(type) {
		case *types.Union:
			for j := 0; j < e.Len(); j++ {
				if nilableSeen(e.Term(j).Type(), seen) {
					return true
				}
			}
		case *types.Interface:
			// an embedded constraint, e.g. comparable or another union
			if e.IsMethodSet() && !e.IsComparable() {
				continue
			}
			if e.IsComparable() && e.NumEmbeddeds() == 0 {
				return true
			}
		default:
			if nilableSeen(iface.EmbeddedType(i), seen) {
				return true
			}
		}
	}
	return false
}

// nilTarget returns the type, which the nil literal on top of the stack is assigned to.
func (c *checker) nilTarget(id *ast.Ident) types.Type {
	if t := c.info.TypeOf(id); t != nil && !isUntypedNil(t) {
		return t
	}
	if len(c.stack) < 2 {
		return nil
	}

	index := func(list []ast.Expr) int {
		for i, e := range list {
			if ast.Unparen(e) == ast.Expr(id) {
				return i
			}
		}
		return -1
	}

	switch p := c.stack[len(c.stack)-2].(type) {
	case *ast.ReturnStmt:
		sig := c.enclosingSignature()
		if i := index(p.Results); sig != nil && i >= 0 && i < sig.Results().Len() {
			return sig.Results().At(i).Type()
		}
	case *ast.AssignStmt:
		if i := index(p.Rhs); i >= 0 && i < len(p.Lhs) {
			return c.info.TypeOf(p.Lhs[i])
		}
	case *ast.ValueSpec:
		if i := index(p.Values); i >= 0 && i < len(p.Names) {
			return c.info.TypeOf(p.Names[i])
		}
	case *ast.CallExpr:
		if sig, ok := c.info.TypeOf(p.Fun).Underlying().(*types.Signature); ok {
			if i := index(p.Args); i >= 0 && i < sig.Params().Len() {
				return sig.Params().At(i).Type()
			}
		}
	case *ast.KeyValueExpr:
		return c.info.TypeOf(p.Key)
	}
	return nil
}

func isUntypedNil(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Kind() == types.UntypedNil
}

func (c *checker) enclosingSignature() *types.Signature {
	for i := len(c.stack) - 1; i >= 0; i-- {
		switch f := c.stack[i].(type) {
		case *ast.FuncLit:
			sig, _ := c.info.TypeOf(f).(*types.Signature)
			return sig
		case *ast.FuncDecl:
			if fn, ok := c.info.Defs[f.Name].(*types.Func); ok {
				return fn.Signature()
			}
			return nil
		}
	}
	return nil
}

// example returns an expression, which creates a value of t, for a fix.
func example(t types.Type) string {
	switch {
	case isOpaque(t):
		n := types.Unalias(t).(*types.Named)
		if n.Obj().Name() == "List" {
			return "list.New[" + n.TypeArgs().At(0).String() + "]()"
		}
		return "dict.New[" + n.TypeArgs().At(0).String() + ", " + n.TypeArgs().At(1).String() + "]()"
	case isPointer(t):
		return "&" + types.TypeString(t.Underlying().(*types.Pointer).Elem(), func(p *types.Package) string { return p.Name() }) + "{...}"
	case isInterface(t):
		return "a value of a type, which implements " + t.String()
	}
	if _, ok := t.Underlying().(*types.Signature); ok {
		return "func(...) { ... }"
	}
	return types.TypeString(t, func(p *types.Package) string { return p.Name() }) + "{...} with all fields set"
}

// positiveMax tells, whether e is max(..., c, ...) with a positive constant c, which is never zero.
func (c *checker) positiveMax(e ast.Expr) bool {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	id, ok := ast.Unparen(call.Fun).(*ast.Ident)
	if !ok {
		return false
	}
	if b, ok := c.info.Uses[id].(*types.Builtin); !ok || b.Name() != "max" {
		return false
	}
	for _, a := range call.Args {
		if tv, ok := c.info.Types[a]; ok && tv.Value != nil && constant.Sign(tv.Value) > 0 {
			return true
		}
	}
	return false
}
