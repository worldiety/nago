// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"
)

func jsString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSpace(buf.String())
}

func (c *fnCtx) isNil(x ast.Expr) bool {
	tv, ok := c.info.Types[x]
	return ok && tv.IsNil()
}

// val translates x as a value of type to: a struct read from a variable is copied and a concrete value assigned
// to an interface is boxed with the methods of its type.
func (c *fnCtx) val(x ast.Expr, to types.Type) string {
	if c.isNil(x) {
		return "null"
	}
	from := c.typeOf(x)
	js := c.expr(x)
	if isStruct(from) && c.addressable(x) {
		js = c.e.clone(js, from)
	}
	if to != nil && isInterface(c.sub(to)) && !isInterface(from) {
		js = "$box(" + js + ", " + c.e.methodTable(from) + ")"
	}
	return js
}

// addressable tells, whether x refers to an existing value, which must be copied, when it is assigned.
func (c *fnCtx) addressable(x ast.Expr) bool {
	switch x := ast.Unparen(x).(type) {
	case *ast.Ident:
		_, ok := c.info.Uses[x].(*types.Var)
		return ok
	case *ast.SelectorExpr:
		if sel := c.info.Selections[x]; sel != nil {
			return sel.Kind() == types.FieldVal
		}
		_, ok := c.info.Uses[x.Sel].(*types.Var)
		return ok
	case *ast.StarExpr:
		return true
	}
	return false
}

func (c *fnCtx) expr(x ast.Expr) string {
	if tv, ok := c.info.Types[x]; ok && tv.Value != nil {
		return constJS(tv.Value, c.sub(tv.Type))
	}

	switch x := x.(type) {
	case *ast.ParenExpr:
		return "(" + c.expr(x.X) + ")"
	case *ast.Ident:
		return c.ident(x)
	case *ast.FuncLit:
		return c.funcLit(x)
	case *ast.CompositeLit:
		return c.composite(x)
	case *ast.SelectorExpr:
		return c.selector(x)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return c.instanceValue(x)
	case *ast.StarExpr:
		return c.expr(x.X)
	case *ast.UnaryExpr:
		return c.unary(x)
	case *ast.BinaryExpr:
		return c.binary(x)
	case *ast.CallExpr:
		return c.call(x)
	}
	fail(x.Pos(), "expression %T is not supported", x)
	return ""
}

func constJS(v constant.Value, t types.Type) string {
	switch v.Kind() {
	case constant.Bool:
		return strconv.FormatBool(constant.BoolVal(v))
	case constant.String:
		return jsString(constant.StringVal(v))
	case constant.Int:
		if i, ok := constant.Int64Val(v); ok {
			return wrapNeg(strconv.FormatInt(i, 10))
		}
		f, _ := constant.Float64Val(v)
		return wrapNeg(strconv.FormatFloat(f, 'g', -1, 64))
	case constant.Float:
		if b, ok := t.Underlying().(*types.Basic); ok && b.Info()&types.IsInteger != 0 {
			if i, ok := constant.Int64Val(constant.ToInt(v)); ok {
				return wrapNeg(strconv.FormatInt(i, 10))
			}
		}
		f, _ := constant.Float64Val(v)
		return wrapNeg(strconv.FormatFloat(f, 'g', -1, 64))
	}
	return "null"
}

func wrapNeg(s string) string {
	if strings.HasPrefix(s, "-") {
		return "(" + s + ")"
	}
	return s
}

func (c *fnCtx) ident(id *ast.Ident) string {
	obj := c.info.Uses[id]
	if obj == nil {
		obj = c.info.Defs[id]
	}
	switch obj := obj.(type) {
	case *types.Nil:
		return "null"
	case *types.Var:
		if obj.Pkg() != nil && obj.Parent() == obj.Pkg().Scope() {
			return c.e.global(obj)
		}
		return c.local(obj)
	case *types.Func:
		var targs []types.Type
		if inst, ok := c.info.Instances[id]; ok {
			targs = c.subList(typeArgs(inst.TypeArgs))
		}
		return c.funcRef(obj, targs, id.Pos())
	}
	fail(id.Pos(), "identifier %s is not supported", id.Name)
	return ""
}

func (c *fnCtx) subList(ts []types.Type) []types.Type {
	res := make([]types.Type, len(ts))
	for i, t := range ts {
		res[i] = c.sub(t)
	}
	return res
}

// funcRef returns a function value of a package level function.
func (c *fnCtx) funcRef(fn *types.Func, targs []types.Type, pos token.Pos) string {
	p := fn.Pkg().Path()
	if _, ok := stdAllowed[p]; ok || p == textPath || p == numPath {
		return "$" + fn.Pkg().Name() + "_" + fn.Name()
	}
	src, ok := c.e.decls[fn.Origin()]
	if ok && src.pkg.Kind == KindHeader && src.js != "" {
		// a function of the library with a binding, e.g. of std/big, is called like any other header function
		return c.e.externName(fn, targs, src, pos)
	}
	if strings.HasPrefix(p, stdPrefix) && (!ok || src.pkg.Kind != KindUser) {
		// the built in functions of the library, e.g. list.New, have no function value
		fail(pos, "%s can only be called", fn.FullName())
	}
	if ok && src.pkg.Kind == KindHeader {
		return c.e.externName(fn, targs, src, pos)
	}
	if _, entry := c.e.entries[fn]; entry && fn.Pkg() != c.pkg.Types {
		// another package calls the entry point of a whole program within the aspect, also through a function value
		return c.e.hook(fn)
	}
	return c.e.funcName(fn, nil, targs)
}

func (c *fnCtx) instanceValue(x ast.Expr) string {
	var base ast.Expr
	switch x := x.(type) {
	case *ast.IndexExpr:
		base = x.X
	case *ast.IndexListExpr:
		base = x.X
	}
	if id, ok := ast.Unparen(base).(*ast.Ident); ok {
		return c.ident(id)
	}
	fail(x.Pos(), "index expressions are not supported")
	return ""
}

func (c *fnCtx) funcLit(f *ast.FuncLit) string {
	sig := c.info.TypeOf(f).(*types.Signature)
	sub := c.nested(sig)
	params := sub.params(f.Type.Params)
	sub.namedResults(f.Type.Results)
	sub.stmts(f.Body.List)

	var sb strings.Builder
	sb.WriteString("((" + strings.Join(params, ", ") + ") => {\n")
	if sub.rangeReturn {
		sb.WriteString("\tlet $rr = false, $rv;\n")
	}
	if sub.rangeBranch {
		sb.WriteString("\tlet $rb = null;\n")
	}
	body := sub.b.String()
	// indent the body relative to the current line
	pad := strings.Repeat("\t", c.ind)
	for _, l := range strings.Split(strings.TrimRight(body, "\n"), "\n") {
		sb.WriteString(pad + l + "\n")
	}
	sb.WriteString(pad + "})")
	return sb.String()
}

func (c *fnCtx) composite(lit *ast.CompositeLit) string {
	t := c.typeOf(lit)
	st, ok := t.Underlying().(*types.Struct)
	if !ok || isOpaque(t) {
		fail(lit.Pos(), "composite literal of %s is not supported", t)
	}
	c.e.structName(t)

	// JavaScript evaluates the properties in their order, so the values follow the source like in Go, and the
	// omitted fields come last
	var order []int
	vals := make([]string, st.NumFields())
	for i, el := range lit.Elts {
		if kv, ok := el.(*ast.KeyValueExpr); ok {
			name := kv.Key.(*ast.Ident).Name
			for j := 0; j < st.NumFields(); j++ {
				if st.Field(j).Name() == name {
					vals[j] = c.val(kv.Value, st.Field(j).Type())
					order = append(order, j)
				}
			}
			continue
		}
		vals[i] = c.val(el, st.Field(i).Type())
		order = append(order, i)
	}
	for i := range vals {
		if !slices.Contains(order, i) {
			vals[i] = c.e.zero(st.Field(i).Type())
			order = append(order, i)
		}
	}
	return objectLitOrder(st, vals, order)
}

// objectLit returns the object of a struct with the values of its fields. An object literal is faster than a
// class, e.g. in goja, and in a browser the objects of a type share their shape anyway.
func objectLit(st *types.Struct, vals []string) string {
	order := make([]int, len(vals))
	for i := range order {
		order[i] = i
	}
	return objectLitOrder(st, vals, order)
}

// objectLitOrder returns the object with the fields in the order, in which their values are evaluated.
func objectLitOrder(st *types.Struct, vals []string, order []int) string {
	parts := make([]string, len(order))
	for k, i := range order {
		parts[k] = fieldName(st.Field(i), i) + ": " + vals[i]
	}
	return "({ " + strings.Join(parts, ", ") + " })"
}

func (c *fnCtx) fieldName(se *ast.SelectorExpr) string {
	sel := c.info.Selections[se]
	recv := c.sub(sel.Recv())
	if p, ok := recv.Underlying().(*types.Pointer); ok {
		recv = p.Elem()
	}
	st := recv.Underlying().(*types.Struct)
	idx := sel.Index()[len(sel.Index())-1]
	return fieldName(st.Field(idx), idx)
}

func (c *fnCtx) selector(se *ast.SelectorExpr) string {
	sel := c.info.Selections[se]
	if sel == nil {
		// a qualified identifier
		switch obj := c.info.Uses[se.Sel].(type) {
		case *types.Var:
			return c.e.global(obj)
		case *types.Func:
			var targs []types.Type
			if inst, ok := c.info.Instances[se.Sel]; ok {
				targs = c.subList(typeArgs(inst.TypeArgs))
			}
			return c.funcRef(obj, targs, se.Pos())
		}
		fail(se.Pos(), "selector %s is not supported", se.Sel.Name)
	}

	switch sel.Kind() {
	case types.FieldVal:
		return c.expr(se.X) + "." + c.fieldName(se)
	case types.MethodVal:
		// a method value binds the receiver, when it is evaluated
		r := c.tmp()
		return "((" + r + ") => (...a) => " + c.methodCallJS(se, r, true, "...a") + ")(" + c.receiver(se) + ")"
	}
	fail(se.Pos(), "method expressions are not supported")
	return ""
}

type method struct {
	named    *types.Named
	fn       *types.Func
	ptrRecv  bool
	iface    bool
	recvType types.Type
}

func (c *fnCtx) resolveMethod(se *ast.SelectorExpr) method {
	recvT := c.typeOf(se.X)
	if isInterface(recvT) {
		return method{iface: true, recvType: recvT}
	}

	base := recvT
	if p, ok := recvT.Underlying().(*types.Pointer); ok {
		base = p.Elem()
	}
	named, ok := types.Unalias(base).(*types.Named)
	if !ok {
		fail(se.Pos(), "method of %s is not supported", recvT)
	}

	obj, _, _ := types.LookupFieldOrMethod(recvT, true, c.info.Selections[se].Obj().Pkg(), se.Sel.Name)
	fn, ok := obj.(*types.Func)
	if !ok {
		fail(se.Pos(), "method %s of %s not found", se.Sel.Name, recvT)
	}
	_, ptrRecv := fn.Origin().Signature().Recv().Type().(*types.Pointer)
	return method{named: named, fn: fn, ptrRecv: ptrRecv, recvType: recvT}
}

// receiver translates the receiver of a method call. A value receiver gets a copy, a pointer receiver the value
// itself, which is the address of a struct in the translated code.
func (c *fnCtx) receiver(se *ast.SelectorExpr) string {
	m := c.resolveMethod(se)
	js := c.expr(se.X)
	if m.iface || m.ptrRecv || isOpaque(m.named) {
		return js
	}
	if isStruct(m.named) && (isPointer(m.recvType) || c.addressable(se.X)) {
		return c.e.clone(js, m.named)
	}
	return js
}

func (c *fnCtx) methodTypeArgs(se *ast.SelectorExpr) []types.Type {
	if inst, ok := c.info.Instances[se.Sel]; ok {
		return c.subList(typeArgs(inst.TypeArgs))
	}
	return nil
}

// methodCallJS calls the method of se on the receiver recv with the translated arguments.
func (c *fnCtx) methodCallJS(se *ast.SelectorExpr, recv string, _ bool, args string) string {
	m := c.resolveMethod(se)
	if m.iface {
		if args != "" {
			args = ", " + args
		}
		return "$ic(" + recv + ", " + jsString(se.Sel.Name) + args + ")"
	}
	if isOpaque(m.named) {
		fail(se.Pos(), "%s can only be called", se.Sel.Name)
	}
	if src, ok := c.e.decls[m.fn.Origin()]; ok && src.pkg.Kind == KindHeader {
		// a method of a header package is a function of the host or a binding, which takes the receiver first
		if args != "" {
			args = ", " + args
		}
		return c.e.externName(m.fn.Origin(), nil, src, se.Pos()) + "(" + recv + args + ")"
	}
	name := c.e.funcName(m.fn.Origin(), typeArgs(m.named.TypeArgs()), c.methodTypeArgs(se))
	if args != "" {
		args = ", " + args
	}
	return name + "(" + recv + args + ")"
}

func (c *fnCtx) unary(u *ast.UnaryExpr) string {
	x := c.expr(u.X)
	switch u.Op {
	case token.AND:
		return x
	case token.NOT:
		return "!" + x
	case token.ADD:
		return x
	case token.SUB:
		switch basicKind(c.typeOf(u.X)) {
		case types.Int32:
			return "((0 - " + x + ") | 0)"
		case types.Uint32:
			return "((0 - " + x + ") >>> 0)"
		default:
			return "(-" + x + ")"
		}
	case token.XOR:
		if basicKind(c.typeOf(u.X)) == types.Uint32 {
			return "(~" + x + " >>> 0)"
		}
		return "(~" + x + ")"
	}
	fail(u.Pos(), "operator %s is not supported", u.Op)
	return ""
}

func basicKind(t types.Type) types.BasicKind {
	b, ok := t.Underlying().(*types.Basic)
	if !ok {
		return types.Invalid
	}
	switch b.Kind() {
	case types.UntypedInt, types.UntypedRune:
		return types.Int
	case types.UntypedFloat:
		return types.Float64
	case types.UntypedString:
		return types.String
	case types.UntypedBool:
		return types.Bool
	}
	return b.Kind()
}

func (c *fnCtx) binary(b *ast.BinaryExpr) string {
	switch b.Op {
	case token.LAND:
		return "(" + c.expr(b.X) + " && " + c.expr(b.Y) + ")"
	case token.LOR:
		return "(" + c.expr(b.X) + " || " + c.expr(b.Y) + ")"
	case token.EQL, token.NEQ:
		neg := ""
		if b.Op == token.NEQ {
			neg = "!"
		}
		switch {
		case c.isNil(b.X):
			return neg + "(" + c.expr(b.Y) + " === null)"
		case c.isNil(b.Y):
			return neg + "(" + c.expr(b.X) + " === null)"
		}
		// a variant is compared with a sealed interface as an interface value
		t := c.typeOf(b.X)
		if isInterface(c.typeOf(b.Y)) {
			t = c.typeOf(b.Y)
		}
		return neg + c.e.eq(c.val(b.X, t), c.val(b.Y, t), t)
	}
	return c.binaryJS(b.Op, c.expr(b.X), c.expr(b.Y), c.typeOf(b.X), c.typeOf(b.Y), b.Pos())
}

// binaryJS applies the arithmetic or comparison of Go to the translated operands.
func (c *fnCtx) binaryJS(op token.Token, x, y string, xt, yt types.Type, pos token.Pos) string {
	k := basicKind(xt)
	switch op {
	case token.LSS, token.LEQ, token.GTR, token.GEQ:
		if k == types.String {
			return "($scmp(" + x + ", " + y + ") " + op.String() + " 0)"
		}
		return "(" + x + " " + op.String() + " " + y + ")"
	case token.SHL, token.SHR:
		// the checker allows only constant shift counts from 0 to 31, which JavaScript shifts like Go
		switch {
		case k == types.Int32 && op == token.SHL:
			return "(" + x + " << " + y + ")"
		case k == types.Int32:
			return "(" + x + " >> " + y + ")"
		case k == types.Uint32 && op == token.SHL:
			return "((" + x + " << " + y + ") >>> 0)"
		case k == types.Uint32:
			return "(" + x + " >>> " + y + ")"
		}
		fail(pos, "shift of %s is not supported", xt)
	}

	switch k {
	case types.String:
		if op == token.ADD {
			return "(" + x + " + " + y + ")"
		}
	case types.Float64:
		switch op {
		case token.ADD, token.SUB, token.MUL, token.QUO:
			return "(" + x + " " + op.String() + " " + y + ")"
		}
	case types.Int32, types.Uint32:
		conv := " | 0)"
		if k == types.Uint32 {
			conv = " >>> 0)"
		}
		switch op {
		case token.ADD, token.SUB, token.AND, token.OR, token.XOR:
			return "((" + x + " " + op.String() + " " + y + ")" + conv
		case token.AND_NOT:
			return "((" + x + " & ~" + y + ")" + conv
		case token.MUL:
			return "($M.imul(" + x + ", " + y + ")" + conv
		case token.QUO, token.REM:
			// the checker proves, that the divisor is not zero
			return "((" + x + " " + op.String() + " " + y + ")" + conv
		}
	}
	fail(pos, "operator %s on %s is not supported", op, xt)
	return ""
}

func (c *fnCtx) call(call *ast.CallExpr) string {
	if tv, ok := c.info.Types[call.Fun]; ok && tv.IsType() {
		return c.conversion(call, c.sub(tv.Type))
	}

	fun := ast.Unparen(call.Fun)
	if id, ok := fun.(*ast.Ident); ok {
		if b, ok := c.info.Uses[id].(*types.Builtin); ok {
			return c.builtin(b, call)
		}
	}

	sig, _ := c.typeOf(call.Fun).Underlying().(*types.Signature)
	if sig == nil {
		fail(call.Pos(), "call of %s is not supported", c.typeOf(call.Fun))
	}

	// strip the instantiation, the type arguments are recorded at the identifier
	base := fun
	switch x := fun.(type) {
	case *ast.IndexExpr:
		base = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		base = ast.Unparen(x.X)
	}

	switch x := base.(type) {
	case *ast.SelectorExpr:
		if sel := c.info.Selections[x]; sel != nil && sel.Kind() == types.MethodVal {
			return c.methodCall(x, call, sig)
		}
		if sel := c.info.Selections[x]; sel == nil {
			if fn, ok := c.info.Uses[x.Sel].(*types.Func); ok {
				return c.funcCall(fn, x.Sel, call, sig)
			}
		}
	case *ast.Ident:
		if fn, ok := c.info.Uses[x].(*types.Func); ok {
			return c.funcCall(fn, x, call, sig)
		}
	}

	return c.expr(call.Fun) + "(" + c.args(call, sig) + ")"
}

// args translates the arguments for the parameters of sig. The variadic arguments of a function of a header
// package are passed as an array.
func (c *fnCtx) args(call *ast.CallExpr, sig *types.Signature) string {
	return c.argsFor(nil, call, sig)
}

// argsFor translates the arguments of a call of fn, which needs no copies of the structs, which it only reads.
func (c *fnCtx) argsFor(fn *types.Func, call *ast.CallExpr, sig *types.Signature) string {
	params := sig.Params()
	if len(call.Args) == 1 && params.Len() > 1 {
		// f(g()) with several results of g
		return "..." + c.expr(call.Args[0])
	}

	var res, rest []string
	for i, a := range call.Args {
		if sig.Variadic() && i >= params.Len()-1 {
			elem := params.At(params.Len() - 1).Type().(*types.Slice).Elem()
			rest = append(rest, c.val(a, elem))
			continue
		}
		res = append(res, c.argVal(fn, i, a, params.At(i).Type()))
	}
	if sig.Variadic() {
		res = append(res, "["+strings.Join(rest, ", ")+"]")
	}
	return strings.Join(res, ", ")
}

func (c *fnCtx) funcCall(fn *types.Func, id *ast.Ident, call *ast.CallExpr, sig *types.Signature) string {
	p := fn.Pkg().Path()
	var targs []types.Type
	if inst, ok := c.info.Instances[id]; ok {
		targs = c.subList(typeArgs(inst.TypeArgs))
	}

	switch p {
	case listPath:
		return c.listFunc(fn.Name(), targs, call, sig)
	case msgPath:
		if fn.Name() == "New" {
			return "[]"
		}
	case dictPath:
		if fn.Name() == "New" {
			return "$dnew()"
		}
	case optPath:
		switch fn.Name() {
		case "Some":
			return "{ v: " + c.val(call.Args[0], targs[0]) + " }"
		case "None":
			return "null"
		}
	}

	if src, ok := c.e.decls[fn.Origin()]; ok && src.pkg.Kind == KindHeader && (!strings.HasPrefix(p, stdPrefix) || src.js != "") {
		return c.e.externName(fn, targs, src, id.Pos()) + "(" + c.args(call, sig) + ")"
	}

	return c.funcRef(fn, targs, id.Pos()) + "(" + c.argsFor(fn, call, sig) + ")"
}

func (c *fnCtx) listFunc(name string, targs []types.Type, call *ast.CallExpr, sig *types.Signature) string {
	switch name {
	case "New":
		return "[]"
	case "Of":
		var vals []string
		for _, a := range call.Args {
			vals = append(vals, c.val(a, targs[0]))
		}
		return "[" + strings.Join(vals, ", ") + "]"
	case "Make":
		return "$lmake(" + c.expr(call.Args[0]) + ", " + c.e.zeroFn(targs[0]) + ")"
	case "SortFunc":
		return "$lsort(" + c.args(call, sig) + ")"
	case "Freeze":
		return "$lclone(" + c.expr(call.Args[0]) + ", " + c.e.cloneFn(targs[0]) + ")"
	}
	fail(call.Pos(), "list.%s is not supported", name)
	return ""
}

func (c *fnCtx) methodCall(se *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature) string {
	m := c.resolveMethod(se)
	if !m.iface && isOpaque(m.named) {
		return c.opaqueMethod(m.named, se, call, sig)
	}
	recv := ""
	if !m.iface && isStruct(m.named) && !isPointer(m.recvType) && !m.ptrRecv && c.unaliased(se.X) && c.e.readonly(m.fn).recv {
		// the method only reads its receiver, which the call does not need to copy, unlike a method value
		recv = c.expr(se.X)
	} else {
		recv = c.receiver(se)
	}
	return c.methodCallJS(se, recv, false, c.argsFor(m.fn, call, sig))
}

func (c *fnCtx) opaqueMethod(named *types.Named, se *ast.SelectorExpr, call *ast.CallExpr, sig *types.Signature) string {
	r := c.expr(se.X)
	args := typeArgs(named.TypeArgs())
	arg := func(i int, t types.Type) string { return c.val(call.Args[i], t) }

	method := named.Obj().Name() + "." + se.Sel.Name
	if rest, ok := strings.CutPrefix(method, "Frozen."); ok {
		// a frozen list is an array like a list, which the program can only read
		method = "List." + rest
	}
	if rest, ok := strings.CutPrefix(method, "Out."); ok {
		// an outbox is an array like a list, to which the program only appends
		method = "List." + strings.Replace(rest, "Send", "Append", 1)
	}
	switch method {
	case "List.Len":
		return r + ".length"
	case "List.Get":
		return "$lget(" + r + ", " + arg(0, nil) + ", " + c.e.zeroFn(args[0]) + ", " + c.e.cloneFn(args[0]) + ")"
	case "List.Set":
		return "$lset(" + r + ", " + arg(0, nil) + ", " + arg(1, args[0]) + ")"
	case "List.Append":
		var vals []string
		for i := range call.Args {
			vals = append(vals, arg(i, args[0]))
		}
		return "$lappend(" + r + ", [" + strings.Join(vals, ", ") + "])"
	case "List.Pop":
		return "$lpop(" + r + ", " + c.e.zeroFn(args[0]) + ")"
	case "List.Delete":
		return "$ldelete(" + r + ", " + arg(0, nil) + ")"
	case "List.Clone":
		return "$lclone(" + r + ", " + c.e.cloneFn(args[0]) + ")"
	case "List.All":
		return "$lall(" + r + ", " + c.e.cloneFn(args[0]) + ")"
	case "List.Values":
		return "$lvalues(" + r + ", " + c.e.cloneFn(args[0]) + ")"
	case "Map.Len":
		return r + ".size"
	case "Map.Get":
		return "$dget(" + r + ", " + arg(0, nil) + ", " + c.e.zeroFn(args[1]) + ", " + c.e.cloneFn(args[1]) + ")"
	case "Map.Has":
		return r + ".has(" + arg(0, nil) + ")"
	case "Map.Set":
		return "void " + r + ".set(" + arg(0, nil) + ", " + arg(1, args[1]) + ")"
	case "Map.Delete":
		return "void " + r + ".delete(" + arg(0, nil) + ")"
	case "Map.All":
		return "$dall(" + r + ", " + c.e.cloneFn(args[1]) + ")"
	case "Map.Keys":
		return "$dkeysSeq(" + r + ")"
	case "Option.Get":
		return "$oget(" + r + ", " + c.e.zeroFn(args[0]) + ", " + c.e.cloneFn(args[0]) + ")"
	case "Option.Or":
		return "$oor(" + r + ", " + arg(0, args[0]) + ", " + c.e.cloneFn(args[0]) + ")"
	case "Option.IsSome":
		return "(" + r + " !== null)"
	}
	fail(se.Pos(), "%s.%s is not supported", named.Obj().Name(), se.Sel.Name)
	return ""
}

func (c *fnCtx) builtin(b *types.Builtin, call *ast.CallExpr) string {
	switch b.Name() {
	case "min", "max":
		var args []string
		for _, a := range call.Args {
			args = append(args, c.expr(a))
		}
		return "$M." + b.Name() + "(" + strings.Join(args, ", ") + ")"
	}
	fail(call.Pos(), "builtin %s is not supported", b.Name())
	return ""
}

func (c *fnCtx) conversion(call *ast.CallExpr, to types.Type) string {
	a := call.Args[0]
	from := c.typeOf(a)
	if isInterface(to) {
		return c.val(a, to)
	}
	if isStruct(to) {
		return c.e.structName(to) + "$c(" + c.expr(a) + ")"
	}

	x := c.expr(a)
	fk, tk := basicKind(from), basicKind(to)
	switch {
	case tk == types.Int32 && fk == types.Uint32:
		return "(" + x + " | 0)"
	case tk == types.Uint32 && fk == types.Int32:
		return "(" + x + " >>> 0)"
	case tk != fk && fk == types.Float64:
		fail(call.Pos(), "conversion of float64 to %s is not supported", to)
	}
	return x
}
