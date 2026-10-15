// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// fnCtx translates the body of a function. The function literals inside share its names and type arguments.
type fnCtx struct {
	e     *emitter
	pkg   *Package
	info  *types.Info
	subst map[*types.TypeParam]types.Type
	names *localNames
	b     strings.Builder
	ind   int
	sig   *types.Signature
	// results are the names of the named results.
	results []string
	targets []target
	// rangeDepth counts the ranges over iterators around the current statement, whose bodies are callbacks.
	rangeDepth int
	// rangeReturn tells, whether a return inside a range over an iterator needs $rr and $rv.
	rangeReturn bool
	// rangeBranch tells, whether a labeled break or continue leaves a range over an iterator, which needs $rb.
	rangeBranch bool
	// captured are the variables, which a function literal of the function refers to, see unaliased
	captured map[types.Object]bool
	// pos is the position of the statement, which is translated, for the source map
	pos token.Pos
}

type targetKind int

const (
	targetLoop targetKind = iota
	targetSwitch
	targetRangeFunc
)

type target struct {
	kind  targetKind
	label string // the label of Go or ""
	js    string // the label of JavaScript or ""
	// pending are the labeled branches, which leave the callback of a range over an iterator, see branchTo
	pending *[]pendingBranch
}

// pendingBranch is a labeled break or continue, which the callback of a range over an iterator stores in $rb,
// before it stops the iteration, so that the code after the call of the iterator performs it.
type pendingBranch struct {
	tok   token.Token
	label string
}

func (p pendingBranch) code() string {
	return jsString(p.tok.String() + " " + p.label)
}

type localNames struct {
	m    map[types.Object]string
	used map[string]bool
	tmp  int
}

func newCtx(e *emitter, p *Package, subst map[*types.TypeParam]types.Type) *fnCtx {
	return &fnCtx{
		e:     e,
		pkg:   p,
		info:  p.Info,
		subst: subst,
		names: &localNames{m: map[types.Object]string{}, used: map[string]bool{}},
	}
}

// nested returns a context for a function literal.
func (c *fnCtx) nested(sig *types.Signature) *fnCtx {
	return &fnCtx{e: c.e, pkg: c.pkg, info: c.info, subst: c.subst, names: c.names, ind: 1, sig: sig, captured: c.captured, pos: c.pos}
}

var jsReserved = set("break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete", "do",
	"else", "enum", "export", "extends", "false", "finally", "for", "function", "if", "import", "in", "instanceof",
	"new", "null", "return", "super", "switch", "this", "throw", "true", "try", "typeof", "var", "void", "while",
	"with", "yield", "let", "static", "implements", "interface", "package", "private", "protected", "public",
	"await", "arguments", "eval", "undefined", "NaN", "Infinity")

// local returns the name of a local variable and declares it on first use.
func (c *fnCtx) local(obj types.Object) string {
	if n, ok := c.names.m[obj]; ok {
		return n
	}
	base := obj.Name()
	if jsReserved[base] {
		base += "$"
	}
	name := base
	for i := 2; c.names.used[name]; i++ {
		name = fmt.Sprintf("%s$%d", base, i)
	}
	c.names.used[name] = true
	c.names.m[obj] = name
	return name
}

func (c *fnCtx) tmp() string {
	c.names.tmp++
	return fmt.Sprintf("$t%d", c.names.tmp)
}

func (c *fnCtx) line(format string, args ...any) {
	c.b.WriteString(strings.Repeat("\t", c.ind))
	c.b.WriteString(c.e.mark(c.pos))
	fmt.Fprintf(&c.b, format, args...)
	c.b.WriteByte('\n')
}

// sub replaces the type parameters of t by the type arguments of the translated instance.
func (c *fnCtx) sub(t types.Type) types.Type {
	if len(c.subst) == 0 || t == nil {
		return t
	}
	return c.e.substitute(t, c.subst)
}

func (e *emitter) substitute(t types.Type, m map[*types.TypeParam]types.Type) types.Type {
	switch t := types.Unalias(t).(type) {
	case *types.TypeParam:
		if r, ok := m[t]; ok {
			return r
		}
		return t
	case *types.Pointer:
		return types.NewPointer(e.substitute(t.Elem(), m))
	case *types.Named:
		if t.TypeArgs().Len() == 0 {
			return t
		}
		var args []types.Type
		changed := false
		for a := range t.TypeArgs().Types() {
			r := e.substitute(a, m)
			changed = changed || r != a
			args = append(args, r)
		}
		if !changed {
			return t
		}
		inst, err := types.Instantiate(e.tctx, t.Origin(), args, false)
		if err != nil {
			fail(t.Obj().Pos(), "cannot instantiate %s: %v", t, err)
		}
		return inst
	case *types.Signature:
		return types.NewSignatureType(nil, nil, nil, e.substituteTuple(t.Params(), m), e.substituteTuple(t.Results(), m), t.Variadic())
	case *types.Tuple:
		return e.substituteTuple(t, m)
	case *types.Slice:
		return types.NewSlice(e.substitute(t.Elem(), m))
	}
	return t
}

func (e *emitter) substituteTuple(t *types.Tuple, m map[*types.TypeParam]types.Type) *types.Tuple {
	if t == nil {
		return nil
	}
	var vars []*types.Var
	for v := range t.Variables() {
		vars = append(vars, types.NewParam(v.Pos(), v.Pkg(), v.Name(), e.substitute(v.Type(), m)))
	}
	return types.NewTuple(vars...)
}

func (c *fnCtx) typeOf(x ast.Expr) types.Type {
	return c.sub(c.info.TypeOf(x))
}

// emitFunc translates a function or method for the given type arguments.
func (e *emitter) emitFunc(name string, src funcSrc, recvArgs, targs []types.Type) {
	decl := src.decl
	fn := src.pkg.Info.Defs[decl.Name].(*types.Func)
	sig := fn.Signature()

	subst := map[*types.TypeParam]types.Type{}
	for i := 0; i < sig.RecvTypeParams().Len() && i < len(recvArgs); i++ {
		subst[sig.RecvTypeParams().At(i)] = recvArgs[i]
	}
	for i := 0; i < sig.TypeParams().Len() && i < len(targs); i++ {
		subst[sig.TypeParams().At(i)] = targs[i]
	}

	c := newCtx(e, src.pkg, subst)
	c.sig = sig
	c.ind = 1
	if decl.Body != nil {
		c.captured = capturedObjects(src.pkg.Info, decl.Body)
	}

	var params []string
	if decl.Recv != nil {
		params = append(params, c.params(decl.Recv)...)
	}
	params = append(params, c.params(decl.Type.Params)...)
	c.namedResults(decl.Type.Results)

	if decl.Body == nil {
		fail(decl.Pos(), "%s has no body", fn.FullName())
	}
	c.stmts(decl.Body.List)

	fmt.Fprintf(&e.out, "%sfunction %s(%s) {\n", e.mark(decl.Pos()), name, strings.Join(params, ", "))
	if c.rangeReturn {
		e.out.WriteString("\tlet $rr = false, $rv;\n")
	}
	if c.rangeBranch {
		e.out.WriteString("\tlet $rb = null;\n")
	}
	e.out.WriteString(c.b.String())
	e.out.WriteString("}\n")
}

func (c *fnCtx) params(fields *ast.FieldList) []string {
	var res []string
	if fields == nil {
		return res
	}
	for _, f := range fields.List {
		if len(f.Names) == 0 {
			res = append(res, c.tmp())
			continue
		}
		for _, n := range f.Names {
			obj := c.info.Defs[n]
			if obj == nil || n.Name == "_" {
				res = append(res, c.tmp())
				continue
			}
			res = append(res, c.local(obj))
		}
	}
	return res
}

func (c *fnCtx) namedResults(fields *ast.FieldList) {
	c.results = nil
	if fields == nil {
		return
	}
	for _, f := range fields.List {
		for _, n := range f.Names {
			obj := c.info.Defs[n]
			if obj == nil {
				continue
			}
			name := c.tmp()
			if n.Name != "_" {
				name = c.local(obj)
			}
			c.results = append(c.results, name)
			c.line("let %s = %s;", name, c.zero(obj.Type()))
		}
	}
}

func (c *fnCtx) zero(t types.Type) string {
	return c.e.zero(c.sub(t))
}

func (c *fnCtx) initializer(ini *types.Initializer) {
	if len(ini.Lhs) == 1 {
		v := ini.Lhs[0]
		if v.Name() == "_" {
			c.line("void (%s);", c.expr(ini.Rhs))
			return
		}
		c.line("%s = %s;", c.e.global(v), c.val(ini.Rhs, v.Type()))
		return
	}

	t := c.tmp()
	c.line("const %s = %s;", t, c.expr(ini.Rhs))
	for i, v := range ini.Lhs {
		if v.Name() != "_" {
			c.line("%s = %s[%d];", c.e.global(v), t, i)
		}
	}
}

func (c *fnCtx) stmts(list []ast.Stmt) {
	for _, s := range list {
		c.stmt(s)
	}
}

func (c *fnCtx) block(list []ast.Stmt) {
	c.ind++
	c.stmts(list)
	c.ind--
}

func (c *fnCtx) stmt(s ast.Stmt) {
	// the lines of the statement map to it in the source map
	prev := c.pos
	c.pos = s.Pos()
	defer func() { c.pos = prev }()

	switch s := s.(type) {
	case *ast.ExprStmt:
		c.line("%s;", c.expr(s.X))
	case *ast.AssignStmt:
		c.assign(s)
	case *ast.IncDecStmt:
		op := token.ADD
		if s.Tok == token.DEC {
			op = token.SUB
		}
		c.opAssign(s.X, op, "1", c.typeOf(s.X))
	case *ast.DeclStmt:
		c.declStmt(s)
	case *ast.BlockStmt:
		c.line("{")
		c.block(s.List)
		c.line("}")
	case *ast.IfStmt:
		c.ifStmt(s)
	case *ast.ForStmt:
		c.forStmt(s, "")
	case *ast.RangeStmt:
		c.rangeStmt(s, "")
	case *ast.SwitchStmt:
		c.switchStmt(s, "")
	case *ast.TypeSwitchStmt:
		c.typeSwitch(s, "")
	case *ast.LabeledStmt:
		c.labeled(s)
	case *ast.ReturnStmt:
		c.returnStmt(s)
	case *ast.BranchStmt:
		c.branch(s)
	case *ast.EmptyStmt:
	default:
		fail(s.Pos(), "statement %T is not supported", s)
	}
}

func jsLabel(name string) string {
	return "$L_" + name
}

func (c *fnCtx) labeled(s *ast.LabeledStmt) {
	label := s.Label.Name
	switch inner := s.Stmt.(type) {
	case *ast.ForStmt:
		c.forStmt(inner, label)
	case *ast.RangeStmt:
		c.rangeStmt(inner, label)
	case *ast.SwitchStmt:
		c.switchStmt(inner, label)
	case *ast.TypeSwitchStmt:
		c.typeSwitch(inner, label)
	default:
		c.stmt(inner)
	}
}

func (c *fnCtx) push(t target) {
	c.targets = append(c.targets, t)
}

func (c *fnCtx) pop() {
	c.targets = c.targets[:len(c.targets)-1]
}

func (c *fnCtx) branch(s *ast.BranchStmt) {
	if s.Label != nil {
		c.branchTo(pendingBranch{tok: s.Tok, label: s.Label.Name})
		return
	}
	find := func(pred func(target) bool) target {
		for i := len(c.targets) - 1; i >= 0; i-- {
			if pred(c.targets[i]) {
				return c.targets[i]
			}
		}
		fail(s.Pos(), "no target for %s", s.Tok)
		return target{}
	}

	switch s.Tok {
	case token.BREAK:
		var t target
		if s.Label != nil {
			t = find(func(t target) bool { return t.label == s.Label.Name })
		} else {
			t = find(func(target) bool { return true })
		}
		switch t.kind {
		case targetRangeFunc:
			c.line("return false;")
		case targetSwitch:
			c.line("break %s;", t.js)
		default:
			if s.Label != nil {
				c.line("break %s;", t.js)
			} else {
				c.line("break;")
			}
		}
	case token.CONTINUE:
		var t target
		if s.Label != nil {
			t = find(func(t target) bool { return t.label == s.Label.Name })
		} else {
			t = find(func(t target) bool { return t.kind != targetSwitch })
		}
		switch {
		case t.kind == targetRangeFunc:
			c.line("return true;")
		case s.Label != nil:
			c.line("continue %s;", t.js)
		default:
			c.line("continue;")
		}
	default:
		fail(s.Pos(), "%s is not supported", s.Tok)
	}
}

// branchTo translates a labeled break or continue. If it leaves the callback of a range over an iterator, the
// callback stores it in $rb and stops the iteration, and the code after the call of the iterator performs it.
func (c *fnCtx) branchTo(b pendingBranch) {
	idx := -1
	for i := len(c.targets) - 1; i >= 0; i-- {
		if c.targets[i].label == b.label {
			idx = i
			break
		}
	}
	if idx < 0 {
		fail(token.NoPos, "no target for %s %s", b.tok, b.label)
	}
	for i := len(c.targets) - 1; i > idx; i-- {
		if c.targets[i].kind == targetRangeFunc {
			*c.targets[i].pending = append(*c.targets[i].pending, b)
			c.rangeBranch = true
			c.line("$rb = %s;", b.code())
			c.line("return false;")
			return
		}
	}

	t := c.targets[idx]
	switch {
	case t.kind == targetRangeFunc && b.tok == token.BREAK:
		c.line("return false;")
	case t.kind == targetRangeFunc:
		c.line("return true;")
	case b.tok == token.BREAK:
		c.line("break %s;", t.js)
	default:
		c.line("continue %s;", t.js)
	}
}

func (c *fnCtx) assign(s *ast.AssignStmt) {
	define := s.Tok == token.DEFINE
	switch s.Tok {
	case token.DEFINE, token.ASSIGN:
	default:
		op := map[token.Token]token.Token{
			token.ADD_ASSIGN: token.ADD, token.SUB_ASSIGN: token.SUB, token.MUL_ASSIGN: token.MUL,
			token.QUO_ASSIGN: token.QUO, token.REM_ASSIGN: token.REM, token.AND_ASSIGN: token.AND,
			token.OR_ASSIGN: token.OR, token.XOR_ASSIGN: token.XOR, token.SHL_ASSIGN: token.SHL,
			token.SHR_ASSIGN: token.SHR, token.AND_NOT_ASSIGN: token.AND_NOT,
		}[s.Tok]
		c.opAssign(s.Lhs[0], op, c.expr(s.Rhs[0]), c.typeOf(s.Rhs[0]))
		return
	}

	if len(s.Lhs) == len(s.Rhs) && len(s.Lhs) == 1 {
		c.store(s.Lhs[0], c.val(s.Rhs[0], c.typeOf(s.Lhs[0])), define)
		return
	}

	// Like Go, the operands of the left sides are evaluated first, then the right sides, and the assignments
	// follow from left to right, so that p, p.X = q, 1 sets X of the former p.
	targets := make([]func(js string), len(s.Lhs))
	for i, l := range s.Lhs {
		targets[i] = c.target(l, define)
	}

	if len(s.Lhs) == len(s.Rhs) {
		var tmps []string
		for i, r := range s.Rhs {
			t := c.tmp()
			c.line("const %s = %s;", t, c.val(r, c.typeOf(s.Lhs[i])))
			tmps = append(tmps, t)
		}
		for i := range s.Lhs {
			targets[i](tmps[i])
		}
		return
	}

	t := c.tmp()
	c.line("const %s = %s;", t, c.expr(s.Rhs[0]))
	for i := range s.Lhs {
		targets[i](fmt.Sprintf("%s[%d]", t, i))
	}
}

// target evaluates the operand of a field on the left side of an assignment and returns the function, which
// assigns the value later. A variable is assigned directly.
func (c *fnCtx) target(lhs ast.Expr, define bool) func(js string) {
	se, ok := ast.Unparen(lhs).(*ast.SelectorExpr)
	if !ok || c.info.Selections[se] == nil {
		return func(js string) { c.store(lhs, js, define) }
	}
	obj := c.tmp()
	c.line("const %s = %s;", obj, c.expr(se.X))
	field := obj + "." + c.fieldName(se)
	t := c.typeOf(lhs)
	return func(js string) {
		if isStruct(t) && !c.e.mut.immutable(t) {
			c.line("%s$s(%s, %s);", c.e.structName(t), field, js)
			return
		}
		c.line("%s = %s;", field, js)
	}
}

// store assigns the value js to lhs. A struct is copied into the existing value, so that a pointer to it sees
// the new value, as in Go, unless its type is immutable, see [mutation]: then its value may be shared and is
// replaced.
func (c *fnCtx) store(lhs ast.Expr, js string, define bool) {
	if id, ok := lhs.(*ast.Ident); ok {
		if id.Name == "_" {
			c.line("void (%s);", js)
			return
		}
		if define && c.info.Defs[id] != nil {
			c.line("let %s = %s;", c.local(c.info.Defs[id]), js)
			return
		}
	}

	t := c.typeOf(lhs)
	if isStruct(t) && !c.e.mut.immutable(t) {
		c.line("%s$s(%s, %s);", c.e.structName(t), c.expr(lhs), js)
		return
	}
	c.line("%s = %s;", c.lvalue(lhs), js)
}

// lvalue returns the target of an assignment.
func (c *fnCtx) lvalue(lhs ast.Expr) string {
	switch l := ast.Unparen(lhs).(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return c.expr(l)
	}
	fail(lhs.Pos(), "cannot assign to %T", lhs)
	return ""
}

func (c *fnCtx) opAssign(lhs ast.Expr, op token.Token, rhs string, rt types.Type) {
	lt := c.typeOf(lhs)
	target := ast.Unparen(lhs)
	// evaluate the operand of a selector once
	if se, ok := target.(*ast.SelectorExpr); ok {
		if _, isIdent := ast.Unparen(se.X).(*ast.Ident); !isIdent && c.info.Selections[se] != nil {
			t := c.tmp()
			c.line("const %s = %s;", t, c.expr(se.X))
			field := t + "." + c.fieldName(se)
			c.line("%s = %s;", field, c.binaryJS(op, field, rhs, lt, rt, lhs.Pos()))
			return
		}
	}
	l := c.lvalue(lhs)
	c.line("%s = %s;", l, c.binaryJS(op, l, rhs, lt, rt, lhs.Pos()))
}

func (c *fnCtx) declStmt(s *ast.DeclStmt) {
	g, ok := s.Decl.(*ast.GenDecl)
	if !ok {
		fail(s.Pos(), "declaration is not supported")
	}
	switch g.Tok {
	case token.CONST:
		return
	case token.VAR:
	default:
		fail(s.Pos(), "%s declarations are not supported inside functions", g.Tok)
	}

	for _, spec := range g.Specs {
		vs := spec.(*ast.ValueSpec)
		switch {
		case len(vs.Values) == 0:
			for _, n := range vs.Names {
				if n.Name == "_" {
					continue
				}
				obj := c.info.Defs[n]
				c.line("let %s = %s;", c.local(obj), c.zero(obj.Type()))
			}
		case len(vs.Values) == len(vs.Names):
			for i, n := range vs.Names {
				js := c.val(vs.Values[i], c.typeOf(n))
				if n.Name == "_" {
					c.line("void (%s);", js)
					continue
				}
				c.line("let %s = %s;", c.local(c.info.Defs[n]), js)
			}
		default:
			t := c.tmp()
			c.line("const %s = %s;", t, c.expr(vs.Values[0]))
			for i, n := range vs.Names {
				if n.Name != "_" {
					c.line("let %s = %s[%d];", c.local(c.info.Defs[n]), t, i)
				}
			}
		}
	}
}

func (c *fnCtx) ifStmt(s *ast.IfStmt) {
	if s.Init != nil {
		c.line("{")
		c.ind++
		c.stmt(s.Init)
	}

	c.line("if (%s) {", c.expr(s.Cond))
	c.block(s.Body.List)
	if s.Else != nil {
		c.line("} else {")
		c.ind++
		if b, ok := s.Else.(*ast.BlockStmt); ok {
			c.stmts(b.List)
		} else {
			c.stmt(s.Else)
		}
		c.ind--
	}
	c.line("}")

	if s.Init != nil {
		c.ind--
		c.line("}")
	}
}

func (c *fnCtx) labelPrefix(label string) string {
	if label == "" {
		return ""
	}
	return jsLabel(label) + ": "
}

func (c *fnCtx) forStmt(s *ast.ForStmt, label string) {
	init := ""
	wrapped := false
	if s.Init != nil {
		if a, ok := s.Init.(*ast.AssignStmt); ok && a.Tok == token.DEFINE && len(a.Lhs) == len(a.Rhs) {
			var parts []string
			for i, l := range a.Lhs {
				id := l.(*ast.Ident)
				name := c.tmp()
				if id.Name != "_" {
					name = c.local(c.info.Defs[id])
				}
				parts = append(parts, name+" = "+c.val(a.Rhs[i], c.typeOf(l)))
			}
			init = "let " + strings.Join(parts, ", ")
		} else if ok && a.Tok == token.DEFINE {
			// for a, b := f(); ...: the variables of the results are per iteration like the others
			wrapped = true
			c.line("{")
			c.ind++
			res := c.tmp()
			c.line("const %s = %s;", res, c.expr(a.Rhs[0]))
			var parts []string
			for i, l := range a.Lhs {
				id := l.(*ast.Ident)
				name := c.tmp()
				if id.Name != "_" {
					name = c.local(c.info.Defs[id])
				}
				parts = append(parts, fmt.Sprintf("%s = %s[%d]", name, res, i))
			}
			init = "let " + strings.Join(parts, ", ")
		} else {
			// the variables are not per iteration, but the statement does not define any anyway
			wrapped = true
			c.line("{")
			c.ind++
			c.stmt(s.Init)
		}
	}

	cond := ""
	if s.Cond != nil {
		cond = c.expr(s.Cond)
	}

	post := ""
	if s.Post != nil {
		post = c.simpleExpr(s.Post)
	}

	c.line("%sfor (%s; %s; %s) {", c.labelPrefix(label), init, cond, post)
	c.push(target{kind: targetLoop, label: label, js: jsLabel(label)})
	c.block(s.Body.List)
	c.pop()
	c.line("}")

	if wrapped {
		c.ind--
		c.line("}")
	}
}

// simpleExpr translates the post statement of a for loop into an expression.
func (c *fnCtx) simpleExpr(s ast.Stmt) string {
	sub := c.nested(c.sig)
	sub.subst = c.subst
	sub.ind = 0
	sub.stmt(s)
	code := strings.TrimSpace(sub.b.String())
	if strings.Contains(code, "\n") || strings.HasPrefix(code, "const ") || strings.HasPrefix(code, "let ") {
		// several statements become an immediately called function
		return "(() => { " + strings.ReplaceAll(code, "\n", " ") + " })()"
	}
	return strings.TrimSuffix(code, ";")
}

func (c *fnCtx) bindRange(key ast.Expr, js string, define bool) {
	if key == nil {
		return
	}
	if id, ok := key.(*ast.Ident); ok && id.Name == "_" {
		return
	}
	c.store(key, js, define)
}

func (c *fnCtx) rangeStmt(s *ast.RangeStmt, label string) {
	define := s.Tok == token.DEFINE
	xt := c.typeOf(s.X)

	if b, ok := xt.Underlying().(*types.Basic); ok && b.Info()&types.IsInteger != 0 {
		i, n := c.tmp(), c.tmp()
		c.line("%sfor (let %s = 0, %s = %s; %s < %s; %s++) {", c.labelPrefix(label), i, n, c.expr(s.X), i, n, i)
		c.ind++
		c.bindRange(s.Key, i, define)
		c.ind--
		c.push(target{kind: targetLoop, label: label, js: jsLabel(label)})
		c.block(s.Body.List)
		c.pop()
		c.line("}")
		return
	}

	if c.builtinRange(s, label, define) {
		return
	}

	c.rangeFunc(s, label, define)
}

// builtinRange translates a range over the iterators of list.List and dict.Map into a plain loop.
func (c *fnCtx) builtinRange(s *ast.RangeStmt, label string, define bool) bool {
	call, ok := ast.Unparen(s.X).(*ast.CallExpr)
	if !ok {
		return false
	}
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	recvT := c.typeOf(se.X)
	if !isOpaque(recvT) {
		return false
	}

	named := types.Unalias(recvT).(*types.Named)
	args := typeArgs(named.TypeArgs())
	method := named.Obj().Name() + "." + se.Sel.Name

	src, i := c.tmp(), c.tmp()
	c.line("{")
	c.ind++
	c.line("const %s = %s;", src, c.expr(se.X))

	switch method {
	case "List.All", "List.Values", "Frozen.All", "Frozen.Values", "Out.Values":
		c.line("%sfor (let %s = 0; %s < %s.length; %s++) {", c.labelPrefix(label), i, i, src, i)
		c.ind++
		elem := c.e.clone(src+"["+i+"]", args[0])
		if strings.HasSuffix(method, ".All") {
			c.bindRange(s.Key, i, define)
			c.bindRange(s.Value, elem, define)
		} else {
			c.bindRange(s.Key, elem, define)
		}
		c.ind--
	case "Map.All", "Map.Keys":
		keys, k := c.tmp(), c.tmp()
		c.line("const %s = $dkeys(%s);", keys, src)
		c.line("%sfor (let %s = 0; %s < %s.length; %s++) {", c.labelPrefix(label), i, i, keys, i)
		c.ind++
		c.line("const %s = %s[%s];", k, keys, i)
		c.line("if (!%s.has(%s)) continue;", src, k)
		c.bindRange(s.Key, k, define)
		if method == "Map.All" {
			c.bindRange(s.Value, c.e.clone(src+".get("+k+")", args[1]), define)
		}
		c.ind--
	default:
		fail(s.Pos(), "range over %s is not supported", method)
	}

	c.push(target{kind: targetLoop, label: label, js: jsLabel(label)})
	c.block(s.Body.List)
	c.pop()
	c.line("}")
	c.ind--
	c.line("}")
	return true
}

// rangeFunc translates a range over an iterator into a call of the iterator with the body as callback. The
// callback returns false to stop, a return of the function sets $rr and $rv, which are checked after the call.
func (c *fnCtx) rangeFunc(s *ast.RangeStmt, label string, define bool) {
	sig, ok := c.typeOf(s.X).Underlying().(*types.Signature)
	if !ok {
		fail(s.X.Pos(), "range over %s is not supported", c.typeOf(s.X))
	}
	yield := sig.Params().At(0).Type().Underlying().(*types.Signature)

	vars := []ast.Expr{s.Key, s.Value}
	var params []string
	for i := 0; i < yield.Params().Len(); i++ {
		params = append(params, c.tmp())
	}

	c.line("%s((%s) => {", c.expr(s.X), strings.Join(params, ", "))
	c.ind++
	for i, p := range params {
		if i < len(vars) {
			c.bindRange(vars[i], c.e.clone(p, yield.Params().At(i).Type()), define)
		}
	}
	pending := &[]pendingBranch{}
	c.push(target{kind: targetRangeFunc, label: label, pending: pending})
	c.rangeDepth++
	c.stmts(s.Body.List)
	c.rangeDepth--
	c.pop()
	c.line("return true;")
	c.ind--
	c.line("});")
	defer c.dispatchBranches(*pending)

	if c.rangeReturn {
		if c.rangeDepth > 0 {
			c.line("if ($rr) return false;")
		} else if c.sig.Results().Len() > 0 {
			c.line("if ($rr) return $rv;")
		} else {
			c.line("if ($rr) return;")
		}
	}
}

// dispatchBranches performs the labeled branches, which the callback of a range over an iterator has stored in $rb.
func (c *fnCtx) dispatchBranches(pending []pendingBranch) {
	seen := map[pendingBranch]bool{}
	for _, b := range pending {
		if seen[b] {
			continue
		}
		seen[b] = true
		c.line("if ($rb === %s) {", b.code())
		c.ind++
		c.line("$rb = null;")
		c.branchTo(b)
		c.ind--
		c.line("}")
	}
}

func (c *fnCtx) switchStmt(s *ast.SwitchStmt, label string) {
	c.line("{")
	c.ind++
	if s.Init != nil {
		c.stmt(s.Init)
	}

	tag := ""
	var tagType types.Type
	if s.Tag != nil {
		tag = c.tmp()
		tagType = c.typeOf(s.Tag)
		c.line("const %s = %s;", tag, c.expr(s.Tag))
	}

	js := c.tmp() + "sw"
	if label != "" {
		js = jsLabel(label)
	}
	c.line("%s: {", js)
	c.ind++
	c.push(target{kind: targetSwitch, label: label, js: js})

	first := true
	var def *ast.CaseClause
	for _, st := range s.Body.List {
		cc := st.(*ast.CaseClause)
		if cc.List == nil {
			def = cc
			continue
		}
		var conds []string
		for _, x := range cc.List {
			switch {
			case tag == "":
				conds = append(conds, c.expr(x))
			case c.isNil(x):
				conds = append(conds, "("+tag+" === null)")
			default:
				conds = append(conds, c.e.eq(tag, c.expr(x), tagType))
			}
		}
		if first {
			c.line("if (%s) {", strings.Join(conds, " || "))
		} else {
			c.line("} else if (%s) {", strings.Join(conds, " || "))
		}
		first = false
		c.block(cc.Body)
	}
	if def != nil {
		if first {
			c.line("{")
		} else {
			c.line("} else {")
		}
		c.block(def.Body)
		first = false
	}
	if !first {
		c.line("}")
	}

	c.pop()
	c.ind--
	c.line("}")
	c.ind--
	c.line("}")
}

// typeSwitch translates a type switch over a sealed interface: the box of an interface value holds the method
// table of its type, which identifies the variant.
func (c *fnCtx) typeSwitch(s *ast.TypeSwitchStmt, label string) {
	c.line("{")
	c.ind++
	if s.Init != nil {
		c.stmt(s.Init)
	}

	var x ast.Expr
	switch a := s.Assign.(type) {
	case *ast.ExprStmt:
		x = a.X.(*ast.TypeAssertExpr).X
	case *ast.AssignStmt:
		x = a.Rhs[0].(*ast.TypeAssertExpr).X
	}
	v := c.tmp()
	c.line("const %s = %s;", v, c.expr(x))

	js := c.tmp() + "sw"
	if label != "" {
		js = jsLabel(label)
	}
	c.line("%s: {", js)
	c.ind++
	c.push(target{kind: targetSwitch, label: label, js: js})

	bind := func(cc *ast.CaseClause) {
		obj := c.info.Implicits[cc]
		if obj == nil || obj.Name() == "_" {
			return
		}
		t := c.sub(obj.Type())
		value := v
		if !isInterface(t) {
			// a single type binds a copy of the value of the variant
			value = c.e.clone(v+".v", t)
		}
		c.line("let %s = %s;", c.local(obj), value)
	}

	first := true
	var def *ast.CaseClause
	for _, st := range s.Body.List {
		cc := st.(*ast.CaseClause)
		if cc.List == nil {
			def = cc
			continue
		}
		var conds []string
		for _, e := range cc.List {
			conds = append(conds, v+".m === "+c.e.methodTable(c.typeOf(e)))
		}
		if first {
			c.line("if (%s) {", strings.Join(conds, " || "))
		} else {
			c.line("} else if (%s) {", strings.Join(conds, " || "))
		}
		first = false
		c.ind++
		bind(cc)
		c.stmts(cc.Body)
		c.ind--
	}
	if def != nil {
		if first {
			c.line("{")
		} else {
			c.line("} else {")
		}
		c.ind++
		bind(def)
		c.stmts(def.Body)
		c.ind--
		first = false
	}
	if !first {
		c.line("}")
	}

	c.pop()
	c.ind--
	c.line("}")
	c.ind--
	c.line("}")
}

func (c *fnCtx) returnStmt(s *ast.ReturnStmt) {
	results := c.sub(c.sig.Results()).(*types.Tuple)
	var vals []string
	switch {
	case len(s.Results) == 0:
		for i, r := range c.results {
			vals = append(vals, c.e.clone(r, results.At(i).Type()))
		}
	case len(s.Results) == 1 && results.Len() > 1:
		// return f() with several results
		vals = append(vals, c.expr(s.Results[0]))
	default:
		for i, r := range s.Results {
			vals = append(vals, c.val(r, results.At(i).Type()))
		}
	}

	js := ""
	switch {
	case len(vals) == 1:
		js = vals[0]
	case len(vals) > 1:
		js = "[" + strings.Join(vals, ", ") + "]"
	}

	if c.rangeDepth > 0 {
		c.rangeReturn = true
		if js != "" {
			c.line("$rv = %s;", js)
		}
		c.line("$rr = true;")
		c.line("return false;")
		return
	}

	if js == "" {
		c.line("return;")
	} else {
		c.line("return %s;", js)
	}
}
