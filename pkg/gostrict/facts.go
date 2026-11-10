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
)

// A fact proves, that Get of a list.List, a list.Frozen or a dict.Map finds its element, so that its value may be
// read without ok (wokoda R-SCRIPT-GS-OPTIONAL). The facts are proved on the syntax tree like the guards of guard.go,
// within one function, from these sources:
//   - an index: a counting loop up to n := L.Len(), a range over n or L.Len(), a loop down from L.Len() - 1, a guard
//     i >= 0 && i < L.Len() and a range over L.All();
//   - a constant index c: L := list.Of with more than c elements and a guard on L.Len(), e.g. if L.Len() != 2
//     { return }; and L.Len() - 1 after L.Len() == 0 left the block;
//   - a key: a range over D.Keys() or D.All() and a guard D.Has(k).
//
// A list is a reference, so a fact dies by any Delete or Pop of a list.List, whose element type is compatible with
// the one of L, by a call of an own function, whose arguments or receiver may reach such a list, by each call of a
// function value, of a method of an interface or of a type parameter and each range over a function value, which no
// header returned, by a header only through the functions, which it calls (wokoda R-SCRIPT-GS-HEADER-LISTS), and by an
// assignment to L, to a field on its path, to the bound or to the index. A list.Frozen never shrinks. A call, in whose
// arguments, receiver or function the access lies, runs after it, so it kills it only, where everything in the whole
// body of a loop or of a literal kills, except the header, in whose literal the access lies. What is not recognized, is
// reported, never kept silent.

type containerKind int

const (
	kindList containerKind = iota
	kindFrozen
	kindDict
)

// getCall is a call of Get of a list.List, a list.Frozen or a dict.Map.
type getCall struct {
	call      *ast.CallExpr
	recv, arg ast.Expr
	kind      containerKind
	// elem is the type of the elements of a list or of the values of a dict.
	elem types.Type
}

// containerMethod returns the receiver, the kind, the name and the element type of a call of a method of a list or
// a dict.
func (c *checker) containerMethod(call *ast.CallExpr) (recv ast.Expr, kind containerKind, name string, elem types.Type, ok bool) {
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return nil, 0, "", nil, false
	}
	sel := c.info.Selections[se]
	if sel == nil || sel.Kind() != types.MethodVal {
		return nil, 0, "", nil, false
	}
	n, ok := types.Unalias(derefType(sel.Recv())).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return nil, 0, "", nil, false
	}
	if n.Obj().Pkg().Path() == binPath && (n.Obj().Name() == "Bytes" || n.Obj().Name() == "Buffer") {
		// a sequence of bytes, whose elements are uint32 like those of a list
		kind = kindFrozen
		if n.Obj().Name() == "Buffer" {
			kind = kindList
		}
		return se.X, kind, se.Sel.Name, types.Typ[types.Uint32], true
	}
	if n.TypeArgs().Len() == 0 {
		return nil, 0, "", nil, false
	}
	switch p, name := n.Obj().Pkg().Path(), n.Obj().Name(); {
	case p == listPath && name == "List":
		kind = kindList
	case p == listPath && name == "Frozen":
		kind = kindFrozen
	case p == dictPath && name == "Map":
		kind = kindDict
	default:
		return nil, 0, "", nil, false
	}
	elem = n.TypeArgs().At(0)
	if kind == kindDict {
		elem = n.TypeArgs().At(1)
	}
	return se.X, kind, se.Sel.Name, elem, true
}

// getCallOf returns the call, if it calls Get of a list or a dict.
func (c *checker) getCallOf(call *ast.CallExpr) (*getCall, bool) {
	recv, kind, name, elem, ok := c.containerMethod(call)
	if !ok || name != "Get" || len(call.Args) != 1 {
		return nil, false
	}
	return &getCall{call: call, recv: recv, arg: call.Args[0], kind: kind, elem: elem}, true
}

// objPath is a local variable or a path of fields of one, e.g. s.items.
type objPath struct {
	root   types.Object
	fields []types.Object
}

// pathOf returns the path of e, if it is a local variable or a field path of one without a pointer on the path.
func (c *checker) pathOf(e ast.Expr) (objPath, bool) {
	switch e := ast.Unparen(e).(type) {
	case *ast.Ident:
		obj := c.info.Uses[e]
		if obj == nil || !c.isLocal(obj) || isPointer(obj.Type()) {
			return objPath{}, false
		}
		return objPath{root: obj}, true
	case *ast.SelectorExpr:
		sel := c.info.Selections[e]
		if sel == nil || sel.Kind() != types.FieldVal || sel.Indirect() || isPointer(c.info.TypeOf(e.X)) {
			return objPath{}, false
		}
		base, ok := c.pathOf(e.X)
		if !ok {
			return objPath{}, false
		}
		fields := append(append([]types.Object(nil), base.fields...), sel.Obj())
		return objPath{root: base.root, fields: fields}, true
	}
	return objPath{}, false
}

func (p objPath) same(q objPath) bool {
	return len(p.fields) == len(q.fields) && p.prefixOf(q)
}

// prefixOf tells, whether p is q or a path, which q continues: an assignment to p changes q.
func (p objPath) prefixOf(q objPath) bool {
	if p.root != q.root || len(p.fields) > len(q.fields) {
		return false
	}
	for i, f := range p.fields {
		if q.fields[i] != f {
			return false
		}
	}
	return true
}

// samePath tells, whether e is the path p.
func (c *checker) samePath(e ast.Expr, p objPath) bool {
	q, ok := c.pathOf(e)
	return ok && q.same(p)
}

// lenOf tells, whether e is the call L.Len() of a list or a frozen list with the path p.
func (c *checker) lenOf(e ast.Expr, p objPath) bool {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	recv, kind, name, _, ok := c.containerMethod(call)
	return ok && name == "Len" && kind != kindDict && len(call.Args) == 0 && c.samePath(recv, p)
}

// factKill is a fact, which would hold, but dies at pos.
type factKill struct {
	src, pos token.Pos
	what     string
}

func (k *factKill) text(fset *token.FileSet) string {
	return "the fact of line " + strconv.Itoa(fset.Position(k.src).Line) + ", which would prove it, dies at line " + strconv.Itoa(fset.Position(k.pos).Line) + " by " + k.what
}

// event is a statement or an expression of a function, which may kill a fact.
type event struct {
	pos token.Pos
	// what describes it for a message
	what string
	// shrink is a Delete or Pop of a container of kind with the element type elem
	shrink bool
	kind   containerKind
	elem   types.Type
	// call is a call, which may change a list, which it reaches, and always tells, that the event kills each fact, e.g.
	// a range over a function value
	call   *ast.CallExpr
	always bool
	// path is an assigned field path
	path objPath
}

// funcEvents are the events of a function and the definitions of its local variables.
type funcEvents struct {
	shrinks, calls, fields []event
	// defs are the statements, which define a local variable with a single value, e.g. n := l.Len(), and the value
	defs map[types.Object]ast.Expr
	// defEnd are the ends of these statements
	defEnd map[types.Object]token.Pos
}

// events returns the events of the function declaration.
func (c *checker) events(fd *ast.FuncDecl) *funcEvents {
	if fe, ok := c.funcEvents[fd]; ok {
		return fe
	}
	fe := &funcEvents{defs: map[types.Object]ast.Expr{}, defEnd: map[types.Object]token.Pos{}}
	c.funcEvents[fd] = fe
	if fd.Body == nil {
		return fe
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if _, kind, name, elem, ok := c.containerMethod(n); ok {
				if (name == "Delete" && kind != kindFrozen) || (name == "Pop" && kind == kindList) || name == "Freeze" || name == "Reset" {
					fe.shrinks = append(fe.shrinks, event{pos: n.Pos(), what: "the call of " + types.ExprString(n.Fun), shrink: true, kind: kind, elem: elem})
				}
				return true
			}
			if c.mayChange(n) {
				fe.calls = append(fe.calls, event{pos: n.Pos(), what: "the call of " + types.ExprString(n.Fun), call: n})
			}
		case *ast.RangeStmt:
			// a range over a function value calls it, unless a header returned it, e.g. L.All(), which calls only the body
			if _, ok := c.info.TypeOf(n.X).Underlying().(*types.Signature); ok && !c.headerResult(n.X) {
				fe.calls = append(fe.calls, event{pos: n.Pos(), what: "the range over " + types.ExprString(n.X), always: true})
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE && len(n.Lhs) == len(n.Rhs) {
				for i, l := range n.Lhs {
					if id, ok := l.(*ast.Ident); ok {
						if obj := c.info.Defs[id]; obj != nil {
							fe.defs[obj] = n.Rhs[i]
							fe.defEnd[obj] = n.End()
						}
					}
				}
			}
			for _, l := range n.Lhs {
				if se, ok := ast.Unparen(l).(*ast.SelectorExpr); ok {
					if p, ok := c.pathOf(se); ok {
						fe.fields = append(fe.fields, event{pos: se.Pos(), what: "the assignment to " + types.ExprString(se), path: p})
					}
				}
			}
		}
		return true
	})
	return fe
}

// mayChange tells, whether a call may change a list or a dict: a call of an own function of the program (std/decimal and
// std/fuel included), of a function value and of a method of an interface or of a type parameter, and a call of a
// header or of the standard library, which kills by the functions, which it calls, see [checker.headerKills].
func (c *checker) mayChange(call *ast.CallExpr) bool {
	if tv, ok := c.info.Types[call.Fun]; ok && (tv.IsType() || tv.IsBuiltin()) {
		return false
	}
	fn := c.namedCallee(call)
	if fn == nil {
		return true
	}
	if fn.Pkg() == nil {
		return false
	}
	if p := c.prog.Package(fn.Pkg().Path()); p == nil || p.Kind != KindUser {
		return c.headerKills(call, fn)
	}
	return true
}

// headerKills tells, whether a call of a function of a header or of the standard library may change a list or a dict
// of the program: only through the functions, which it calls (R-SCRIPT-GS-HEADER-LISTS). With //gostrict:calls earlier
// always; with function parameters and now or without a directive through each argument of a function parameter, which
// is no function literal, whose body counts where it stands; never with later or concurrent and without function
// parameters.
func (c *checker) headerKills(call *ast.CallExpr, fn *types.Func) bool {
	calls := c.decls[fn.Origin()].calls
	if calls == CallsEarlier {
		return true
	}
	sig := fn.Signature()
	params := funcParams(sig)
	if len(params) == 0 || calls.afterReturn() {
		return false
	}
	for _, i := range params {
		args := call.Args
		switch {
		case sig.Variadic() && i == sig.Params().Len()-1 && i < len(args):
			args = args[i:]
		case i < len(args):
			args = args[i : i+1]
		default:
			continue
		}
		for _, a := range args {
			if _, lit := ast.Unparen(a).(*ast.FuncLit); !lit {
				return true
			}
		}
	}
	return false
}

// headerResult tells, whether e is the result of a call of a header or of the standard library, e.g. L.All().
func (c *checker) headerResult(e ast.Expr) bool {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := c.namedCallee(call)
	if fn == nil || fn.Pkg() == nil {
		return false
	}
	p := c.prog.Package(fn.Pkg().Path())
	return p == nil || p.Kind != KindUser
}

// callReaches tells, whether the event of [checker.mayChange] may reach a container of the kind with the element type
// elem: a call of an own function through its arguments and its receiver, any other one always.
func (c *checker) callReaches(e event, elem types.Type, dict bool) bool {
	if e.always || e.call == nil {
		return true
	}
	fn := c.namedCallee(e.call)
	if fn == nil {
		return true
	}
	if p := c.prog.Package(fn.Pkg().Path()); p == nil || p.Kind != KindUser {
		return true
	}
	if se, ok := ast.Unparen(e.call.Fun).(*ast.SelectorExpr); ok {
		if sel := c.info.Selections[se]; sel != nil && sel.Kind() == types.MethodVal && reaches(c.info.TypeOf(se.X), elem, dict, map[types.Type]bool{}) {
			return true
		}
	}
	for _, a := range e.call.Args {
		if reaches(c.info.TypeOf(a), elem, dict, map[types.Type]bool{}) {
			return true
		}
	}
	return false
}

// reaches tells, whether a value of t may reach a list.List with elements, which are compatible with elem, or, if
// dict is set, a dict.Map with such values. A function, an interface, a handle and a type parameter may reach
// anything; a big.Int is a number.
func reaches(t types.Type, elem types.Type, dict bool, seen map[types.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	if isValueHandle(t) {
		return false
	}
	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() != nil && n.TypeArgs().Len() > 0 {
		args := n.TypeArgs()
		switch p, name := n.Obj().Pkg().Path(), n.Obj().Name(); {
		case p == listPath && name == "List":
			return (!dict && compatible(args.At(0), elem)) || reaches(args.At(0), elem, dict, seen)
		case p == listPath && name == "Frozen", p == optPath && name == "Option", p == msgPath && name == "Out":
			return reaches(args.At(0), elem, dict, seen)
		case p == dictPath && name == "Map":
			return (dict && compatible(args.At(1), elem)) || reaches(args.At(0), elem, dict, seen) || reaches(args.At(1), elem, dict, seen)
		}
	}
	if isHandle(t) {
		return true
	}
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return true
	}
	switch u := t.Underlying().(type) {
	case *types.Signature, *types.Interface:
		return true
	case *types.Pointer:
		return reaches(u.Elem(), elem, dict, seen)
	case *types.Struct:
		for f := range u.Fields() {
			if reaches(f.Type(), elem, dict, seen) {
				return true
			}
		}
	}
	return false
}

// compatible tells, whether two element types may be the same: identical, or one contains a type parameter.
func compatible(a, b types.Type) bool {
	return types.Identical(a, b) || hasTypeParam(a) || hasTypeParam(b)
}

// span is a range of positions [from, to).
type span struct{ from, to token.Pos }

func inSpans(pos token.Pos, spans []span) bool {
	for _, s := range spans {
		if s.from <= pos && pos < s.to {
			return true
		}
	}
	return false
}

// loopSpans returns the loops of the stack after the index from, which do not contain the position src: a fact from
// src must hold in their whole body, since they run it again. A function literal, which a header calls now, e.g. the
// comparison of list.SortFunc, may run several times like the body of a loop.
func loopSpans(stack []ast.Node, from int, src token.Pos) []span {
	var res []span
	for _, n := range stack[from+1:] {
		switch n.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.FuncLit:
			if !(n.Pos() <= src && src < n.End()) {
				res = append(res, span{n.Pos(), n.End()})
			}
		}
	}
	return res
}

// litBetween tells, whether the stack holds a function literal after the index from, which is no argument of a header
// with //gostrict:calls now: a fact ends at such a literal, which may run later.
func (c *checker) litBetween(stack []ast.Node, from int) bool {
	for k := from + 1; k < len(stack); k++ {
		lit, ok := stack[k].(*ast.FuncLit)
		if !ok {
			continue
		}
		call, ok := stack[k-1].(*ast.CallExpr)
		if !ok || ast.Unparen(call.Fun) == ast.Expr(lit) {
			return true
		}
		fn := c.headerCall(call)
		if fn == nil || c.decls[fn.Origin()].calls != CallsNow {
			return true
		}
	}
	return false
}

// kept tells, whether the variable is not assigned in the spans and not in a function literal, which captures it, and
// returns the first such assignment otherwise.
func (c *checker) kept(obj types.Object, spans []span) (token.Pos, bool) {
	for _, a := range c.assigns[obj] {
		if a.captured || inSpans(a.pos, spans) {
			return a.pos, false
		}
	}
	return token.NoPos, true
}

// killed returns the first event, which kills the fact about the container of the Get: a shrink or a call in shrink,
// an assignment to the container or a field on its path in assign. A call, in whose arguments, receiver or function
// the access lies, runs after it, so it kills only in whole, the spans, in which everything in the whole body kills;
// except the header, in whose function literal the access lies, which counts where it stands.
func (f *factSearch) killed(shrink, whole, assign []span) (token.Pos, string, bool) {
	c, g, p := f.c, f.g, f.p
	if g.kind != kindFrozen {
		for _, e := range f.fe.shrinks {
			if inSpans(e.pos, shrink) && (e.kind == kindDict) == (g.kind == kindDict) && compatible(e.elem, g.elem) {
				return e.pos, e.what, true
			}
		}
		for _, e := range f.fe.calls {
			spans := shrink
			if e.call != nil && e.call.Pos() <= f.access && f.access < e.call.End() && !f.lits[e.call] {
				spans = whole
			}
			if inSpans(e.pos, spans) && c.callReaches(e, g.elem, g.kind == kindDict) {
				return e.pos, e.what, true
			}
		}
	}
	if pos, ok := c.kept(p.root, assign); !ok {
		return pos, "the assignment to " + p.root.Name(), true
	}
	for _, e := range f.fe.fields {
		if e.path.prefixOf(p) && inSpans(e.pos, assign) {
			return e.pos, e.what, true
		}
	}
	return token.NoPos, "", false
}

// factFor tells, whether a fact proves, that the Get g, which is a part of the node on top of the stack, finds its
// element. Otherwise it returns the kill of a fact, which would hold without it, if there is one.
func (c *checker) factFor(stack []ast.Node, g *getCall) (bool, *factKill) {
	fdIdx := -1
	for i := len(stack) - 1; i >= 0; i-- {
		if _, ok := stack[i].(*ast.FuncDecl); ok {
			fdIdx = i
			break
		}
	}
	if fdIdx < 0 {
		return false, nil
	}
	p, ok := c.pathOf(g.recv)
	if !ok {
		return false, nil
	}
	f := &factSearch{c: c, stack: stack, fdIdx: fdIdx, fe: c.events(stack[fdIdx].(*ast.FuncDecl)), g: g, p: p, access: g.call.Pos(), lits: map[*ast.CallExpr]bool{}}
	for k := fdIdx + 1; k < len(stack); k++ {
		if _, ok := stack[k].(*ast.FuncLit); ok {
			if call, ok := stack[k-1].(*ast.CallExpr); ok {
				f.lits[call] = true
			}
		}
	}
	if g.kind == kindDict {
		return f.keyFacts(), f.kill
	}

	arg := ast.Unparen(g.arg)
	if tv, ok := c.info.Types[arg]; ok && tv.Value != nil {
		v, exact := constant.Int64Val(constant.ToInt(tv.Value))
		return exact && v >= 0 && f.lenFacts(v+1), f.kill
	}
	if b, ok := arg.(*ast.BinaryExpr); ok && b.Op == token.SUB && c.lenOf(b.X, p) && isConstInt(c.info, b.Y, 1) {
		return f.lenFacts(1), f.kill
	}
	if id, ok := arg.(*ast.Ident); ok {
		if obj := c.info.Uses[id]; obj != nil && c.isLocal(obj) {
			return f.loopFacts(obj) || f.rangeGuard(obj), f.kill
		}
	}
	return false, nil
}

func isConstInt(info *types.Info, e ast.Expr, want int64) bool {
	tv, ok := info.Types[e]
	if !ok || tv.Value == nil {
		return false
	}
	v, exact := constant.Int64Val(constant.ToInt(tv.Value))
	return exact && v == want
}

// factSearch searches a fact for the Get g on the container p.
type factSearch struct {
	c      *checker
	stack  []ast.Node
	fdIdx  int
	fe     *funcEvents
	g      *getCall
	p      objPath
	access token.Pos
	// lits are the calls, in whose function literal the access lies
	lits map[*ast.CallExpr]bool
	// kill is the first kill of a fact, which would hold
	kill *factKill
}

// holds checks, that nothing kills a fact from src about the container in shrink (whole are the spans, in which
// everything in the whole body kills) and assign and that the variables keep their values in their spans; it records
// the kill otherwise.
func (f *factSearch) holds(src token.Pos, shrink, whole, assign []span, vars map[types.Object][]span) bool {
	if pos, what, dead := f.killed(shrink, whole, assign); dead {
		f.died(src, pos, what)
		return false
	}
	for obj, spans := range vars {
		if pos, ok := f.c.kept(obj, spans); !ok {
			f.died(src, pos, "the assignment to "+obj.Name())
			return false
		}
	}
	return true
}

func (f *factSearch) died(src, pos token.Pos, what string) {
	if f.kill == nil {
		f.kill = &factKill{src: src, pos: pos, what: what}
	}
}

// guardSpans returns the spans from src up to the access and of the loops after the index depth, which do not
// contain src.
func (f *factSearch) guardSpans(depth int, src token.Pos) []span {
	return append([]span{{src, f.access}}, loopSpans(f.stack, depth, src)...)
}

// loopFacts proves the index i by a loop, whose body holds the access.
func (f *factSearch) loopFacts(i types.Object) bool {
	c := f.c
	for k := len(f.stack) - 2; k > f.fdIdx; k-- {
		child := f.stack[k+1]
		switch n := f.stack[k].(type) {
		case *ast.FuncLit:
			if c.litBetween(f.stack, k-1) {
				return false
			}
		case *ast.ForStmt:
			if child != ast.Node(n.Body) {
				continue
			}
			if ok, done := f.countingLoop(k, n, i); done {
				return ok
			}
		case *ast.RangeStmt:
			if child != ast.Node(n.Body) || n.Tok != token.DEFINE || n.Key == nil {
				continue
			}
			key, ok := n.Key.(*ast.Ident)
			if !ok || c.info.Defs[key] != i {
				continue
			}
			return f.rangeLoop(k, n, i)
		}
	}
	return false
}

// countingLoop proves the index i of for i := a; i < n; i++ with n := L.Len() and for i := L.Len() - 1; i >= 0; i--.
// done tells, that the loop declares i.
func (f *factSearch) countingLoop(k int, n *ast.ForStmt, i types.Object) (ok, done bool) {
	c := f.c
	init, isAssign := n.Init.(*ast.AssignStmt)
	if !isAssign || init.Tok != token.DEFINE || len(init.Lhs) != 1 || len(init.Rhs) != 1 {
		return false, false
	}
	if id, isID := init.Lhs[0].(*ast.Ident); !isID || c.info.Defs[id] != i {
		return false, false
	}
	cond, isBin := ast.Unparen(n.Cond).(*ast.BinaryExpr)
	post, isInc := n.Post.(*ast.IncDecStmt)
	if !isBin || !isInc || !c.refersTo(cond.X, i) || !c.refersTo(post.X, i) {
		return false, true
	}
	body := []span{{n.Body.Pos(), n.Body.End()}}
	if f.c.litBetween(f.stack, k) {
		return false, true
	}

	// upwards from a constant a >= 0 up to a snapshot n := L.Len()
	if cond.Op == token.LSS && post.Tok == token.INC {
		tv, isConst := c.info.Types[init.Rhs[0]]
		if !isConst || tv.Value == nil || constant.Sign(tv.Value) < 0 {
			return false, true
		}
		bound, isID := ast.Unparen(cond.Y).(*ast.Ident)
		if !isID {
			return false, true
		}
		nObj := c.info.Uses[bound]
		def, isDef := f.fe.defs[nObj]
		if nObj == nil || !isDef || !c.lenOf(def, f.p) {
			return false, true
		}
		snap := f.fe.defEnd[nObj]
		spans := append([]span{{snap, n.End()}}, loopSpans(f.stack, f.fdIdx, snap)...)
		return f.holds(snap, spans, spans, spans, map[types.Object][]span{nObj: spans, i: body}), true
	}

	// downwards from L.Len() - 1 to 0
	if cond.Op == token.GEQ && post.Tok == token.DEC && isConstInt(c.info, cond.Y, 0) {
		b, isSub := ast.Unparen(init.Rhs[0]).(*ast.BinaryExpr)
		if !isSub || b.Op != token.SUB || !c.lenOf(b.X, f.p) || !isConstInt(c.info, b.Y, 1) {
			return false, true
		}
		snap := init.End()
		spans := append([]span{{snap, n.End()}}, loopSpans(f.stack, f.fdIdx, snap)...)
		return f.holds(snap, spans, spans, spans, map[types.Object][]span{i: body}), true
	}
	return false, true
}

// rangeLoop proves the index i of for i := range n with n := L.Len(), for i := range L.Len() and for i := range
// L.All().
func (f *factSearch) rangeLoop(k int, n *ast.RangeStmt, i types.Object) bool {
	c := f.c
	if c.litBetween(f.stack, k) {
		return false
	}
	x := ast.Unparen(n.X)
	if call, ok := x.(*ast.CallExpr); ok {
		if recv, kind, name, _, ok := c.containerMethod(call); ok && name == "All" && kind != kindDict && c.samePath(recv, f.p) {
			// the iteration yields an index of the list, as it is, when the body starts
			src := n.Body.Lbrace
			whole := loopSpans(f.stack, k, src)
			shrink := append([]span{{src, f.access}}, whole...)
			assign := []span{{n.Body.Pos(), n.Body.End()}}
			return f.holds(src, shrink, whole, assign, map[types.Object][]span{i: shrink})
		}
	}
	snap := token.NoPos
	var vars = map[types.Object][]span{}
	switch {
	case c.lenOf(x, f.p):
		snap = x.End()
	default:
		id, ok := x.(*ast.Ident)
		if !ok {
			return false
		}
		nObj := c.info.Uses[id]
		def, isDef := f.fe.defs[nObj]
		if nObj == nil || !isDef || !c.lenOf(def, f.p) {
			return false
		}
		snap = f.fe.defEnd[nObj]
		vars[nObj] = nil
	}
	spans := append([]span{{snap, n.End()}}, loopSpans(f.stack, f.fdIdx, snap)...)
	for obj := range vars {
		vars[obj] = spans
	}
	vars[i] = []span{{n.Body.Pos(), n.Body.End()}}
	return f.holds(snap, spans, spans, spans, vars)
}

// rangeGuard proves the index i by the guards i >= 0 and i < L.Len(), e.g. if i >= 0 && i < l.Len() { ... } or if
// i < 0 || i >= l.Len() { return }.
func (f *factSearch) rangeGuard(i types.Object) bool {
	c := f.c
	d1, from1, ok1 := guardIn(f.stack, c.cmpConst(i, func(op token.Token, v constant.Value) bool {
		switch op {
		case token.GEQ, token.EQL:
			return constant.Sign(v) >= 0
		case token.GTR:
			return constant.Compare(v, token.GEQ, constant.MakeInt64(-1))
		}
		return false
	}), token.NoPos)
	d2, from2, ok2 := guardIn(f.stack, c.belowLen(i, f.p), token.NoPos)
	if !ok1 || !ok2 {
		return false
	}
	depth, src := min(d1, d2), min(from1, from2)
	if c.litBetween(f.stack, depth) {
		return false
	}
	spans := f.guardSpans(d2, from2)
	return f.holds(from2, spans, loopSpans(f.stack, d2, from2), spans, map[types.Object][]span{i: f.guardSpans(depth, src)})
}

// lenFacts proves a length of at least min of the list: by list.Of with enough elements or by a guard on L.Len().
func (f *factSearch) lenFacts(least int64) bool {
	c := f.c
	if len(f.p.fields) == 0 {
		if def, ok := f.fe.defs[f.p.root]; ok {
			if call, ok := ast.Unparen(def).(*ast.CallExpr); ok && call.Ellipsis == token.NoPos && int64(len(call.Args)) >= least {
				if fn := c.namedCallee(call); fn != nil && fn.Pkg() != nil && fn.Pkg().Path() == listPath && fn.Name() == "Of" {
					src := f.fe.defEnd[f.p.root]
					lit := false
					for k := f.fdIdx + 1; k < len(f.stack); k++ {
						if l, ok := f.stack[k].(*ast.FuncLit); ok && !(l.Pos() <= src && src < l.End()) && c.litBetween(f.stack[:k+1], k-1) {
							lit = true
						}
					}
					if !lit {
						spans := f.guardSpans(f.fdIdx, src)
						if f.holds(src, spans, loopSpans(f.stack, f.fdIdx, src), spans, nil) {
							return true
						}
					}
				}
			}
		}
	}
	depth, from, ok := guardIn(f.stack, c.lenAtLeast(f.p, least), token.NoPos)
	if !ok || c.litBetween(f.stack, depth) {
		return false
	}
	spans := f.guardSpans(depth, from)
	return f.holds(from, spans, loopSpans(f.stack, depth, from), spans, nil)
}

// keyFacts proves the key of D.Get(k) by a range over D.Keys() or D.All() and by a guard D.Has(k).
func (f *factSearch) keyFacts() bool {
	c := f.c
	arg := ast.Unparen(f.g.arg)
	var key types.Object
	if id, ok := arg.(*ast.Ident); ok {
		key = c.info.Uses[id]
		if key == nil || !c.isLocal(key) {
			key = nil
		}
	}
	if key != nil {
		for k := len(f.stack) - 2; k > f.fdIdx; k-- {
			n, ok := f.stack[k].(*ast.RangeStmt)
			if !ok || f.stack[k+1] != ast.Node(n.Body) || n.Tok != token.DEFINE || n.Key == nil {
				continue
			}
			id, ok := n.Key.(*ast.Ident)
			if !ok || c.info.Defs[id] != key {
				continue
			}
			call, ok := ast.Unparen(n.X).(*ast.CallExpr)
			if !ok {
				return false
			}
			recv, kind, name, _, ok := c.containerMethod(call)
			if !ok || kind != kindDict || (name != "Keys" && name != "All") || !c.samePath(recv, f.p) || c.litBetween(f.stack, k) {
				return false
			}
			src := n.Body.Lbrace
			whole := loopSpans(f.stack, k, src)
			shrink := append([]span{{src, f.access}}, whole...)
			return f.holds(src, shrink, whole, []span{{n.Body.Pos(), n.Body.End()}}, map[types.Object][]span{key: shrink})
		}
	}
	depth, from, ok := guardIn(f.stack, c.hasKey(f.p, arg, key), token.NoPos)
	if !ok || c.litBetween(f.stack, depth) {
		return false
	}
	spans := f.guardSpans(depth, from)
	vars := map[types.Object][]span{}
	if key != nil {
		vars[key] = spans
	}
	return f.holds(from, spans, loopSpans(f.stack, depth, from), spans, vars)
}

// cmpConst is the property, that the comparison of obj with a constant, as obj op constant, satisfies want.
func (c *checker) cmpConst(obj types.Object, want func(op token.Token, v constant.Value) bool) atom {
	return func(cond ast.Expr, positive bool) bool {
		b, ok := cond.(*ast.BinaryExpr)
		if !ok {
			return false
		}
		op, other, ok := orient(b, func(e ast.Expr) bool { return c.refersTo(e, obj) })
		if !ok {
			return false
		}
		tv, ok := c.info.Types[other]
		if !ok || tv.Value == nil || tv.Value.Kind() != constant.Int {
			return false
		}
		if !positive {
			op = negate(op)
		}
		return want(op, tv.Value)
	}
}

// belowLen is the property i < L.Len().
func (c *checker) belowLen(i types.Object, p objPath) atom {
	return func(cond ast.Expr, positive bool) bool {
		b, ok := cond.(*ast.BinaryExpr)
		if !ok {
			return false
		}
		op, other, ok := orient(b, func(e ast.Expr) bool { return c.refersTo(e, i) })
		if !ok || !c.lenOf(other, p) {
			return false
		}
		if !positive {
			op = negate(op)
		}
		return op == token.LSS
	}
}

// lenAtLeast is the property L.Len() >= least.
func (c *checker) lenAtLeast(p objPath, least int64) atom {
	return func(cond ast.Expr, positive bool) bool {
		b, ok := cond.(*ast.BinaryExpr)
		if !ok {
			return false
		}
		op, other, ok := orient(b, func(e ast.Expr) bool { return c.lenOf(e, p) })
		if !ok {
			return false
		}
		tv, ok := c.info.Types[other]
		if !ok || tv.Value == nil {
			return false
		}
		k, exact := constant.Int64Val(constant.ToInt(tv.Value))
		if !exact {
			return false
		}
		if !positive {
			op = negate(op)
		}
		switch op {
		case token.EQL, token.GEQ:
			return k >= least
		case token.GTR:
			return k+1 >= least
		case token.NEQ:
			return k == 0 && least <= 1
		}
		return false
	}
}

// hasKey is the property D.Has(k) for the key arg, which is the variable key or a constant.
func (c *checker) hasKey(p objPath, arg ast.Expr, key types.Object) atom {
	return func(cond ast.Expr, positive bool) bool {
		call, ok := cond.(*ast.CallExpr)
		if !ok || !positive || len(call.Args) != 1 {
			return false
		}
		recv, kind, name, _, ok := c.containerMethod(call)
		if !ok || kind != kindDict || name != "Has" || !c.samePath(recv, p) {
			return false
		}
		if key != nil {
			return c.refersTo(call.Args[0], key)
		}
		a, b := c.info.Types[call.Args[0]], c.info.Types[arg]
		return a.Value != nil && b.Value != nil && constant.Compare(a.Value, token.EQL, b.Value)
	}
}

// orient returns the operator of the comparison as subject op other, where is tells the subject.
func orient(b *ast.BinaryExpr, is func(ast.Expr) bool) (token.Token, ast.Expr, bool) {
	switch {
	case is(b.X):
		return b.Op, b.Y, true
	case is(b.Y):
		mirror := map[token.Token]token.Token{token.LSS: token.GTR, token.GTR: token.LSS, token.LEQ: token.GEQ, token.GEQ: token.LEQ, token.EQL: token.EQL, token.NEQ: token.NEQ}
		op, ok := mirror[b.Op]
		return op, b.X, ok
	}
	return 0, nil, false
}

func negate(op token.Token) token.Token {
	return map[token.Token]token.Token{token.LSS: token.GEQ, token.GTR: token.LEQ, token.LEQ: token.GTR, token.GEQ: token.LSS, token.EQL: token.NEQ, token.NEQ: token.EQL}[op]
}
