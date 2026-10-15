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
	"math"
	"slices"
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

// checkLoops reports the loops of the package, whose end the compiler can not prove.
func (c *checker) checkLoops(f *ast.File) {
	const why = "the terminating profile requires, that every loop ends"
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ForStmt:
			if msg := c.boundedFor(n); msg != "" {
				c.report(n, CodeTermination, msg, why, "count with a bound, e.g. for i := int32(0); i < n; i++, range over a list or a number, or limit the loop with fuel: f := fuel.New(1000); for f.Burn() && cond { ... }")
			}
		case *ast.RangeStmt:
			if msg := c.boundedRange(n); msg != "" {
				c.report(n, CodeTermination, msg, why, "do not append to the list, while ranging over it, e.g. collect the new elements in another list")
			}
		}
		return true
	})
}

// boundedFor returns, why the loop may not end, or "".
func (c *checker) boundedFor(n *ast.ForStmt) string {
	if n.Cond == nil {
		return "the loop has no condition"
	}
	if c.burnsFuel(n.Cond) {
		return ""
	}

	// for i := a; i < b; i++ with a counter, which only the post statement changes, and a bound, which does not
	// change
	counter, step := c.loopStep(n.Post)
	if counter == nil {
		return "the loop has no counter, which its post statement increments or decrements by a constant"
	}
	if c.assignedIn(counter, n.Body) {
		return "the body assigns the counter " + counter.Name()
	}
	b, ok := ast.Unparen(n.Cond).(*ast.BinaryExpr)
	if !ok {
		return "the condition does not compare the counter " + counter.Name() + " with a bound"
	}
	op, bound := b.Op, b.Y
	if c.isObj(b.Y, counter) {
		// b > i is i < b
		op, bound = mirror(op), b.X
	} else if !c.isObj(b.X, counter) {
		return "the condition does not compare the counter " + counter.Name() + " with a bound"
	}
	switch {
	case step > 0 && (op == token.LSS || op == token.LEQ), step < 0 && (op == token.GTR || op == token.GEQ):
	default:
		return "the condition does not stop the counter " + counter.Name() + " in its direction"
	}
	if msg := c.fixedBound(bound, n.Body, op == token.LEQ || op == token.GEQ); msg != "" {
		return msg
	}
	return ""
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

// fixedBound returns, why the bound may change during the loop or let the counter wrap around, or "".
func (c *checker) fixedBound(bound ast.Expr, body *ast.BlockStmt, inclusive bool) string {
	if tv, ok := c.info.Types[bound]; ok && tv.Value != nil {
		if v, exact := constant.Int64Val(constant.ToInt(tv.Value)); inclusive && (!exact || v >= math.MaxInt32 || v <= math.MinInt32) {
			return "the counter wraps around at the bound"
		}
		return ""
	}
	if inclusive {
		return "an inclusive bound, which is not a constant, may let the counter wrap around"
	}
	id, ok := ast.Unparen(bound).(*ast.Ident)
	if !ok {
		return "the bound must be a constant or a variable, because an expression may change in the loop"
	}
	v, ok := c.info.Uses[id].(*types.Var)
	if !ok || c.assignedIn(v, body) {
		return "the body assigns the bound " + id.Name
	}
	return ""
}

// loopStep returns the counter and the sign of its step of the post statement i++, i--, i += c or i -= c.
func (c *checker) loopStep(post ast.Stmt) (*types.Var, int) {
	var x ast.Expr
	step := 0
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
		step = 1
		if s.Tok == token.SUB_ASSIGN {
			step = -1
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

// boundedRange returns, why the range may not end, or "". A range over a list sees the elements, which its body
// appends.
func (c *checker) boundedRange(n *ast.RangeStmt) string {
	call, ok := ast.Unparen(n.X).(*ast.CallExpr)
	if !ok {
		return ""
	}
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	t := c.info.TypeOf(se.X)
	named, ok := types.Unalias(t).(*types.Named)
	if !ok || !isOpaque(t) || named.Obj().Name() != "List" {
		return ""
	}
	root := rootObject(c.info, se.X)
	if root == nil {
		return "the list may grow, while the loop ranges over it"
	}
	grows := false
	ast.Inspect(n.Body, func(m ast.Node) bool {
		mc, ok := m.(*ast.CallExpr)
		if !ok {
			return true
		}
		ms, ok := ast.Unparen(mc.Fun).(*ast.SelectorExpr)
		if ok && ms.Sel.Name == "Append" && isOpaque(c.info.TypeOf(ms.X)) && rootObject(c.info, ms.X) == root {
			grows = true
		}
		return true
	})
	if grows {
		return "the body appends to the list, over which the loop ranges"
	}
	return ""
}

// callNode is a function or a function literal of the call graph.
type callNode struct {
	fn  *types.Func
	lit *ast.FuncLit
}

func (n callNode) name() string {
	if n.fn != nil {
		return n.fn.Name()
	}
	return "a function literal"
}

func (n callNode) pos() token.Pos {
	if n.fn != nil {
		return n.fn.Pos()
	}
	return n.lit.Pos()
}

// checkRecursion reports the functions of the packages of the profile, which may call themselves. The call graph
// contains the static calls of all packages of the program in the subset. A call of a function value may reach
// every function, which is used as a value, and every function literal, a call of a method of an interface every
// method of this name. A cycle of function literals alone is not reported, e.g. an iterator, which calls the
// body of its loop.
func checkRecursion(prog *Program, d *diagnoser, decls map[*types.Func]funcSrc) {
	var profiled []*Package
	for _, p := range prog.Packages {
		if p.Kind == KindUser && prog.terminating(p) {
			profiled = append(profiled, p)
		}
	}
	if len(profiled) == 0 {
		return
	}

	edges := map[callNode][]callNode{}
	var valueFuncs []callNode
	ifaceMethods := map[string][]callNode{}
	type dynamic struct {
		from   callNode
		method string // "" for a function value
	}
	var dynamics []dynamic
	decreasing := map[*types.Func]bool{}

	for _, p := range prog.Packages {
		if p.Kind != KindUser {
			continue
		}
		info := p.Info
		// the methods, which a call of a method of an interface may reach, are approximated by all methods
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Recv == nil {
					continue
				}
				if fn, ok := info.Defs[fd.Name].(*types.Func); ok {
					ifaceMethods[fn.Name()] = append(ifaceMethods[fn.Name()], callNode{fn: fn})
				}
			}
		}
		for _, f := range p.Files {
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				fn, ok := info.Defs[fd.Name].(*types.Func)
				if !ok {
					continue
				}
				if decreasingRecursion(info, fd, fn) {
					decreasing[fn] = true
				}

				calls := callIdents(fd.Body)
				var stack []callNode
				stack = append(stack, callNode{fn: fn})
				var visit func(n ast.Node) bool
				visit = func(n ast.Node) bool {
					cur := stack[len(stack)-1]
					switch n := n.(type) {
					case *ast.FuncLit:
						lit := callNode{lit: n}
						valueFuncs = append(valueFuncs, lit)
						stack = append(stack, lit)
						ast.Inspect(n.Body, visit)
						stack = stack[:len(stack)-1]
						return false
					case *ast.CallExpr:
						callee := staticCallee(info, n)
						switch {
						case callee != nil:
							if src, ok := decls[callee.Origin()]; ok && src.pkg.Kind == KindUser {
								edges[cur] = append(edges[cur], callNode{fn: callee.Origin()})
							}
						case isMethodOfInterface(info, n):
							dynamics = append(dynamics, dynamic{from: cur, method: ast.Unparen(n.Fun).(*ast.SelectorExpr).Sel.Name})
						case isDynamicCall(info, n):
							dynamics = append(dynamics, dynamic{from: cur})
						}
					case *ast.Ident:
						// a function, which is used as a value, may be called by any call of a function value
						if fn, ok := info.Uses[n].(*types.Func); ok && !calls[n] {
							valueFuncs = append(valueFuncs, callNode{fn: fn.Origin()})
						}
					}
					return true
				}
				ast.Inspect(fd.Body, visit)
			}
		}
	}

	for _, dy := range dynamics {
		if dy.method != "" {
			edges[dy.from] = append(edges[dy.from], ifaceMethods[dy.method]...)
			continue
		}
		edges[dy.from] = append(edges[dy.from], valueFuncs...)
	}

	inProfile := map[*types.Package]bool{}
	for _, p := range profiled {
		inProfile[p.Types] = true
	}
	for _, scc := range stronglyConnected(edges) {
		var funcs []callNode
		for _, n := range scc {
			if n.fn != nil {
				funcs = append(funcs, n)
			}
		}
		if len(funcs) == 0 {
			continue
		}
		self := len(scc) == 1 && slices.Contains(edges[scc[0]], scc[0])
		if len(scc) == 1 && !self {
			continue
		}
		if len(scc) == 1 && decreasing[scc[0].fn] {
			continue
		}
		var names []string
		for _, n := range funcs {
			names = append(names, n.name())
		}
		slices.Sort(names)
		for _, n := range funcs {
			if !inProfile[n.fn.Pkg()] {
				continue
			}
			pos := n.pos()
			d.report(pos, pos+token.Pos(len(n.name())), CodeTermination, n.name()+" may call itself through "+strings.Join(names, ", "), "the terminating profile requires, that every call ends, which recursion can not prove", "use a loop with a bound, or recurse with an argument, which decreases: if n <= 0 { return ... } as the first statement and f(n-1) for each call")
		}
	}
}

// staticCallee returns the function or method, which the call calls statically, or nil.
func staticCallee(info *types.Info, call *ast.CallExpr) *types.Func {
	fun := ast.Unparen(call.Fun)
	switch x := fun.(type) {
	case *ast.IndexExpr:
		fun = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		fun = ast.Unparen(x.X)
	}
	switch x := fun.(type) {
	case *ast.Ident:
		fn, _ := info.Uses[x].(*types.Func)
		return fn
	case *ast.SelectorExpr:
		if sel := info.Selections[x]; sel != nil {
			if sel.Kind() != types.MethodVal {
				return nil
			}
			if _, iface := sel.Recv().Underlying().(*types.Interface); iface {
				return nil
			}
			fn, _ := sel.Obj().(*types.Func)
			return fn
		}
		fn, _ := info.Uses[x.Sel].(*types.Func)
		return fn
	}
	return nil
}

func isMethodOfInterface(info *types.Info, call *ast.CallExpr) bool {
	se, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	sel := info.Selections[se]
	if sel == nil || sel.Kind() != types.MethodVal {
		return false
	}
	_, iface := sel.Recv().Underlying().(*types.Interface)
	return iface
}

// isDynamicCall tells, whether the call calls a function value, e.g. a parameter or a field.
func isDynamicCall(info *types.Info, call *ast.CallExpr) bool {
	if tv, ok := info.Types[call.Fun]; ok && (tv.IsType() || tv.IsBuiltin()) {
		return false
	}
	_, ok := info.TypeOf(call.Fun).Underlying().(*types.Signature)
	return ok
}

// callIdents returns the identifiers of the functions, which the calls in the body call by name.
func callIdents(body *ast.BlockStmt) map[*ast.Ident]bool {
	res := map[*ast.Ident]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fun := ast.Unparen(call.Fun)
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
		return true
	})
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

// stronglyConnected returns the strongly connected components of the graph by the algorithm of Tarjan.
func stronglyConnected(edges map[callNode][]callNode) [][]callNode {
	index := map[callNode]int{}
	low := map[callNode]int{}
	onStack := map[callNode]bool{}
	var stack []callNode
	var res [][]callNode
	next := 0

	var nodes []callNode
	seen := map[callNode]bool{}
	for from, tos := range edges {
		for _, n := range append([]callNode{from}, tos...) {
			if !seen[n] {
				seen[n] = true
				nodes = append(nodes, n)
			}
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
		for _, w := range edges[v] {
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
