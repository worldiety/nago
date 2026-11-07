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
	"strconv"
	"strings"
)

// The rule GS033 of this file reads the values of optional results only, where they are valid (wokoda
// R-SCRIPT-GS-OPTIONAL). A call with two or more results, whose last one is a bool or an error, pairs each result
// before it with that bool (ok) or error (err), and with both, if the last is an error and the one before it a bool,
// whoever the callee is and whatever the type is. A paired value is bound with := or var to a new variable and read
// only, where a guard after its binding proves ok true or err nil (guard.go), or where a fact proves, that Get of a
// list or a dict finds its element (facts.go). It may be passed on unread: return v, ok with its own ok, return v,
// false, return v, err with its own err, return v, e with an error, which is not nil, return g(), opt.OrZero(g())
// and opt.OrZeroErr(g()). An error may not be thrown away: _ in its place, a call as a statement, whose last result
// is an error, and an error, which is assigned again before it is read, are reported as well.

// resultPairing tells, how the results pair: values is the number of results before the partners, okIdx the index of
// the bool and errIdx the index of the error, each -1 without one. ok is false, if the results do not pair.
func resultPairing(t *types.Tuple) (values, okIdx, errIdx int, ok bool) {
	if t == nil || t.Len() < 2 {
		return 0, -1, -1, false
	}
	n := t.Len()
	last := t.At(n - 1).Type()
	switch {
	case isError(last):
		okIdx, errIdx = -1, n-1
		if n >= 3 && basicKind(t.At(n-2).Type()) == types.Bool {
			okIdx = n - 2
		}
	case basicKind(last) == types.Bool:
		okIdx, errIdx = n-1, -1
	default:
		return 0, -1, -1, false
	}
	values = n - 1
	if okIdx >= 0 && okIdx < values {
		values = okIdx
	}
	return values, okIdx, errIdx, true
}

// pairing is a call, whose results pair, see [resultPairing].
type pairing struct {
	call                  *ast.CallExpr
	tuple                 *types.Tuple
	values, okIdx, errIdx int
}

// pairingOf returns the pairing of the call e, if its results pair.
func (c *checker) pairingOf(e ast.Expr) (pairing, bool) {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return pairing{}, false
	}
	t, ok := c.info.TypeOf(call).(*types.Tuple)
	if !ok {
		return pairing{}, false
	}
	values, okIdx, errIdx, ok := resultPairing(t)
	return pairing{call: call, tuple: t, values: values, okIdx: okIdx, errIdx: errIdx}, ok
}

// pairInfo is the binding of a paired value to a new variable.
type pairInfo struct {
	// ok and err are the partners or nil without one.
	ok, err types.Object
	// bind is the end of the binding statement: a guard comes after it.
	bind token.Pos
	call *ast.CallExpr
	// get is the call, if it is Get of a list or a dict, whose element a fact may prove, and stack the stack of the
	// binding, at which the fact is proved.
	get   *getCall
	stack []ast.Node
	// the result of the fact, once it is proved
	proved    bool
	holds     bool
	factKill  *factKill
	valueType types.Type
}

func isBlank(e ast.Expr) bool {
	id, ok := ast.Unparen(e).(*ast.Ident)
	return ok && id.Name == "_"
}

// calleeText is the text of the function, which the call calls, for a message, e.g. l.Get or text.ParseInt.
func calleeText(call *ast.CallExpr) string {
	return types.ExprString(call.Fun)
}

// returnsError tells, whether the call is no conversion and its last result is an error.
func (c *checker) returnsError(call *ast.CallExpr) bool {
	if tv, ok := c.info.Types[call.Fun]; ok && (tv.IsType() || tv.IsBuiltin()) {
		return false
	}
	switch t := c.info.TypeOf(call).(type) {
	case *types.Tuple:
		return t.Len() > 0 && isError(t.At(t.Len()-1).Type())
	default:
		return isError(t)
	}
}

// checkBinding checks the binding of the results of a call: lhs = rhs, lhs := rhs or var lhs = rhs.
func (c *checker) checkBinding(n ast.Node, lhs, rhs []ast.Expr, define bool) {
	as, _ := n.(*ast.AssignStmt)
	if len(lhs) == len(rhs) {
		// a call with only an error
		for i, r := range rhs {
			call, ok := ast.Unparen(r).(*ast.CallExpr)
			if !ok || !c.returnsError(call) {
				continue
			}
			if _, tuple := c.info.TypeOf(call).(*types.Tuple); tuple {
				continue
			}
			if isBlank(lhs[i]) {
				c.reportDiscard(n, call)
			} else if id, ok := ast.Unparen(lhs[i]).(*ast.Ident); ok && as != nil {
				c.checkOverwrite(id)
			}
		}
		return
	}
	if len(rhs) != 1 {
		return
	}
	p, ok := c.pairingOf(rhs[0])
	if !ok || len(lhs) != p.tuple.Len() {
		return
	}

	okBlank := p.okIdx >= 0 && isBlank(lhs[p.okIdx])
	errBlank := p.errIdx >= 0 && isBlank(lhs[p.errIdx])
	var bound []int
	for j := range p.values {
		if !isBlank(lhs[j]) {
			bound = append(bound, j)
		}
	}
	if p.errIdx >= 0 && !errBlank && as != nil {
		if id, ok := ast.Unparen(lhs[p.errIdx]).(*ast.Ident); ok {
			c.checkOverwrite(id)
		}
	}
	if len(bound) == 0 {
		if errBlank {
			c.reportDiscard(n, p.call)
		}
		return
	}

	g, _ := c.getCallOf(p.call)
	if p.errIdx >= 0 {
		g = nil
	}

	// = and a redeclaration bind to a variable, whose guard the checker can not track
	for _, j := range bound {
		id, isID := ast.Unparen(lhs[j]).(*ast.Ident)
		if define && isID && c.info.Defs[id] != nil {
			continue
		}
		c.reportRebind(n, lhs[j], p, as, g, okBlank || errBlank)
		return
	}

	if okBlank || errBlank {
		var kill *factKill
		if !errBlank && g != nil {
			holds, k := c.factFor(c.stack, g)
			if holds {
				return
			}
			kill = k
		}
		c.reportBlank(n, lhs[bound[0]], p, as, g, errBlank, kill)
		return
	}

	var okObj, errObj types.Object
	partner := func(e ast.Expr) types.Object {
		id := ast.Unparen(e).(*ast.Ident)
		if obj := c.info.Defs[id]; obj != nil {
			return obj
		}
		return c.info.Uses[id]
	}
	if p.okIdx >= 0 {
		okObj = partner(lhs[p.okIdx])
	}
	if p.errIdx >= 0 {
		errObj = partner(lhs[p.errIdx])
	}
	for _, j := range bound {
		id := ast.Unparen(lhs[j]).(*ast.Ident)
		obj := c.info.Defs[id]
		if obj == nil {
			continue
		}
		info := &pairInfo{ok: okObj, err: errObj, bind: n.End(), call: p.call, valueType: p.tuple.At(j).Type()}
		if g != nil {
			info.get = g
			info.stack = append([]ast.Node(nil), c.stack...)
		}
		c.pairs[obj] = info
	}
}

// assigned tells, whether the identifier on top of the stack is assigned with =, which does not read it.
func (c *checker) assigned(id *ast.Ident) bool {
	if len(c.stack) < 2 {
		return false
	}
	as, ok := c.stack[len(c.stack)-2].(*ast.AssignStmt)
	if !ok || as.Tok != token.ASSIGN {
		return false
	}
	for _, l := range as.Lhs {
		if l == ast.Expr(id) {
			return true
		}
	}
	return false
}

// checkPaired checks the use of a paired value.
func (c *checker) checkPaired(id *ast.Ident, p *pairInfo) {
	if c.passedOn(id, p) {
		return
	}
	okFine := p.ok == nil || c.guardedAfter(c.truthy(p.ok), p.ok, p.bind)
	errFine := p.err == nil || c.guardedAfter(c.isNilErr(p.err), p.err, p.bind)
	if okFine && errFine {
		return
	}
	if p.get != nil && p.err == nil {
		if !p.proved {
			p.proved = true
			p.holds, p.factKill = c.factFor(p.stack, p.get)
		}
		if p.holds {
			return
		}
	}

	var unchecked []string
	if !okFine {
		unchecked = append(unchecked, p.ok.Name())
	}
	if !errFine {
		unchecked = append(unchecked, p.err.Name())
	}
	verb := " is"
	if len(unchecked) > 1 {
		verb = " are"
	}
	msg := id.Name + " may be invalid, because " + strings.Join(unchecked, " and ") + verb + " not checked"
	if p.factKill != nil {
		msg = id.Name + " may be invalid: " + p.factKill.text(c.d.fset)
	}
	var fix []string
	if p.ok != nil && !okFine {
		fix = append(fix, "use "+id.Name+" inside if "+p.ok.Name()+" { ... } or after if !"+p.ok.Name()+" { return ... }")
	}
	if p.err != nil && !errFine {
		fix = append(fix, "pass the error on with if "+p.err.Name()+" != nil { return ..., "+p.err.Name()+" } and use "+id.Name+" after it, or inside if "+p.err.Name()+" == nil { ... }")
	}
	fix = append(fix, c.defaultHint(p.call, p.get, p.err != nil))
	c.report(id, CodeGuard, msg, c.zeroWhy(p.call, p.valueType, p.ok, p.err), strings.Join(fix, "; "))
}

// guardedAfter tells, whether a guard after the position bind proves the property of the partner obj at the use
// on top of the stack, and obj keeps its value from bind up to the guard.
func (c *checker) guardedAfter(a atom, obj types.Object, bind token.Pos) bool {
	depth, from, ok := guardIn(c.stack, a, bind)
	return ok && c.keptUntil(obj, bind, from, depth)
}

// keptUntil tells, whether the partner obj keeps the value of its binding, which ends at bind, up to the position
// from, which the node stack[depth] contains: it is not assigned in between, not in a loop or a function literal,
// which contains the node, but not the binding, after the binding, and not in a function literal, which captures it.
func (c *checker) keptUntil(obj types.Object, bind, from token.Pos, depth int) bool {
	for _, a := range c.assigns[obj] {
		if a.captured {
			return false
		}
		if a.pos < bind {
			continue
		}
		if a.pos < from {
			return false
		}
		for _, n := range c.stack[:depth+1] {
			holdsBind := n.Pos() <= bind && bind <= n.End()
			switch n.(type) {
			case *ast.ForStmt, *ast.RangeStmt:
				if !holdsBind && n.Pos() <= a.pos && a.pos < n.End() {
					return false
				}
			case *ast.FuncLit:
				if !holdsBind {
					return false
				}
			}
		}
	}
	return true
}

// passedOn tells, whether the use of the paired value on top of the stack passes it on in a return, whose partners
// tell the caller, that it may be invalid.
func (c *checker) passedOn(id *ast.Ident, p *pairInfo) bool {
	if len(c.stack) < 2 {
		return false
	}
	ret, ok := c.stack[len(c.stack)-2].(*ast.ReturnStmt)
	if !ok {
		return false
	}
	idx := -1
	for i, r := range ret.Results {
		if r == ast.Expr(id) {
			idx = i
		}
	}
	sig := c.enclosingSignature()
	if idx < 0 || sig == nil || len(ret.Results) != sig.Results().Len() {
		return false
	}
	values, okIdx, errIdx, paired := resultPairing(sig.Results())
	if !paired || idx >= values {
		return false
	}
	if okIdx >= 0 && isFalseConst(c.info, ret.Results[okIdx]) {
		return true
	}
	if errIdx >= 0 && c.nonNilError(ret.Results[errIdx]) {
		return true
	}
	if p.ok == nil && p.err == nil {
		return false
	}
	use := id.Pos()
	if p.ok != nil && !(okIdx >= 0 && c.refersTo(ret.Results[okIdx], p.ok) && c.keptUntil(p.ok, p.bind, use, len(c.stack)-1)) {
		return false
	}
	if p.err != nil && !(errIdx >= 0 && c.refersTo(ret.Results[errIdx], p.err) && c.keptUntil(p.err, p.bind, use, len(c.stack)-1)) {
		return false
	}
	return true
}

func isFalseConst(info *types.Info, e ast.Expr) bool {
	tv, ok := info.Types[e]
	return ok && tv.Value != nil && tv.Value.Kind() == constant.Bool && !constant.BoolVal(tv.Value)
}

func isTrueConst(info *types.Info, e ast.Expr) bool {
	tv, ok := info.Types[e]
	return ok && tv.Value != nil && tv.Value.Kind() == constant.Bool && constant.BoolVal(tv.Value)
}

// nonNilError tells, whether the error e can not be nil: errors.New(...) or a variable inside if e != nil { ... }.
func (c *checker) nonNilError(e ast.Expr) bool {
	e = ast.Unparen(e)
	if call, ok := e.(*ast.CallExpr); ok {
		fn := c.namedCallee(call)
		return fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == "errors" && fn.Name() == "New"
	}
	id, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	obj := c.info.Uses[id]
	if obj == nil || !c.isLocal(obj) || !isError(obj.Type()) {
		return false
	}
	_, from, guarded := c.guardFor(c.nonNil(obj))
	return guarded && c.stable(obj, from)
}

// namedCallee returns the function, which the call calls by its name, or nil for a function value, a method of an
// interface or of a type parameter, a builtin and a conversion.
func (c *checker) namedCallee(call *ast.CallExpr) *types.Func {
	fun := ast.Unparen(call.Fun)
	switch f := fun.(type) {
	case *ast.IndexExpr:
		fun = f.X
	case *ast.IndexListExpr:
		fun = f.X
	}
	switch f := fun.(type) {
	case *ast.Ident:
		fn, _ := c.info.Uses[f].(*types.Func)
		return fn
	case *ast.SelectorExpr:
		if sel := c.info.Selections[f]; sel != nil {
			if sel.Kind() != types.MethodVal {
				return nil
			}
			recv := types.Unalias(derefType(sel.Recv()))
			if _, ok := recv.(*types.TypeParam); ok || isInterface(recv) {
				return nil
			}
			fn, _ := sel.Obj().(*types.Func)
			return fn
		}
		fn, _ := c.info.Uses[f.Sel].(*types.Func)
		return fn
	}
	return nil
}

// checkPairedArgs rejects f(g()), unless f is opt.OrZero or opt.OrZeroErr: the results of g would reach f unchecked.
func (c *checker) checkPairedArgs(call *ast.CallExpr) {
	if len(call.Args) != 1 {
		return
	}
	p, ok := c.pairingOf(call.Args[0])
	if !ok {
		return
	}
	if fn := c.namedCallee(call); fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == optPath && (fn.Name() == "OrZero" || fn.Name() == "OrZeroErr") {
		return
	}
	g := calleeText(p.call)
	what := "ok"
	if p.errIdx >= 0 {
		what = "the error"
	}
	fix := "check " + what + " first, e.g. if v, ok := " + types.ExprString(p.call) + "; ok { " + calleeText(call) + "(v, ok) }"
	if p.errIdx >= 0 {
		fix = "check the error first, e.g. if v, err := " + types.ExprString(p.call) + "; err == nil { ... }, or pass it on: if err != nil { return ..., err }"
	}
	fix += "; " + c.defaultHint(p.call, nil, p.errIdx >= 0)
	c.report(call, CodeGuard, "the results of "+g+" are passed to "+calleeText(call)+" without a check of "+what, c.zeroWhy(p.call, p.tuple.At(0).Type(), nil, nil), fix)
}

// checkCallStmt rejects a call as a statement, whose last result is an error: the error would be thrown away.
func (c *checker) checkCallStmt(s *ast.ExprStmt) {
	call, ok := ast.Unparen(s.X).(*ast.CallExpr)
	if !ok || !c.returnsError(call) {
		return
	}
	c.reportDiscard(s, call)
}

// reportDiscard reports a call, whose error is thrown away.
func (c *checker) reportDiscard(n ast.Node, call *ast.CallExpr) {
	text := types.ExprString(call)
	why := "a failed call would go unnoticed"
	fix := "pass the error on: if err := " + text + "; err != nil { return ..., err }"
	if t, tuple := c.info.TypeOf(call).(*types.Tuple); tuple {
		why += ", and its other results may be invalid then"
		fix = "pass the error on: if _, err := " + text + "; err != nil { return ..., err }"
		if t.Len() == 2 {
			fix += ", or write v := opt.OrZeroErr(" + text + ") to go on with the zero value deliberately"
		}
	}
	c.report(n, CodeGuard, "the error of "+calleeText(call)+" is thrown away", why, fix)
}

// reportRebind reports a paired value, which is bound with = or by a redeclaration to an existing variable.
func (c *checker) reportRebind(n ast.Node, target ast.Expr, p pairing, as *ast.AssignStmt, g *getCall, blank bool) {
	what := "ok"
	if p.errIdx >= 0 && p.okIdx < 0 {
		what = "err"
	}
	t := types.ExprString(target)
	fix := "write if t, " + what + " := " + types.ExprString(p.call) + "; " + what + " { " + t + " = t }"
	if p.errIdx >= 0 && p.okIdx < 0 {
		fix = "write if t, err := " + types.ExprString(p.call) + "; err == nil { " + t + " = t }, or pass the error on"
	}
	if blank {
		if text, ok := c.optionalFix(as, p, g); ok {
			fix = "write " + text + ", which takes the zero value, if " + calleeText(p.call) + " fails, or " + fix
		}
	}
	c.report(n, CodeGuard, "bind "+t+" with := to a new variable", "the result of "+calleeText(p.call)+" may be invalid, and the checker tracks the check of "+what+" only for a new variable", fix)
}

// reportBlank reports a paired value, whose ok or error is _.
func (c *checker) reportBlank(n ast.Node, value ast.Expr, p pairing, as *ast.AssignStmt, g *getCall, errBlank bool, kill *factKill) {
	v := types.ExprString(value)
	what := "ok"
	if errBlank {
		what = "the error"
	}
	msg := v + " may be invalid, because " + what + " is ignored"
	if kill != nil {
		msg = v + " may be invalid: " + kill.text(c.d.fset)
	}
	fix := "write " + v + ", ok := " + types.ExprString(p.call) + " and use " + v + " inside if ok { ... } or after if !ok { return ... }"
	switch {
	case p.okIdx >= 0 && p.errIdx >= 0:
		fix = "name ok and the error and check both, e.g. if err != nil || !ok { return ... }, before " + v + " is used"
	case errBlank:
		fix = "write " + v + ", err := " + types.ExprString(p.call) + " and pass the error on: if err != nil { return ..., err }"
	}
	if text, ok := c.optionalFix(as, p, g); ok {
		fix = "write " + text + ", which takes the zero value, if " + calleeText(p.call) + " fails, or " + fix
	} else {
		fix += "; " + c.defaultHint(p.call, g, p.errIdx >= 0)
	}
	c.report(n, CodeGuard, msg, c.zeroWhy(p.call, p.tuple.At(0).Type(), nil, nil), fix)
}

// zeroWhy is the reason, why a paired value of the type t may be invalid.
func (c *checker) zeroWhy(call *ast.CallExpr, t types.Type, ok, err types.Object) string {
	tn := types.TypeString(t, func(p *types.Package) string { return p.Name() })
	p, paired := c.pairingOf(call)
	cond := "its ok is true"
	switch {
	case paired && p.okIdx >= 0 && p.errIdx >= 0:
		cond = "its ok is true and its error is nil"
	case paired && p.errIdx >= 0:
		cond = "its error is nil"
	}
	why := "the value of " + tn + ", which " + calleeText(call) + " returns, is only valid, if " + cond
	switch g, isGet := c.getCallOf(call); {
	case isGet && g.kind == kindDict:
		why = calleeText(call) + " returns the zero value of " + tn + " and false, if " + types.ExprString(g.arg) + " is no key of " + types.ExprString(g.recv)
	case isGet:
		why = calleeText(call) + " returns the zero value of " + tn + " and false, if " + types.ExprString(g.arg) + " is no index of " + types.ExprString(g.recv)
	case paired && c.zeroOnFailure(p):
		why = calleeText(call) + " returns the zero value of " + tn + ", if it fails"
	}
	if nilable(t) {
		why += "; the zero value of " + tn + " contains nil"
	}
	return why
}

// defaultHint names the forms, which go on with a default: GetOr of a list or a dict, and opt.OrZero or
// opt.OrZeroErr for a type, whose zero value contains no nil.
func (c *checker) defaultHint(call *ast.CallExpr, g *getCall, isErr bool) string {
	if g == nil {
		g, _ = c.getCallOf(call)
	}
	t := types.Type(nil)
	p, ok := c.pairingOf(call)
	if ok {
		t = p.tuple.At(0).Type()
	}
	switch {
	case ok && p.tuple.Len() > 2:
		return "a call with more than two results has no default form, so check its results"
	case g != nil && !isErr && t != nil && nilable(t):
		return "or take a default with " + types.ExprString(g.recv) + ".GetOr(" + types.ExprString(g.arg) + ", def)"
	case g != nil && !isErr:
		return "or take a default with " + types.ExprString(g.recv) + ".GetOr(" + types.ExprString(g.arg) + ", def) or opt.OrZero(" + types.ExprString(call) + ")"
	case t != nil && nilable(t):
		return "the zero value contains nil, so the result must be checked"
	case isErr:
		return "or go on with the zero value deliberately: opt.OrZeroErr(" + types.ExprString(call) + ")"
	}
	return "or go on with the zero value deliberately: opt.OrZero(" + types.ExprString(call) + ")"
}

// checkOverwrite reports the error variable id of the assignment on top of the stack, if it is assigned again on a
// way before it is read: the error of the first call would go unnoticed.
func (c *checker) checkOverwrite(id *ast.Ident) {
	obj := c.info.Defs[id]
	if obj == nil {
		obj = c.info.Uses[id]
	}
	if obj == nil || !c.isLocal(obj) || !isError(obj.Type()) {
		return
	}
	pos, ok := c.overwritten(obj)
	if !ok {
		return
	}
	line := strconv.Itoa(c.d.fset.Position(pos).Line)
	c.report(id, CodeGuard, id.Name+" is assigned again at line "+line+", before it is checked", "the error of this call would go unnoticed", "check "+id.Name+" after each call, e.g. if "+id.Name+" != nil { return ..., "+id.Name+" }")
}

// The accesses of a variable on the ways through statements.
const (
	accNone  = 1 << iota // the statements end without an access
	accRead              // a way reads the variable first
	accWrite             // a way assigns it first
	accEnd               // a way leaves the function or the loop without an access
)

// overwritten tells, whether the error obj, which the statement on top of the stack assigns, is assigned again on a
// way after it, before it is read, and returns the position of that assignment.
func (c *checker) overwritten(obj types.Object) (token.Pos, bool) {
	c.writeAt = token.NoPos
	child := c.stack[len(c.stack)-1]
	for i := len(c.stack) - 2; i >= 0; i-- {
		var acc int
		switch p := c.stack[i].(type) {
		case *ast.BlockStmt:
			acc = c.listAccess(after(p.List, child), obj)
		case *ast.CaseClause:
			acc = c.listAccess(after(p.Body, child), obj)
			if acc&accNone != 0 && i >= 2 {
				// continue after the switch
				i -= 2
				child = c.stack[i]
				if acc&accWrite != 0 {
					return c.writeAt, true
				}
				continue
			}
		case *ast.IfStmt:
			switch child {
			case p.Init:
				if c.mentions(p.Cond, obj) {
					return token.NoPos, false
				}
				acc = c.listAccess(p.Body.List, obj)
				if p.Else != nil {
					acc |= c.stmtAccess(p.Else, obj)
				} else {
					acc |= accNone
				}
			default:
				acc = accNone
			}
		case *ast.LabeledStmt:
			acc = accNone
		default:
			return token.NoPos, false
		}
		if acc&accWrite != 0 {
			return c.writeAt, true
		}
		if acc&accNone == 0 {
			return token.NoPos, false
		}
		child = c.stack[i]
	}
	return token.NoPos, false
}

func after(list []ast.Stmt, child ast.Node) []ast.Stmt {
	for i, s := range list {
		if ast.Node(s) == child {
			return list[i+1:]
		}
	}
	return nil
}

func (c *checker) listAccess(list []ast.Stmt, obj types.Object) int {
	res := 0
	for _, s := range list {
		a := c.stmtAccess(s, obj)
		res |= a &^ accNone
		if a&accNone == 0 {
			return res
		}
	}
	return res | accNone
}

func (c *checker) stmtAccess(s ast.Stmt, obj types.Object) int {
	switch s := s.(type) {
	case *ast.AssignStmt:
		for _, r := range s.Rhs {
			if c.mentions(r, obj) {
				return accRead
			}
		}
		write := false
		for _, l := range s.Lhs {
			if id, ok := ast.Unparen(l).(*ast.Ident); ok && (c.info.Uses[id] == obj || c.info.Defs[id] == obj) {
				write = true
			} else if c.mentions(l, obj) {
				return accRead
			}
		}
		if write {
			if s.Tok != token.ASSIGN && s.Tok != token.DEFINE {
				return accRead
			}
			if !c.writeAt.IsValid() {
				c.writeAt = s.Pos()
			}
			return accWrite
		}
		return accNone
	case *ast.IfStmt:
		res := 0
		if s.Init != nil {
			a := c.stmtAccess(s.Init, obj)
			if a&accNone == 0 {
				return a
			}
			res = a &^ accNone
		}
		if c.mentions(s.Cond, obj) {
			return res | accRead
		}
		res |= c.listAccess(s.Body.List, obj)
		if s.Else != nil {
			res |= c.stmtAccess(s.Else, obj)
		} else {
			res |= accNone
		}
		return res
	case *ast.BlockStmt:
		return c.listAccess(s.List, obj)
	case *ast.ReturnStmt:
		if len(s.Results) == 0 || c.mentions(s, obj) {
			return accRead
		}
		return accEnd
	case *ast.BranchStmt:
		return accEnd
	case *ast.SwitchStmt:
		if c.mentions(s.Init, obj) || c.mentions(s.Tag, obj) {
			return accRead
		}
		res, hasDefault := 0, false
		for _, cl := range s.Body.List {
			cc := cl.(*ast.CaseClause)
			if cc.List == nil {
				hasDefault = true
			}
			for _, e := range cc.List {
				if c.mentions(e, obj) {
					return accRead
				}
			}
			res |= c.listAccess(cc.Body, obj)
		}
		if !hasDefault {
			res |= accNone
		}
		return res
	}
	if c.mentions(s, obj) {
		return accRead
	}
	return accNone
}

// mentions tells, whether the node refers to obj.
func (c *checker) mentions(n ast.Node, obj types.Object) bool {
	if n == nil || (n == ast.Node(nil)) {
		return false
	}
	found := false
	ast.Inspect(n, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok && c.info.Uses[id] == obj {
			found = true
		}
		return true
	})
	return found
}
