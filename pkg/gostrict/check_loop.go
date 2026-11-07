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
	"strconv"
	"strings"
)

// The rules of this file complete the check of the loops of the terminating profile (check_term.go): a bound or a
// counter, which a function literal changes, that the loop may call, and the rewrites of the reports with their
// mechanical fixes (R-SCRIPT-GS-LOOP-FIX).

// loopFacts are the facts of a file, which the check of its loops needs.
type loopFacts struct {
	// lits are the function literals of the file in the order of their position
	lits []*ast.FuncLit
	// labels are the labeled loops
	labels map[*ast.ForStmt]*ast.LabeledStmt
	// callees are the identifiers, which a call calls directly, e.g. grow of grow(), and targets those, to which an
	// assignment assigns, e.g. grow of grow = func() { ... }
	callees, targets map[*ast.Ident]bool
	// vars are the variables, to which a literal is assigned, e.g. grow := func() { ... }
	vars map[*ast.FuncLit]*types.Var
	// uses are the identifiers, which use an object
	uses map[types.Object][]*ast.Ident
	// assigns are the positions of the assignments of a variable or of a field of it, e.g. of s by s.n = 1
	assigns map[types.Object][]token.Pos
	// declared are the names, which the fixes of the file declare in a block
	declared map[*types.Scope]map[string]bool
}

func (c *checker) loopFacts(f *ast.File) *loopFacts {
	facts := &loopFacts{
		labels:   map[*ast.ForStmt]*ast.LabeledStmt{},
		callees:  map[*ast.Ident]bool{},
		targets:  map[*ast.Ident]bool{},
		vars:     map[*ast.FuncLit]*types.Var{},
		uses:     map[types.Object][]*ast.Ident{},
		assigns:  map[types.Object][]token.Pos{},
		declared: map[*types.Scope]map[string]bool{},
	}
	assigned := func(x ast.Expr) {
		if id := rootIdent(x); id != nil {
			if obj := c.info.Uses[id]; obj != nil {
				facts.assigns[obj] = append(facts.assigns[obj], id.Pos())
			}
		}
	}
	bind := func(lhs ast.Expr, rhs ast.Expr) {
		lit, ok := ast.Unparen(rhs).(*ast.FuncLit)
		id, isID := ast.Unparen(lhs).(*ast.Ident)
		if !ok || !isID {
			return
		}
		obj := c.info.Defs[id]
		if obj == nil {
			obj = c.info.Uses[id]
		}
		if v, ok := obj.(*types.Var); ok {
			facts.vars[lit] = v
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			facts.lits = append(facts.lits, n)
		case *ast.LabeledStmt:
			if loop, ok := n.Stmt.(*ast.ForStmt); ok {
				facts.labels[loop] = n
			}
		case *ast.CallExpr:
			if id, ok := ast.Unparen(n.Fun).(*ast.Ident); ok {
				facts.callees[id] = true
			}
		case *ast.Ident:
			if obj := c.info.Uses[n]; obj != nil {
				facts.uses[obj] = append(facts.uses[obj], n)
			}
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if n.Tok != token.DEFINE {
					assigned(lhs)
				}
				if id, ok := ast.Unparen(lhs).(*ast.Ident); ok {
					facts.targets[id] = true
				}
				if len(n.Lhs) == len(n.Rhs) {
					bind(lhs, n.Rhs[i])
				}
			}
		case *ast.ValueSpec:
			if len(n.Names) == len(n.Values) {
				for i, id := range n.Names {
					bind(id, n.Values[i])
				}
			}
		case *ast.IncDecStmt:
			assigned(n.X)
		case *ast.RangeStmt:
			if n.Tok == token.ASSIGN {
				assigned(n.Key)
				if n.Value != nil {
					assigned(n.Value)
				}
			}
		}
		return true
	})
	return facts
}

// rootIdent returns the variable, whose value an assignment to x changes, e.g. s of s.f.g, or nil.
func rootIdent(x ast.Expr) *ast.Ident {
	for {
		switch e := x.(type) {
		case *ast.Ident:
			return e
		case *ast.ParenExpr:
			x = e.X
		case *ast.SelectorExpr:
			x = e.X
		case *ast.StarExpr:
			x = e.X
		case *ast.IndexExpr:
			x = e.X
		default:
			return nil
		}
	}
}

func within(pos token.Pos, n ast.Node) bool {
	return n != nil && pos >= n.Pos() && pos < n.End()
}

// inLoop tells, whether pos is in the condition, the post statement or the body of the loop, which run in every
// iteration, unlike its init statement.
func (l *loopCheck) inLoop(pos token.Pos) bool {
	from := l.n.Body.Pos()
	if l.n.Cond != nil {
		from = l.n.Cond.Pos()
	}
	return pos >= from && pos < l.n.End()
}

// outerLit returns the outermost function literal, which contains pos, but not the loop, or nil: the value, which
// the loop must call to run the code at pos.
func (l *loopCheck) outerLit(pos token.Pos) *ast.FuncLit {
	for _, lit := range l.facts.lits {
		if within(pos, lit) && !within(l.n.Pos(), lit) {
			return lit
		}
	}
	return nil
}

// changedByLiteral tells, whether a function literal outside of the loop, which the loop may call, assigns v or a
// field of it, e.g. grow := func() { n++ } with grow() in the body.
func (l *loopCheck) changedByLiteral(v types.Object) bool {
	for _, pos := range l.facts.assigns[v] {
		if l.inLoop(pos) {
			// the loop itself assigns v, which the callers check
			continue
		}
		if lit := l.outerLit(pos); lit != nil && l.reaches(lit, map[*ast.FuncLit]bool{}) {
			return true
		}
	}
	return false
}

// reaches tells, whether the loop may call the function literal: a call of the variable, to which it is assigned,
// in the loop or in another literal, which the loop reaches. A literal, which escapes, e.g. into a field, a list or
// as an argument, may be reached by any call, see [loopCheck.mayCallValues].
func (l *loopCheck) reaches(lit *ast.FuncLit, seen map[*ast.FuncLit]bool) bool {
	if seen[lit] {
		return false
	}
	seen[lit] = true
	v, ok := l.facts.vars[lit]
	if !ok {
		return l.mayCallValues()
	}
	for _, u := range l.facts.uses[v] {
		if l.facts.targets[u] {
			// v = func() { ... } assigns a function, which a call of v reaches
			continue
		}
		if l.inLoop(u.Pos()) {
			return true
		}
		if !l.facts.callees[u] {
			// the value escapes
			return l.mayCallValues()
		}
		if other := l.outerLit(u.Pos()); other != nil && other != lit && l.reaches(other, seen) {
			return true
		}
	}
	return false
}

// mayCallValues tells, whether the loop may call a function value: it calls a function other than a conversion, a
// builtin or a function of the library without function parameters, or it ranges over a function value.
func (l *loopCheck) mayCallValues() bool {
	found := false
	l.inspectLoop(func(n ast.Node) {
		switch n := n.(type) {
		case *ast.CallExpr:
			if fn := l.c.staticCallee(n); fn == nil || !isLibFunc(fn) || hasFuncParam(fn) {
				if !l.c.isConversion(n) && !l.c.isBuiltinCall(n) {
					found = true
				}
			}
		case *ast.RangeStmt:
			if l.c.rangesOverValue(n) {
				found = true
			}
		}
	})
	return found
}

// inspectLoop visits the nodes of the condition, the post statement and the body of the loop.
func (l *loopCheck) inspectLoop(visit func(ast.Node)) {
	var nodes []ast.Node
	if l.n.Cond != nil {
		nodes = append(nodes, l.n.Cond)
	}
	if l.n.Post != nil {
		nodes = append(nodes, l.n.Post)
	}
	for _, n := range append(nodes, l.n.Body) {
		ast.Inspect(n, func(x ast.Node) bool {
			if x != nil {
				visit(x)
			}
			return true
		})
	}
}

// staticCallee returns the function or method, which the call calls by its name, or nil for a function value, a
// method of an interface, a conversion or a builtin.
func (c *checker) staticCallee(call *ast.CallExpr) *types.Func {
	fun := ast.Unparen(call.Fun)
	switch x := fun.(type) {
	case *ast.IndexExpr:
		fun = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		fun = ast.Unparen(x.X)
	}
	switch x := fun.(type) {
	case *ast.Ident:
		fn, _ := c.info.Uses[x].(*types.Func)
		return fn
	case *ast.SelectorExpr:
		if sel := c.info.Selections[x]; sel != nil {
			if sel.Kind() != types.MethodVal || isInterface(sel.Recv()) {
				return nil
			}
			fn, _ := sel.Obj().(*types.Func)
			return fn
		}
		fn, _ := c.info.Uses[x.Sel].(*types.Func)
		return fn
	}
	return nil
}

func (c *checker) isConversion(call *ast.CallExpr) bool {
	tv, ok := c.info.Types[call.Fun]
	return ok && tv.IsType()
}

func (c *checker) isBuiltinCall(call *ast.CallExpr) bool {
	id, ok := ast.Unparen(call.Fun).(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = c.info.Uses[id].(*types.Builtin)
	return ok
}

// isLibFunc tells, whether fn belongs to the library of gostrict or to an allowed package of the standard library.
func isLibFunc(fn *types.Func) bool {
	if fn.Pkg() == nil {
		return false
	}
	p := fn.Pkg().Path()
	_, std := stdAllowed[p]
	return std || strings.HasPrefix(p, stdPrefix)
}

// hasFuncParam tells, whether a parameter of fn is a function, which fn may call, e.g. the cmp of list.SortFunc.
func hasFuncParam(fn *types.Func) bool {
	for p := range fn.Origin().Signature().Params().Variables() {
		if _, ok := p.Type().Underlying().(*types.Signature); ok {
			return true
		}
	}
	return false
}

// rangesOverValue tells, whether the range calls a function value: a range over a function, which is not an
// iterator of the library, e.g. l.All().
func (c *checker) rangesOverValue(r *ast.RangeStmt) bool {
	if _, ok := c.info.TypeOf(r.X).Underlying().(*types.Signature); !ok {
		return false
	}
	call, ok := ast.Unparen(r.X).(*ast.CallExpr)
	if !ok {
		return true
	}
	fn := c.staticCallee(call)
	return fn == nil || !isLibFunc(fn)
}

// listMutators are the methods of the library, which change a list, a dict or an outbox.
var listMutators = map[string]bool{
	listPath + ".List.Append": true, listPath + ".List.Set": true, listPath + ".List.Delete": true,
	listPath + ".List.Pop": true, dictPath + ".Map.Set": true, dictPath + ".Map.Delete": true,
	msgPath + ".Out.Send": true,
}

// keepsLists tells, whether the node changes no list, dict or outbox: it does not call one of their methods, which
// change them, and calls only functions of the library, which take no function, conversions, builtins and functions
// and methods by their name, whose receiver and arguments can not refer to a list, dict, outbox or function.
func (c *checker) keepsLists(n ast.Node) bool {
	keeps := true
	ast.Inspect(n, func(x ast.Node) bool {
		switch x := x.(type) {
		case *ast.CallExpr:
			if !c.keepsListsCall(x) {
				keeps = false
			}
		case *ast.RangeStmt:
			if c.rangesOverValue(x) {
				keeps = false
			}
		}
		return keeps
	})
	return keeps
}

func (c *checker) keepsListsCall(call *ast.CallExpr) bool {
	if c.isConversion(call) || c.isBuiltinCall(call) {
		return true
	}
	fn := c.staticCallee(call)
	if fn == nil {
		return false
	}
	if isLibFunc(fn) {
		key := fn.Pkg().Path() + "." + fn.Name()
		if r := fn.Signature().Recv(); r != nil {
			if n, ok := types.Unalias(derefType(r.Type())).(*types.Named); ok {
				key = fn.Pkg().Path() + "." + n.Obj().Name() + "." + fn.Name()
			}
		}
		return !listMutators[key] && !hasFuncParam(fn)
	}
	if se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok && c.info.Selections[se] != nil && mayRefer(c.info.TypeOf(se.X)) {
		return false
	}
	for _, a := range call.Args {
		t := c.info.TypeOf(a)
		if tuple, ok := t.(*types.Tuple); ok {
			for v := range tuple.Variables() {
				if mayRefer(v.Type()) {
					return false
				}
			}
			continue
		}
		if mayRefer(t) {
			return false
		}
	}
	return true
}

// mayRefer tells, whether a value of t may refer to a list, a dict, an outbox or a function, through which a call
// may change a list. Handles, big.Int and decimal.Dec refer to nothing of the program.
func mayRefer(t types.Type) bool {
	return mayReferSeen(t, map[types.Type]bool{})
}

func mayReferSeen(t types.Type, seen map[types.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	if isHandle(t) || isHandleUnion(t) {
		return false
	}
	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() != nil {
		switch p := n.Obj().Pkg().Path(); {
		case p == bigPath || p == decimalPath:
			return false
		case isOption(t) || (p == listPath && n.Obj().Name() == "Frozen"):
			// an option and a frozen list can not change, but their elements may refer to a list
			return mayReferSeen(n.TypeArgs().At(0), seen)
		case isOpaque(t):
			return true
		}
	}
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return true
	}
	switch u := t.Underlying().(type) {
	case *types.Pointer:
		return mayReferSeen(u.Elem(), seen)
	case *types.Signature, *types.Interface:
		return true
	case *types.Struct:
		for f := range u.Fields() {
			if mayReferSeen(f.Type(), seen) {
				return true
			}
		}
	}
	return false
}

// start returns the position of the statement of the loop, before which a fix declares a variable: its label or
// the loop.
func (l *loopCheck) start() token.Pos {
	if lab, ok := l.facts.labels[l.n]; ok {
		return lab.Pos()
	}
	return l.n.Pos()
}

// freeName returns base or base with a number, which no declaration of the block of the loop, of the loop itself
// and of the enclosing scopes uses, nor the fix of another loop of the block, so that a variable of this name,
// declared before the loop, neither collides nor hides another one, and reserves it for the fix of this loop.
func (l *loopCheck) freeName(base string) string {
	scope := l.c.info.Scopes[l.n]
	var outer *types.Scope
	if scope != nil {
		outer = scope.Parent()
	}
	taken := func(name string) bool {
		if scope == nil || outer == nil {
			return false
		}
		if scope.Lookup(name) != nil || l.facts.declared[outer][name] {
			return true
		}
		_, obj := outer.LookupParent(name, token.NoPos)
		return obj != nil
	}
	name := base
	for i := 2; taken(name); i++ {
		name = base + strconv.Itoa(i)
	}
	if outer != nil {
		if l.facts.declared[outer] == nil {
			l.facts.declared[outer] = map[string]bool{}
		}
		l.facts.declared[outer][name] = true
	}
	return name
}

// declare returns the edit, which declares the variable of the statement decl before the loop or its label: on a
// line of its own with the indentation of the loop, or before it on its line, if code precedes it there.
func (l *loopCheck) declare(decl string) loopEdit {
	start := l.start()
	f := l.c.d.fset.File(start)
	p := f.Position(start)
	line := l.c.d.line(p.Filename, p.Line)
	if strings.TrimLeft(line[:min(p.Column-1, len(line))], " \t") != "" {
		return loopEdit{from: start, to: start, text: decl + "; "}
	}
	lineStart := f.LineStart(p.Line)
	return loopEdit{from: lineStart, to: lineStart, text: l.indent(l.n.Pos()) + decl + "\n"}
}

// text returns the source code of the node.
func (l *loopCheck) text(n ast.Node) string {
	f := l.c.d.fset.File(n.Pos())
	if f != nil {
		if src, ok := l.c.d.sources[f.Name()]; ok {
			from, to := f.Offset(n.Pos()), f.Offset(n.End())
			if from >= 0 && to <= len(src) && from <= to {
				return string(src[from:to])
			}
		}
	}
	if x, ok := n.(ast.Expr); ok {
		return types.ExprString(x)
	}
	return ""
}

// operand returns the source code of x as an operand of + or -, in parentheses, if needed.
func (l *loopCheck) operand(x ast.Expr) string {
	switch ast.Unparen(x).(type) {
	case *ast.Ident, *ast.BasicLit, *ast.SelectorExpr, *ast.CallExpr:
		return l.text(x)
	}
	if _, ok := x.(*ast.ParenExpr); ok {
		return l.text(x)
	}
	if b, ok := x.(*ast.BinaryExpr); ok && (b.Op == token.ADD || b.Op == token.SUB || b.Op == token.MUL || b.Op == token.QUO || b.Op == token.REM) {
		return l.text(x)
	}
	return "(" + l.text(x) + ")"
}

// indent returns the white space at the beginning of the line of pos.
func (l *loopCheck) indent(pos token.Pos) string {
	p := l.c.d.fset.Position(pos)
	line := l.c.d.line(p.Filename, p.Line)
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// exprBound reports an expression as the bound of i op bound with the rewrite, which assigns it to a variable
// before the loop, and its fix, if the expression has the same value in every iteration: no variable of it changes
// in the loop, it calls only Len and the loop changes no list.
func (l *loopCheck) exprBound(op token.Token, bound ast.Expr) loopFinding {
	base := "end"
	if l.isLen(bound) {
		base = "n"
	}
	// the rewrite names the variable like the contract, the fix declares a free name
	text := l.text(bound)
	r := loopFinding{
		msg: "the bound must be a constant or a variable, because an expression may change in the loop",
		fix: "assign the bound to a variable before the loop: " + base + " := " + text + " and " + l.counter.Name() + " " + op.String() + " " + base,
	}
	if l.hoistable(bound) {
		name := l.freeName(base)
		r.edits = []loopEdit{l.declare(name + " := " + text), {from: bound.Pos(), to: bound.End(), text: name}}
	}
	return r
}

// isLen tells, whether x is the call of Len of a list, a dict or an outbox.
func (l *loopCheck) isLen(x ast.Expr) bool {
	call, ok := ast.Unparen(x).(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok || se.Sel.Name != "Len" {
		return false
	}
	sel := l.c.info.Selections[se]
	return sel != nil && sel.Kind() == types.MethodVal && isOpaque(sel.Recv()) && !isOption(sel.Recv())
}

// hoistable tells, whether the expression keeps its value during the loop, so that it may be evaluated once
// before it: it calls only Len and conversions, it has no shift, whose type would come from the comparison, its
// variables are declared before the loop, are no pointers and are reached without a pointer, neither the loop nor
// a literal, which the loop may call, assigns them, and the loop changes no list.
func (l *loopCheck) hoistable(x ast.Expr) bool {
	c := l.c
	start := l.start()
	ok := true
	var vars []types.Object
	ast.Inspect(x, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if !c.isConversion(n) && !l.isLen(n) {
				ok = false
			}
		case *ast.BinaryExpr:
			if n.Op == token.SHL || n.Op == token.SHR {
				ok = false
			}
		case *ast.FuncLit, *ast.StarExpr, *ast.IndexExpr, *ast.IndexListExpr, *ast.CompositeLit:
			ok = false
		case *ast.SelectorExpr:
			if sel := c.info.Selections[n]; sel != nil && (sel.Indirect() || isPointer(c.info.TypeOf(n.X))) {
				ok = false
			}
		case *ast.Ident:
			v, isVar := c.info.Uses[n].(*types.Var)
			if !isVar {
				break
			}
			if v.Pkg() != nil && v.Parent() == v.Pkg().Scope() {
				// a package variable never changes
				break
			}
			if v.IsField() {
				break
			}
			if v.Pos() >= start || isPointer(v.Type()) {
				ok = false
			}
			vars = append(vars, v)
		}
		return ok
	})
	if !ok {
		return false
	}
	for _, v := range vars {
		for _, pos := range l.facts.assigns[v] {
			if within(pos, l.n) {
				return false
			}
		}
		if l.changedByLiteral(v) {
			return false
		}
	}
	if l.n.Init != nil && !c.keepsLists(l.n.Init) {
		return false
	}
	return c.keepsLists(l.n.Body)
}

// compound reports a condition, which combines the comparison of the counter with && or || with other conditions,
// with the rewrite, which compares only the counter and checks the rest at the beginning of the body, and its fix
// for i < A && B, where i < A is a bound and B calls nothing. It returns false, if no operand compares the
// counter.
func (l *loopCheck) compound(b *ast.BinaryExpr) (loopFinding, bool) {
	var ops []ast.Expr
	var flatten func(x ast.Expr)
	flatten = func(x ast.Expr) {
		if e, ok := ast.Unparen(x).(*ast.BinaryExpr); ok && e.Op == b.Op {
			flatten(e.X)
			flatten(e.Y)
			return
		}
		ops = append(ops, x)
	}
	flatten(b)

	k := slices.IndexFunc(ops, func(x ast.Expr) bool {
		_, _, msg := l.compare(x)
		return msg == ""
	})
	if k < 0 {
		return loopFinding{}, false
	}
	var rest []string
	for i, x := range ops {
		if i != k {
			rest = append(rest, l.text(x))
		}
	}
	i, cmp := l.counter.Name(), l.text(ops[k])
	if b.Op == token.LOR {
		// the loop runs, while one of the conditions holds, so it needs a bound of its own
		limit := i + " < limit"
		if l.step < 0 {
			limit = i + " > limit"
		}
		return loopFinding{
			msg: "the condition combines the comparison of the counter " + i + " with ||, so that the loop may run beyond the bound",
			fix: "give the loop a bound of its own and check the condition at the beginning of the body: for ...; " + limit + "; ... { if !(" + l.text(l.n.Cond) + ") { break }; ... }",
		}, true
	}
	restText := strings.Join(rest, " && ")
	r := loopFinding{
		msg: "the condition combines the comparison of the counter " + i + " with other conditions",
		fix: "compare only the counter in the condition and check the rest at the beginning of the body: for ...; " + cmp + "; ... { if !(" + restText + ") { break }; ... }",
	}
	if k != 0 {
		return r, true
	}
	op, bound, _ := l.compare(ops[0])
	if _, ok := l.fixedBound(op, bound); !ok {
		return r, true
	}
	for _, x := range ops[1:] {
		calls := false
		ast.Inspect(x, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				if !l.c.isConversion(n) {
					calls = true
				}
			case *ast.FuncLit:
				calls = true
			}
			return !calls
		})
		if calls {
			return r, true
		}
	}

	// i < A && B becomes i < A and if !(B) { break } at the beginning of the body; the rest keeps its source code
	cond := ast.Unparen(l.n.Cond)
	src := l.text(&ast.BadExpr{From: ops[1].Pos(), To: cond.End()})
	ind := l.indent(l.start()) + "\t"
	check := "\n" + ind + "if !(" + src + ") {\n" + ind + "\tbreak\n" + ind + "}"
	body := l.n.Body
	p := l.c.d.fset.Position
	if len(body.List) == 0 || p(body.List[0].Pos()).Line == p(body.Lbrace).Line {
		// the body starts on the line of {, which a statement must not follow on the line of }
		check += "\n" + ind
		if len(body.List) == 0 {
			check = check[:len(check)-1]
		}
	}
	r.edits = []loopEdit{
		{from: ops[0].End(), to: cond.End(), text: ""},
		{from: body.Lbrace + 1, to: body.Lbrace + 1, text: check},
	}
	return r, true
}
