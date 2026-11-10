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
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"slices"
	"strconv"
	"strings"
)

// The terminating profile proves, that every call of a package ends: its loops are bounded and it does not
// recurse, unless an argument decreases. A package opts in with the directive //gostrict:terminating in its
// documentation, or a host requires it of all packages with [Program.Terminating], e.g. for the scripts of an
// application, which run on its server.

const terminatingDirective = "//gostrict:terminating"

const fuelPath = "go.wdy.de/nago/pkg/gostrict/std/fuel"

// terminating tells, whether the profile applies to the package.
func (p *Program) terminating(pkg *Package) bool {
	if p.Terminating {
		return true
	}
	for _, f := range pkg.Files {
		if f.Doc == nil {
			continue
		}
		for _, c := range f.Doc.List {
			if strings.TrimSpace(c.Text) == terminatingDirective {
				return true
			}
		}
	}
	return false
}

// loopWhy is the reason of a report of a loop.
const loopWhy = "the terminating profile requires, that every loop ends"

// loopFix is the fix of a loop, whose reason has no rewrite of its own.
const loopFix = "count with a bound, e.g. for i := int32(0); i < n; i++, range over a list or a number, or limit the loop with fuel: f := fuel.New(1000); for f.Burn() && cond { ... }"

// checkLoops reports the loops of the package, whose end the compiler can not prove. Every range ends: a range over
// a list visits the elements, which the list has, when the range starts. The report of an inclusive bound, an
// expression as the bound and a compound condition names the rewrite of the loop, and the edits of a mechanical fix,
// where it keeps the meaning of the loop (R-SCRIPT-GS-LOOP-FIX).
func (c *checker) checkLoops(f *ast.File) {
	facts := c.loopFacts(f)
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ForStmt:
			l := &loopCheck{c: c, n: n, facts: facts}
			r, ok := l.bounded()
			if ok {
				break
			}
			fix := r.fix
			if fix == "" {
				fix = loopFix
			}
			p := c.d.fset.Position(n.Pos())
			if c.d.seen[fmt.Sprintf("%s:%d:%s", p.Filename, p.Line, CodeTermination)] {
				// a loop on the line is reported already, whose edits must not mix with these
				break
			}
			c.report(n, CodeTermination, r.msg, loopWhy, fix)
			for _, e := range r.edits {
				c.d.fix(CodeTermination, n.Pos(), e.from, e.to, e.text)
			}
		}
		return true
	})
}

// loopFinding tells, why a loop may not end, how to rewrite it, or "" for the general fix, and the edits of its
// mechanical fix, if the rewrite keeps the meaning of the loop.
type loopFinding struct {
	msg, fix string
	edits    []loopEdit
}

// loopEdit replaces the source code from up to to by text.
type loopEdit struct {
	from, to token.Pos
	text     string
}

// loopCheck checks a single for loop.
type loopCheck struct {
	c     *checker
	n     *ast.ForStmt
	facts *loopFacts
	// counter is the counter of the post statement and step its step, e.g. 1 for i++ or -2 for i -= 2
	counter *types.Var
	step    int64
}

// bounded tells, whether the loop ends, or returns, why it may not end.
func (l *loopCheck) bounded() (loopFinding, bool) {
	c, n := l.c, l.n
	if n.Cond == nil {
		return loopFinding{msg: "the loop has no condition"}, false
	}
	if c.burnsFuel(n.Cond) {
		return loopFinding{}, true
	}

	// for i := a; i < b; i++ with a counter, which only the post statement changes, and a bound, which does not
	// change
	l.counter, l.step = c.loopStep(n.Post)
	if l.counter == nil {
		return loopFinding{msg: "the loop has no counter, which its post statement increments or decrements by a constant"}, false
	}
	if c.assignedIn(l.counter, n.Body) {
		return loopFinding{msg: "the body assigns the counter " + l.counter.Name()}, false
	}
	if l.changedByLiteral(l.counter) {
		return loopFinding{msg: "a function literal, which the loop may call, assigns the counter " + l.counter.Name()}, false
	}
	if b, ok := ast.Unparen(n.Cond).(*ast.BinaryExpr); ok && (b.Op == token.LAND || b.Op == token.LOR) {
		if r, ok := l.compound(b); ok {
			return r, false
		}
	}
	op, bound, msg := l.compare(n.Cond)
	if msg != "" {
		return loopFinding{msg: msg}, false
	}
	return l.fixedBound(op, bound)
}

// compare returns the operator and the bound of the condition x, which compares the counter, as in i < bound, or
// why x is no comparison, which stops the counter.
func (l *loopCheck) compare(x ast.Expr) (token.Token, ast.Expr, string) {
	name := l.counter.Name()
	b, ok := ast.Unparen(x).(*ast.BinaryExpr)
	if !ok {
		return 0, nil, "the condition does not compare the counter " + name + " with a bound"
	}
	op, bound := b.Op, b.Y
	if l.c.isObj(b.Y, l.counter) {
		// b > i is i < b
		op, bound = mirror(op), b.X
	} else if !l.c.isObj(b.X, l.counter) {
		return 0, nil, "the condition does not compare the counter " + name + " with a bound"
	}
	switch {
	case l.step > 0 && (op == token.LSS || op == token.LEQ), l.step < 0 && (op == token.GTR || op == token.GEQ):
	default:
		return 0, nil, "the condition does not stop the counter " + name + " in its direction"
	}
	return op, bound, ""
}

func mirror(op token.Token) token.Token {
	switch op {
	case token.LSS:
		return token.GTR
	case token.LEQ:
		return token.GEQ
	case token.GTR:
		return token.LSS
	case token.GEQ:
		return token.LEQ
	}
	return op
}

// fixedBound tells, whether the bound of i op bound does not change during the loop and does not let the counter
// wrap around, or returns, why it may.
func (l *loopCheck) fixedBound(op token.Token, bound ast.Expr) (loopFinding, bool) {
	c := l.c
	if tv, ok := c.info.Types[bound]; ok && tv.Value != nil {
		if l.wraps(tv.Value, op) {
			return loopFinding{msg: "the counter wraps around at the bound"}, false
		}
		return loopFinding{}, true
	}
	if op == token.LEQ || op == token.GEQ {
		// no mechanical fix: whether bound + 1 wraps around, is not known
		end, b, i := "end", l.operand(bound), l.counter.Name()
		fix := "put the bound into a variable before the loop and compare without =: " + end + " := " + b + " + 1 and " + i + " < " + end + ", if " + b + " + 1 can not wrap around"
		if l.step < 0 {
			fix = "put the bound into a variable before the loop and compare without =: " + end + " := " + b + " - 1 and " + i + " > " + end + ", if " + b + " - 1 can not wrap around"
		}
		return loopFinding{msg: "an inclusive bound, which is not a constant, may let the counter wrap around", fix: fix}, false
	}
	id, ok := ast.Unparen(bound).(*ast.Ident)
	if !ok {
		return l.exprBound(op, bound), false
	}
	v, ok := c.info.Uses[id].(*types.Var)
	if !ok || c.assignedIn(v, l.n.Body) {
		return loopFinding{msg: "the body assigns the bound " + id.Name}, false
	}
	if l.changedByLiteral(v) {
		return loopFinding{msg: "a function literal, which the loop may call, assigns the bound " + id.Name}, false
	}
	return loopFinding{}, true
}

// wraps tells, whether the counter may wrap around at the constant bound v of i op v: its last value, which passes
// the condition, plus the step leaves the range of its type.
func (l *loopCheck) wraps(v constant.Value, op token.Token) bool {
	b, exact := constant.Int64Val(constant.ToInt(v))
	if !exact {
		return true
	}
	lo, hi := int64(math.MinInt32), int64(math.MaxInt32)
	switch basicKind(l.counter.Type()) {
	case types.Uint32:
		lo, hi = 0, math.MaxUint32
	case types.Int64:
		lo, hi = math.MinInt64, math.MaxInt64
	case types.Uint64:
		// a bound beyond this has no exact int64 and is reported above
		lo, hi = 0, math.MaxInt64
	}
	s := l.step
	if s < 0 {
		s = -s
	}
	switch op {
	case token.LSS:
		return b > hi-s+1
	case token.LEQ:
		return b > hi-s
	case token.GTR:
		return b < lo+s-1
	case token.GEQ:
		return b < lo+s
	}
	return false
}

// loopStep returns the counter and the step of the post statement i++, i--, i += c or i -= c, whose constant c is
// positive.
func (c *checker) loopStep(post ast.Stmt) (*types.Var, int64) {
	var x ast.Expr
	var step int64
	switch s := post.(type) {
	case *ast.IncDecStmt:
		x = s.X
		step = 1
		if s.Tok == token.DEC {
			step = -1
		}
	case *ast.AssignStmt:
		if len(s.Lhs) != 1 || (s.Tok != token.ADD_ASSIGN && s.Tok != token.SUB_ASSIGN) {
			return nil, 0
		}
		tv, ok := c.info.Types[s.Rhs[0]]
		if !ok || tv.Value == nil || constant.Sign(tv.Value) <= 0 {
			return nil, 0
		}
		x = s.Lhs[0]
		step, ok = constant.Int64Val(constant.ToInt(tv.Value))
		if !ok || step > math.MaxUint32 {
			// beyond the range of every counter, so that it wraps around at any bound
			step = math.MaxUint32 + 1
		}
		if s.Tok == token.SUB_ASSIGN {
			step = -step
		}
	default:
		return nil, 0
	}
	id, ok := ast.Unparen(x).(*ast.Ident)
	if !ok || !isInteger(c.info.TypeOf(id)) {
		return nil, 0
	}
	v, _ := c.info.Uses[id].(*types.Var)
	return v, step
}

func (c *checker) isObj(x ast.Expr, v *types.Var) bool {
	id, ok := ast.Unparen(x).(*ast.Ident)
	return ok && c.info.Uses[id] == v
}

// assignedIn tells, whether the variable is assigned within the node.
func (c *checker) assignedIn(v types.Object, n ast.Node) bool {
	for _, a := range c.assigns[v] {
		if a.pos >= n.Pos() && a.pos < n.End() {
			return true
		}
	}
	return false
}

// burnsFuel tells, whether the condition is f.Burn() of package fuel or a conjunction with it.
func (c *checker) burnsFuel(cond ast.Expr) bool {
	switch e := ast.Unparen(cond).(type) {
	case *ast.BinaryExpr:
		return e.Op == token.LAND && (c.burnsFuel(e.X) || c.burnsFuel(e.Y))
	case *ast.CallExpr:
		se, ok := ast.Unparen(e.Fun).(*ast.SelectorExpr)
		if !ok {
			return false
		}
		sel := c.info.Selections[se]
		return sel != nil && sel.Obj().Pkg() != nil && sel.Obj().Pkg().Path() == fuelPath && sel.Obj().Name() == "Burn"
	}
	return false
}

// callNode is a node of the call graph: a function or method, a function literal, the body of a range over a
// function or the initializer of a package variable.
type callNode struct {
	fn   *types.Func
	lit  *ast.FuncLit
	body *ast.RangeStmt
	v    *types.Var
}

func (n callNode) pos() token.Pos {
	switch {
	case n.fn != nil:
		return n.fn.Pos()
	case n.lit != nil:
		return n.lit.Pos()
	case n.body != nil:
		return n.body.Pos()
	}
	return n.v.Pos()
}

// name returns the name of the node in a [CycleLink].
func (n callNode) name() string {
	switch {
	case n.fn != nil:
		return funcName(n.fn)
	case n.lit != nil:
		return LinkLiteral
	case n.body != nil:
		return LinkRangeBody
	}
	return n.v.Name()
}

// marked returns the range of the report at the node: the name of a function, func of a literal or for of a loop.
func (n callNode) marked() (token.Pos, token.Pos) {
	pos := n.pos()
	switch {
	case n.fn != nil:
		return pos, pos + token.Pos(len(n.fn.Name()))
	case n.lit != nil:
		return pos, pos + token.Pos(len("func"))
	case n.body != nil:
		return pos, pos + token.Pos(len("for"))
	}
	return pos, pos + token.Pos(len(n.v.Name()))
}

// callEdge is a call of a node, which may reach another node.
type callEdge struct {
	to   callNode
	kind CallKind
	pos  token.Pos
	expr string
	// bySig tells, that the edge was resolved by the signature of a function value, and arity, that only the
	// number of its parameters and results fit
	bySig, arity bool
}

// valueFunc is a node, which a call of a function value may reach, with its signatures, e.g. of the instances of a
// generic function, which are used as values.
type valueFunc struct {
	node callNode
	sigs []*types.Signature
}

type methodKey struct {
	group *Package
	name  string
}

// arityKey groups the signatures, which may fit to each other.
type arityKey struct {
	group           *Package
	params, results int
	variadic        bool
}

func arityOf(group *Package, sig *types.Signature) arityKey {
	return arityKey{group: group, params: sig.Params().Len(), results: sig.Results().Len(), variadic: sig.Variadic()}
}

type edgeKey struct {
	from, to callNode
	pos      token.Pos
}

// recGraph is the call graph of the check of recursion.
type recGraph struct {
	prog  *Program
	decls map[*types.Func]funcSrc
	group map[*Package]*Package
	edges map[callNode][]callEdge
	seen  map[edgeKey]bool
	pkgOf map[callNode]*Package
	// values are the nodes, which a call of a function value may reach, by their import group and arity, and
	// valueIndex their index in values
	values     map[arityKey][]valueFunc
	valueIndex map[arityKey]map[callNode]int
	// methods are the methods, which a call of a method of an interface may reach, by import group and name
	methods map[methodKey][]callNode
	// headerSigs are the signatures of the function values of header packages in an import group, which call their
	// function arguments, e.g. list.SortFunc used as a value or the iterator, which l.All() returns
	headerSigs map[*Package][]*types.Signature
	// resolve contains the calls of function values, which are resolved, when all values are known
	resolve    []func()
	decreasing map[*types.Func]bool
}

// checkRecursion reports the functions of the packages of the profile, which may call themselves. The call graph
// contains the calls of all packages of the program in the subset, see [CallKind]: static calls, calls of function
// values and of type parameters, whose core type is a function, ranges over functions, whose bodies are nodes of
// their own, the initializers of package variables, calls of methods of interfaces and method values of interfaces
// and type parameters and the function arguments of functions of header packages and of the standard library,
// which call them, before they return, see the directive //gostrict:calls. A call of a function value reaches the
// functions, which are used as values, the function literals and the bodies of ranges over functions, whose
// signature fits, a call of a method of an interface each method of its name, but only of the packages, which are
// connected to the caller by imports: a value reaches another package only as an argument or a result of a call
// along an import. So the result of a package does not depend on the packages of a whole program, which are not
// connected to it.
func checkRecursion(prog *Program, d *diagnoser, decls map[*types.Func]funcSrc) {
	inProfile := map[*Package]bool{}
	for _, p := range prog.Packages {
		if p.Kind == KindUser && prog.terminating(p) {
			inProfile[p] = true
		}
	}
	if len(inProfile) == 0 {
		return
	}

	g := &recGraph{
		prog:       prog,
		decls:      decls,
		group:      importGroups(prog),
		edges:      map[callNode][]callEdge{},
		seen:       map[edgeKey]bool{},
		pkgOf:      map[callNode]*Package{},
		values:     map[arityKey][]valueFunc{},
		valueIndex: map[arityKey]map[callNode]int{},
		methods:    map[methodKey][]callNode{},
		headerSigs: map[*Package][]*types.Signature{},
		decreasing: map[*types.Func]bool{},
	}
	var users []*Package
	for _, p := range prog.Packages {
		if p.Kind == KindUser && p.Info != nil {
			users = append(users, p)
		}
	}
	// the methods, which a call of a method of an interface may reach, are approximated by all methods of its name
	for _, p := range users {
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv != nil && fd.Body != nil {
					if fn, ok := p.Info.Defs[fd.Name].(*types.Func); ok {
						k := methodKey{group: g.group[p], name: fn.Name()}
						g.methods[k] = append(g.methods[k], callNode{fn: fn})
					}
				}
			}
		}
	}
	for _, p := range users {
		g.walkPackage(p)
	}
	for _, r := range g.resolve {
		r()
	}
	g.report(d, inProfile)
}

// walkPackage adds the calls of the functions and of the initializers of the package variables of the package.
func (g *recGraph) walkPackage(p *Package) {
	for _, f := range p.Files {
		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				fn, ok := p.Info.Defs[decl.Name].(*types.Func)
				if !ok || decl.Body == nil {
					continue
				}
				if decreasingRecursion(p.Info, decl, fn) {
					g.decreasing[fn] = true
				}
				g.walk(p, callNode{fn: fn}, decl.Body)
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				for _, spec := range decl.Specs {
					vs := spec.(*ast.ValueSpec)
					for i, name := range vs.Names {
						v, ok := p.Info.Defs[name].(*types.Var)
						if !ok {
							continue
						}
						// var a, b = f() calls f once, which belongs to a
						var values []ast.Expr
						switch {
						case len(vs.Values) == len(vs.Names):
							values = vs.Values[i : i+1]
						case i == 0:
							values = vs.Values
						}
						for _, e := range values {
							g.walk(p, callNode{v: v}, e)
						}
					}
				}
			}
		}
	}
}

func (g *recGraph) walk(p *Package, node callNode, n ast.Node) {
	g.pkgOf[node] = p
	w := &recWalker{g: g, p: p, info: p.Info, grp: g.group[p], calls: callIdents(n), stack: []callNode{node}}
	ast.Inspect(n, w.visit)
}

func (g *recGraph) edge(from callNode, e callEdge) {
	k := edgeKey{from: from, to: e.to, pos: e.pos}
	if g.seen[k] {
		return
	}
	g.seen[k] = true
	g.edges[from] = append(g.edges[from], e)
}

func (g *recGraph) addValue(group *Package, node callNode, sig *types.Signature) {
	if sig == nil {
		return
	}
	k := arityOf(group, sig)
	index := g.valueIndex[k]
	if index == nil {
		index = map[callNode]int{}
		g.valueIndex[k] = index
	}
	i, ok := index[node]
	if !ok {
		index[node] = len(g.values[k])
		g.values[k] = append(g.values[k], valueFunc{node: node, sigs: []*types.Signature{sig}})
		return
	}
	v := &g.values[k][i]
	for _, s := range v.sigs {
		if types.Identical(s, sig) {
			return
		}
	}
	v.sigs = append(v.sigs, sig)
}

// isUser tells, whether the function belongs to a package in the subset.
func (g *recGraph) isUser(fn *types.Func) bool {
	src, ok := g.decls[fn.Origin()]
	return ok && src.pkg.Kind == KindUser
}

// callsNow tells, whether the function of a header package or of the standard library calls its function arguments
// during the call: unless its directive is //gostrict:calls later or concurrent.
func (g *recGraph) callsNow(fn *types.Func) bool {
	return !g.decls[fn.Origin()].calls.afterReturn()
}

// useHeaderValue notes the function of a header package or of the standard library, which the import group uses as
// a value or calls with the signature: a call of a function value with its signature calls the function arguments,
// and its results of a function type, e.g. an iterator, call theirs.
func (g *recGraph) useHeaderValue(group *Package, fn *types.Func, sig *types.Signature, used bool) {
	if sig == nil {
		return
	}
	if used && len(funcParams(sig)) > 0 && g.callsNow(fn) {
		g.headerSigs[group] = append(g.headerSigs[group], sig)
	}
	for v := range sig.Results().Variables() {
		if rs := coreSignature(v.Type()); rs != nil && len(funcParams(rs)) > 0 {
			g.headerSigs[group] = append(g.headerSigs[group], rs)
		}
	}
}

// headerValue tells, whether a function value of the signature may be a function value of a header package, which
// calls its function arguments.
func (g *recGraph) headerValue(group *Package, sig *types.Signature) bool {
	for _, h := range g.headerSigs[group] {
		if fit, _ := signaturesFit(sig, h); fit {
			return true
		}
	}
	return false
}

// valueEdges adds the edges of a call of a function value of the signature.
func (g *recGraph) valueEdges(from callNode, group *Package, sig *types.Signature, kind CallKind, pos token.Pos, expr string) {
	if sig == nil {
		return
	}
	for _, v := range g.values[arityOf(group, sig)] {
		// a value fits exactly, if one of its signatures does
		found, byArity := false, true
		for _, s := range v.sigs {
			if fit, arity := signaturesFit(sig, s); fit {
				found, byArity = true, byArity && arity
			}
		}
		if found {
			// each value is a node once, so the edge is new, unless another kind of call at the same position
			// reaches it, which does not change the graph
			g.edges[from] = append(g.edges[from], callEdge{to: v.node, kind: kind, pos: pos, expr: expr, bySig: true, arity: byArity})
		}
	}
}

// methodEdges adds the edges of a call of a method of an interface or of a type parameter.
func (g *recGraph) methodEdges(from callNode, group *Package, name string, kind CallKind, pos token.Pos, expr string) {
	for _, m := range g.methods[methodKey{group: group, name: name}] {
		g.edge(from, callEdge{to: m, kind: kind, pos: pos, expr: expr})
	}
}

// argEdges adds the edges of a call, which calls its function arguments, e.g. of a function of a header package.
func (g *recGraph) argEdges(from callNode, info *types.Info, group *Package, call *ast.CallExpr, sig *types.Signature, pos token.Pos, expr string) {
	n := sig.Params().Len()
	for _, i := range funcParams(sig) {
		pt := sig.Params().At(i).Type()
		args := call.Args[min(i, len(call.Args)):min(i+1, len(call.Args))]
		if sig.Variadic() && i == n-1 {
			if call.Ellipsis.IsValid() {
				// the elements of a slice are no values of the subset, which a call may reach
				continue
			}
			pt = pt.Underlying().(*types.Slice).Elem()
			args = call.Args[min(i, len(call.Args)):]
		}
		for _, a := range args {
			g.argEdge(from, info, group, a, pt, pos, expr)
		}
	}
}

// argEdge adds the edge of a call of the function argument.
func (g *recGraph) argEdge(from callNode, info *types.Info, group *Package, arg ast.Expr, param types.Type, pos token.Pos, expr string) {
	arg = ast.Unparen(arg)
	// a conversion into a named function type
	if c, ok := arg.(*ast.CallExpr); ok && len(c.Args) == 1 {
		if tv, ok := info.Types[c.Fun]; ok && tv.IsType() {
			g.argEdge(from, info, group, c.Args[0], param, pos, expr)
			return
		}
	}
	if lit, ok := arg.(*ast.FuncLit); ok {
		g.edge(from, callEdge{to: callNode{lit: lit}, kind: CallCallback, pos: pos, expr: expr})
		return
	}
	if fn, iface, ok := funcRef(info, arg); ok {
		switch {
		case iface:
			g.methodEdges(from, group, fn.Name(), CallCallback, pos, expr)
		case g.isUser(fn):
			g.edge(from, callEdge{to: callNode{fn: fn.Origin()}, kind: CallCallback, pos: pos, expr: expr})
		}
		// a function of a header package calls only functions of the host
		return
	}
	sig := coreSignature(info.TypeOf(arg))
	if sig == nil {
		sig = coreSignature(param)
	}
	g.valueEdges(from, group, sig, CallCallback, pos, expr)
}

// recWalker adds the calls of a function or of the initializer of a package variable to the graph.
type recWalker struct {
	g    *recGraph
	p    *Package
	info *types.Info
	grp  *Package
	// calls are the identifiers of the functions, which are called or ranged over by name, so that they are no
	// values
	calls map[*ast.Ident]bool
	stack []callNode
	// ranged is the call of a function of a header package, over whose result a range runs, e.g. l.All(): its
	// iterator calls only the body of the loop
	ranged *ast.CallExpr
}

func (w *recWalker) cur() callNode {
	return w.stack[len(w.stack)-1]
}

// inside adds the calls of the node, e.g. of a function literal.
func (w *recWalker) inside(node callNode, n ast.Node) {
	w.g.pkgOf[node] = w.p
	w.stack = append(w.stack, node)
	ast.Inspect(n, w.visit)
	w.stack = w.stack[:len(w.stack)-1]
}

func (w *recWalker) visit(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.FuncLit:
		lit := callNode{lit: n}
		sig, _ := w.info.TypeOf(n).(*types.Signature)
		w.g.addValue(w.grp, lit, sig)
		w.inside(lit, n.Body)
		return false
	case *ast.RangeStmt:
		if sig := coreSignature(w.info.TypeOf(n.X)); sig != nil {
			w.rangeFunc(n, sig)
			return false
		}
	case *ast.CallExpr:
		w.call(n)
	case *ast.SelectorExpr:
		w.methodValue(n)
	case *ast.Ident:
		w.funcValue(n)
	}
	return true
}

func (w *recWalker) call(call *ast.CallExpr) {
	if tv, ok := w.info.Types[call.Fun]; ok && (tv.IsType() || tv.IsBuiltin()) {
		return
	}
	from, pos, expr := w.cur(), callPos(call), exprString(call.Fun)
	if lit, ok := ast.Unparen(call.Fun).(*ast.FuncLit); ok {
		// the call of a literal calls only it
		w.g.edge(from, callEdge{to: callNode{lit: lit}, kind: CallStatic, pos: pos, expr: expr})
		return
	}
	if callee := staticCallee(w.info, call); callee != nil {
		if w.g.isUser(callee) {
			w.g.edge(from, callEdge{to: callNode{fn: callee.Origin()}, kind: CallStatic, pos: pos, expr: expr})
			return
		}
		if _, ok := w.g.decls[callee.Origin()]; !ok {
			return
		}
		// a function of a header package or of the standard library
		sig := calleeSignature(w.info, call.Fun)
		if sig == nil {
			return
		}
		if call != w.ranged {
			w.g.useHeaderValue(w.grp, callee, sig, false)
		}
		if len(funcParams(sig)) > 0 && w.g.callsNow(callee) {
			info, grp := w.info, w.grp
			w.g.resolve = append(w.g.resolve, func() { w.g.argEdges(from, info, grp, call, sig, pos, expr) })
		}
		return
	}
	if se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
		if sel := w.info.Selections[se]; sel != nil && sel.Kind() == types.MethodVal && dynamicRecv(sel.Recv()) {
			grp, name := w.grp, se.Sel.Name
			w.g.resolve = append(w.g.resolve, func() { w.g.methodEdges(from, grp, name, CallInterface, pos, expr) })
			return
		}
	}
	sig := coreSignature(w.info.TypeOf(call.Fun))
	if sig == nil {
		return
	}
	info, grp := w.info, w.grp
	w.g.resolve = append(w.g.resolve, func() {
		w.g.valueEdges(from, grp, sig, CallValue, pos, expr)
		if len(funcParams(sig)) > 0 && w.g.headerValue(grp, sig) {
			w.g.argEdges(from, info, grp, call, sig, pos, expr)
		}
	})
}

// rangeFunc adds the calls of a range over a function: the range calls the function, which calls the body through
// yield. The body is a node of its own with the signature of yield.
func (w *recWalker) rangeFunc(n *ast.RangeStmt, sig *types.Signature) {
	from, pos, expr := w.cur(), n.X.Pos(), exprString(n.X)
	body := callNode{body: n}
	var yield *types.Signature
	if sig.Params().Len() == 1 {
		yield = coreSignature(sig.Params().At(0).Type())
	}
	x := ast.Unparen(n.X)
	if w.headerCall(x) {
		w.ranged = x.(*ast.CallExpr)
	}
	for _, e := range []ast.Expr{n.Key, n.Value, n.X} {
		if e != nil {
			ast.Inspect(e, w.visit)
		}
	}
	w.ranged = nil

	fn, iface, ref := funcRef(w.info, x)
	lit, isLit := x.(*ast.FuncLit)
	switch {
	case ref && iface:
		w.g.addValue(w.grp, body, yield)
		grp, name := w.grp, fn.Name()
		w.g.resolve = append(w.g.resolve, func() { w.g.methodEdges(from, grp, name, CallRange, pos, expr) })
	case ref && w.g.isUser(fn):
		w.g.addValue(w.grp, body, yield)
		w.g.edge(from, callEdge{to: callNode{fn: fn.Origin()}, kind: CallRange, pos: pos, expr: expr})
	case ref:
		// a function of a header package calls the body
		if w.g.callsNow(fn) {
			w.g.edge(from, callEdge{to: body, kind: CallRange, pos: pos, expr: expr})
		}
	case isLit:
		w.g.addValue(w.grp, body, yield)
		w.g.edge(from, callEdge{to: callNode{lit: lit}, kind: CallRange, pos: pos, expr: expr})
	case w.headerCall(x):
		// the iterator of a header package, e.g. l.All(), calls only this body
		w.g.edge(from, callEdge{to: body, kind: CallRange, pos: pos, expr: expr})
	default:
		w.g.addValue(w.grp, body, yield)
		grp := w.grp
		w.g.resolve = append(w.g.resolve, func() {
			w.g.valueEdges(from, grp, sig, CallRange, pos, expr)
			if w.g.headerValue(grp, sig) {
				w.g.edge(from, callEdge{to: body, kind: CallRange, pos: pos, expr: expr, bySig: true})
			}
		})
	}
	w.inside(body, n.Body)
}

// headerCall tells, whether the expression calls a function of a header package or of the standard library.
func (w *recWalker) headerCall(x ast.Expr) bool {
	call, ok := x.(*ast.CallExpr)
	if !ok {
		return false
	}
	callee := staticCallee(w.info, call)
	if callee == nil || w.g.isUser(callee) {
		return false
	}
	_, ok = w.g.decls[callee.Origin()]
	return ok
}

// methodValue notes a method, which is used as a value, e.g. f := s.Area.
func (w *recWalker) methodValue(se *ast.SelectorExpr) {
	sel := w.info.Selections[se]
	if sel == nil || sel.Kind() != types.MethodVal || w.calls[se.Sel] {
		return
	}
	fn, ok := sel.Obj().(*types.Func)
	if !ok {
		return
	}
	sig, _ := sel.Type().(*types.Signature)
	switch {
	case dynamicRecv(sel.Recv()):
		// a method value of an interface or of a type parameter may be each method of its name
		for _, m := range w.g.methods[methodKey{group: w.grp, name: fn.Name()}] {
			w.g.addValue(w.grp, m, methodValueSignature(m.fn))
		}
	case w.g.isUser(fn):
		w.g.addValue(w.grp, callNode{fn: fn.Origin()}, sig)
	default:
		if _, ok := w.g.decls[fn.Origin()]; ok {
			w.g.useHeaderValue(w.grp, fn, sig, true)
		}
	}
}

// funcValue notes a function, which is used as a value, e.g. apply(Indirect).
func (w *recWalker) funcValue(id *ast.Ident) {
	if w.calls[id] {
		return
	}
	fn, ok := w.info.Uses[id].(*types.Func)
	if !ok || fn.Signature().Recv() != nil {
		// a method is used as a value only by a selector, see methodValue
		return
	}
	sig := identSignature(w.info, id)
	if w.g.isUser(fn) {
		w.g.addValue(w.grp, callNode{fn: fn.Origin()}, sig)
		return
	}
	if _, ok := w.g.decls[fn.Origin()]; ok {
		w.g.useHeaderValue(w.grp, fn, sig, true)
	}
}

// report reports each strongly connected component of the graph, which has a cycle and a node in the profile.
func (g *recGraph) report(d *diagnoser, inProfile map[*Package]bool) {
	for _, scc := range stronglyConnected(g.edges) {
		in := map[callNode]bool{}
		for _, n := range scc {
			in[n] = true
		}
		if len(scc) == 1 {
			n := scc[0]
			self, static := false, true
			for _, e := range g.edges[n] {
				if e.to == n {
					self = true
					static = static && e.kind == CallStatic
				}
			}
			// a function, which calls itself only by name with a decreasing argument, ends
			if !self || n.fn != nil && static && g.decreasing[n.fn] {
				continue
			}
		}

		// the named functions of the profile, else the first literal or body of the profile
		var reported []callNode
		slices.SortFunc(scc, func(a, b callNode) int { return comparePos(g.prog.Fset, a.pos(), b.pos()) })
		for _, n := range scc {
			if n.fn != nil && inProfile[g.pkgOf[n]] {
				reported = append(reported, n)
			}
		}
		if len(reported) == 0 {
			for _, n := range scc {
				if inProfile[g.pkgOf[n]] {
					reported = append(reported, n)
					break
				}
			}
		}
		for _, r := range reported {
			g.reportCycle(d, r, g.shortestCycle(r, in))
		}
	}
}

// cycleStep is an edge of a cycle.
type cycleStep struct {
	from callNode
	e    callEdge
}

// shortestCycle returns a shortest cycle through the node within the component. Of several, it returns the one,
// whose sequence of positions of calls comes first by file, line and column.
func (g *recGraph) shortestCycle(r callNode, in map[callNode]bool) []cycleStep {
	// dist is the length of a shortest path from a node to r
	rev := map[callNode][]callNode{}
	for u := range in {
		for _, e := range g.edges[u] {
			if g.follow(u, e, in) {
				rev[e.to] = append(rev[e.to], u)
			}
		}
	}
	dist := map[callNode]int{r: 0}
	queue := []callNode{r}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, u := range rev[v] {
			if _, ok := dist[u]; !ok {
				dist[u] = dist[v] + 1
				queue = append(queue, u)
			}
		}
	}
	n := -1
	for _, e := range g.edges[r] {
		if d, ok := dist[e.to]; ok && g.follow(r, e, in) && (n < 0 || d+1 < n) {
			n = d + 1
		}
	}

	// follow the shortest paths and keep at each step the targets of the first call
	frontier := []callNode{r}
	var back []map[callNode]cycleStep
	for k := range n {
		remaining := n - k
		var best token.Position
		cand := map[callNode]cycleStep{}
		for _, u := range frontier {
			for _, e := range g.edges[u] {
				if d, ok := dist[e.to]; !ok || !g.follow(u, e, in) || d != remaining-1 {
					continue
				}
				p := g.prog.Fset.Position(e.pos)
				switch c := comparePosition(p, best); {
				case len(cand) == 0 || c < 0:
					cand = map[callNode]cycleStep{e.to: {from: u, e: e}}
					best = p
				case c == 0:
					if _, ok := cand[e.to]; !ok {
						cand[e.to] = cycleStep{from: u, e: e}
					}
				}
			}
		}
		back = append(back, cand)
		frontier = frontier[:0]
		for v := range cand {
			frontier = append(frontier, v)
		}
		slices.SortFunc(frontier, func(a, b callNode) int { return comparePos(g.prog.Fset, a.pos(), b.pos()) })
	}
	steps := make([]cycleStep, n)
	v := r
	for k := n - 1; k >= 0; k-- {
		steps[k] = back[k][v]
		v = steps[k].from
	}
	return steps
}

// follow tells, whether a cycle, which is reported, follows the edge within the component: not a call of a function
// by name to itself with a decreasing argument, which ends, but the other calls, which make the cycle.
func (g *recGraph) follow(from callNode, e callEdge, in map[callNode]bool) bool {
	return in[e.to] && !(e.to == from && e.kind == CallStatic && from.fn != nil && g.decreasing[from.fn])
}

func comparePos(fset *token.FileSet, a, b token.Pos) int {
	return comparePosition(fset.Position(a), fset.Position(b))
}

func comparePosition(a, b token.Position) int {
	if c := strings.Compare(a.Filename, b.Filename); c != 0 {
		return c
	}
	if a.Line != b.Line {
		return a.Line - b.Line
	}
	return a.Column - b.Column
}

// reportCycle reports the cycle at its first node.
func (g *recGraph) reportCycle(d *diagnoser, r callNode, steps []cycleStep) {
	links := make([]CycleLink, len(steps))
	for i, s := range steps {
		p := g.prog.Fset.Position(s.e.pos)
		pkg := ""
		if q := g.pkgOf[s.from]; q != nil {
			pkg = q.Path
		}
		links[i] = CycleLink{Package: pkg, Func: s.from.name(), File: p.Filename, Line: p.Line, Col: p.Column, Call: s.e.kind, Expr: s.e.expr, Arity: s.e.arity}
	}

	// <paket>.<f> may call itself: <glied> → <glied> → … → <paket>.<f>, each member with the call, which leads to the
	// next one, each call of another kind than static with its kind and expression
	var sb strings.Builder
	sb.WriteString(linkName(links[0]))
	sb.WriteString(" may call itself: ")
	fix := "use a loop with a bound, or recurse with an argument, which decreases: if n <= 0 { return ... } as the first statement and f(n-1) for each call"
	value := ""
	for i, l := range links {
		at := l.File + ":" + strconv.Itoa(l.Line)
		fmt.Fprintf(&sb, "%s (%s) → ", linkName(l), at)
		if l.Call == CallStatic {
			continue
		}
		via := string(l.Call)
		switch l.Call {
		case CallValue:
			if value == "" {
				value = l.Expr
			}
		case CallCallback:
			via += " of"
		case CallRange:
			via += " over"
		}
		resolved := ""
		if steps[i].e.bySig {
			resolved = ", resolved by its signature"
			if l.Arity {
				resolved = ", resolved by the number of its parameters and results"
			}
		}
		fmt.Fprintf(&sb, "via the %s %s (%s)%s to ", via, l.Expr, at, resolved)
	}
	sb.WriteString(linkName(links[0]))
	if value != "" {
		fix = "call a named function instead of the function value " + value + ", or " + fix
	}
	pos, end := r.marked()
	d.add(pos, end, Diagnostic{Code: CodeTermination, Msg: sb.String(), Why: "the terminating profile requires, that every call ends, which recursion can not prove", Fix: fix, Cycle: links})
}

// linkName returns the name of the link in a message, e.g. example.com/app.F or example.com/app.function literal.
func linkName(l CycleLink) string {
	return l.Package + "." + l.Func
}

// importGroups maps each package of the subset to a representative of the packages, which are connected to it by
// imports in either direction, also through other packages. Values flow only along imports, e.g. a package, which
// imports two others, passes a function of one to the other.
func importGroups(prog *Program) map[*Package]*Package {
	parent := map[*Package]*Package{}
	var find func(p *Package) *Package
	find = func(p *Package) *Package {
		if parent[p] == p {
			return p
		}
		root := find(parent[p])
		parent[p] = root
		return root
	}
	for _, p := range prog.Packages {
		if p.Kind == KindUser {
			parent[p] = p
		}
	}
	for _, p := range prog.Packages {
		if p.Kind != KindUser || p.Types == nil {
			continue
		}
		for _, imp := range p.Types.Imports() {
			if q := prog.Package(imp.Path()); q != nil && q.Kind == KindUser {
				parent[find(q)] = find(p)
			}
		}
	}
	res := map[*Package]*Package{}
	for p := range parent {
		res[p] = find(p)
	}
	return res
}

// staticCallee returns the function or method, which the call calls statically, or nil.
func staticCallee(info *types.Info, call *ast.CallExpr) *types.Func {
	fn, iface, ok := funcRef(info, call.Fun)
	if !ok || iface {
		return nil
	}
	return fn
}

// funcRef returns the function or method, which the expression names, e.g. F, pkg.F, F[int32] or x.M, and whether
// it is the method of an interface or of a type parameter.
func funcRef(info *types.Info, x ast.Expr) (fn *types.Func, iface, ok bool) {
	x = ast.Unparen(x)
	switch e := x.(type) {
	case *ast.IndexExpr:
		x = ast.Unparen(e.X)
	case *ast.IndexListExpr:
		x = ast.Unparen(e.X)
	}
	switch e := x.(type) {
	case *ast.Ident:
		fn, ok = info.Uses[e].(*types.Func)
		return fn, false, ok
	case *ast.SelectorExpr:
		if sel := info.Selections[e]; sel != nil {
			if sel.Kind() != types.MethodVal {
				return nil, false, false
			}
			fn, ok = sel.Obj().(*types.Func)
			return fn, ok && dynamicRecv(sel.Recv()), ok
		}
		fn, ok = info.Uses[e.Sel].(*types.Func)
		return fn, false, ok
	}
	return nil, false, false
}

// dynamicRecv tells, whether the type of a receiver is an interface or a type parameter, whose method a call
// selects at run time.
func dynamicRecv(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

// coreSignature returns the signature of a function type or of a type parameter, whose core type is a function, or
// nil.
func coreSignature(t types.Type) *types.Signature {
	if t == nil {
		return nil
	}
	tp, ok := types.Unalias(t).(*types.TypeParam)
	if !ok {
		s, _ := t.Underlying().(*types.Signature)
		return s
	}
	iface, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok {
		return nil
	}
	var res *types.Signature
	for i := range iface.NumEmbeddeds() {
		var terms []types.Type
		if u, ok := iface.EmbeddedType(i).(*types.Union); ok {
			for j := range u.Len() {
				terms = append(terms, u.Term(j).Type())
			}
		} else {
			terms = append(terms, iface.EmbeddedType(i))
		}
		for _, term := range terms {
			s := coreSignature(term)
			if s == nil || res != nil && !types.Identical(res, s) {
				return nil
			}
			res = s
		}
	}
	return res
}

// calleeSignature returns the signature of the function of a call, also of an instance of a generic function.
func calleeSignature(info *types.Info, fun ast.Expr) *types.Signature {
	fun = ast.Unparen(fun)
	switch x := fun.(type) {
	case *ast.Ident:
		return identSignature(info, x)
	case *ast.SelectorExpr:
		if info.Selections[x] == nil {
			return identSignature(info, x.Sel)
		}
	}
	return coreSignature(info.TypeOf(fun))
}

// identSignature returns the signature of the function, which the identifier names, of its instance, if it is
// generic.
func identSignature(info *types.Info, id *ast.Ident) *types.Signature {
	if inst, ok := info.Instances[id]; ok {
		if s, ok := inst.Type.(*types.Signature); ok {
			return s
		}
	}
	return coreSignature(info.TypeOf(id))
}

// methodValueSignature returns the signature of a method without its receiver, which is the type of its method
// value.
func methodValueSignature(fn *types.Func) *types.Signature {
	sig := fn.Signature()
	return types.NewSignatureType(nil, nil, nil, sig.Params(), sig.Results(), sig.Variadic())
}

// callPos returns the position of a call: the name of the function or method, or the parenthesis.
func callPos(call *ast.CallExpr) token.Pos {
	fun := ast.Unparen(call.Fun)
	switch x := fun.(type) {
	case *ast.IndexExpr:
		fun = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		fun = ast.Unparen(x.X)
	}
	switch x := fun.(type) {
	case *ast.Ident:
		return x.Pos()
	case *ast.SelectorExpr:
		return x.Sel.Pos()
	}
	return call.Lparen
}

// exprString returns the expression with the arguments of its calls shortened to (…), e.g. ui.BoxAlign(…).WithFrame.
func exprString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	case *ast.CallExpr:
		if len(x.Args) == 0 {
			return exprString(x.Fun) + "()"
		}
		return exprString(x.Fun) + "(…)"
	case *ast.IndexExpr:
		return exprString(x.X) + "[" + exprString(x.Index) + "]"
	case *ast.IndexListExpr:
		var args []string
		for _, a := range x.Indices {
			args = append(args, exprString(a))
		}
		return exprString(x.X) + "[" + strings.Join(args, ", ") + "]"
	case *ast.ParenExpr:
		return "(" + exprString(x.X) + ")"
	case *ast.StarExpr:
		return "*" + exprString(x.X)
	case *ast.FuncLit:
		return "func literal"
	}
	return types.ExprString(e)
}

// callIdents returns the identifiers of the functions, which the calls and the ranges over functions in the node
// call by name.
func callIdents(n ast.Node) map[*ast.Ident]bool {
	res := map[*ast.Ident]bool{}
	mark := func(fun ast.Expr) {
		fun = ast.Unparen(fun)
		switch x := fun.(type) {
		case *ast.IndexExpr:
			fun = ast.Unparen(x.X)
		case *ast.IndexListExpr:
			fun = ast.Unparen(x.X)
		}
		switch x := fun.(type) {
		case *ast.Ident:
			res[x] = true
		case *ast.SelectorExpr:
			res[x.Sel] = true
		}
	}
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			mark(n.Fun)
		case *ast.RangeStmt:
			mark(n.X)
		}
		return true
	})
	return res
}

// stronglyConnected returns the strongly connected components of the graph by the algorithm of Tarjan.
func stronglyConnected(edges map[callNode][]callEdge) [][]callNode {
	index := map[callNode]int{}
	low := map[callNode]int{}
	onStack := map[callNode]bool{}
	var stack []callNode
	var res [][]callNode
	next := 0

	var nodes []callNode
	seen := map[callNode]bool{}
	add := func(n callNode) {
		if !seen[n] {
			seen[n] = true
			nodes = append(nodes, n)
		}
	}
	for from, es := range edges {
		add(from)
		for _, e := range es {
			add(e.to)
		}
	}
	slices.SortFunc(nodes, func(a, b callNode) int { return int(a.pos() - b.pos()) })

	var connect func(v callNode)
	connect = func(v callNode) {
		index[v] = next
		low[v] = next
		next++
		stack = append(stack, v)
		onStack[v] = true
		for _, e := range edges[v] {
			w := e.to
			if _, ok := index[w]; !ok {
				connect(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] == index[v] {
			var scc []callNode
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			res = append(res, scc)
		}
	}
	for _, n := range nodes {
		if _, ok := index[n]; !ok {
			connect(n)
		}
	}
	return res
}

// decreasingRecursion tells, whether each call of the function to itself passes a parameter of an integer type,
// which decreases, and the function returns first, if the parameter is small: if n <= c { return ... }.
func decreasingRecursion(info *types.Info, fd *ast.FuncDecl, fn *types.Func) bool {
	if len(fd.Body.List) == 0 {
		return false
	}
	first, ok := fd.Body.List[0].(*ast.IfStmt)
	if !ok || first.Init != nil || len(first.Body.List) == 0 {
		return false
	}
	if _, ok := first.Body.List[len(first.Body.List)-1].(*ast.ReturnStmt); !ok {
		return false
	}
	cond, ok := ast.Unparen(first.Cond).(*ast.BinaryExpr)
	if !ok || (cond.Op != token.LEQ && cond.Op != token.LSS) {
		return false
	}
	id, ok := ast.Unparen(cond.X).(*ast.Ident)
	if tv, isConst := info.Types[cond.Y]; !ok || !isConst || tv.Value == nil {
		return false
	}
	param, ok := info.Uses[id].(*types.Var)
	if !ok || !isInteger(param.Type()) {
		return false
	}
	idx := -1
	for i := 0; i < fn.Signature().Params().Len(); i++ {
		if fn.Signature().Params().At(i) == param {
			idx = i
		}
	}
	if idx < 0 {
		return false
	}

	// the parameter must not be assigned and each call must pass param - c or param / c with c > 0 and c > 1
	assigned := false
	all := true
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, l := range n.Lhs {
				if lid, ok := ast.Unparen(l).(*ast.Ident); ok && info.Uses[lid] == param {
					assigned = true
				}
			}
		case *ast.IncDecStmt:
			if lid, ok := ast.Unparen(n.X).(*ast.Ident); ok && info.Uses[lid] == param {
				assigned = true
			}
		case *ast.CallExpr:
			if staticCallee(info, n) != fn || idx >= len(n.Args) {
				return true
			}
			b, ok := ast.Unparen(n.Args[idx]).(*ast.BinaryExpr)
			if !ok {
				all = false
				return true
			}
			x, ok := ast.Unparen(b.X).(*ast.Ident)
			tv, isConst := info.Types[b.Y]
			if !ok || info.Uses[x] != param || !isConst || tv.Value == nil {
				all = false
				return true
			}
			switch {
			case b.Op == token.SUB && constant.Sign(tv.Value) > 0:
			case b.Op == token.QUO && constant.Compare(tv.Value, token.GTR, constant.MakeInt64(1)) && constant.Sign(info.Types[cond.Y].Value) >= 0:
				// the guard leaves only positive values, which the division decreases
			default:
				all = false
			}
		}
		return true
	})
	return !assigned && all
}
