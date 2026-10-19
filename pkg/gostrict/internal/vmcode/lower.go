// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"math"
	"strings"
)

// compiler lowers the checked program into the IR and the IR into bytecode.
type compiler struct {
	src     *Source
	pkgs    map[string]*Package
	out     *Program
	types   map[string]*tinfo
	later   []func()
	funcs   map[string]*Func
	decls   map[*types.Func]funcDecl
	queue   []func()
	externs map[string]*extern
	natives map[string]*native
	globals map[*types.Var]*global
	nk      map[uint64]int32
	sk      map[string]int32
	typeIdx map[*tinfo]int32
	tctx    *types.Context
	// lits counts the function literals for their names
	lits int
	// hostInits initialize the variables of header packages
	hostInits []stmt
	// globalList are the package variables in the order of their declaration
	globalList []*global
	// branches counts the codes of the branches out of the bodies of ranges
	branches int32
}

type funcDecl struct {
	pkg  *Package
	decl *ast.FuncDecl
}

// global is a package variable: in the slots of the program, which all instances share, or of each instance, if
// its initialization calls the host, e.g. a scene of three.js.
type global struct {
	t    *tinfo
	off  layout
	inst bool
}

// compileError is the panic of an unsupported construct.
type compileError struct {
	pos token.Pos
	msg string
}

func (c *compiler) fail(pos token.Pos, format string, args ...any) {
	panic(compileError{pos: pos, msg: fmt.Sprintf(format, args...)})
}

// Compile compiles a checked program. The program refers to no syntax or types of the source, so that it does not
// keep them alive.
func Compile(src *Source) (p *Program, err error) {
	c := &compiler{
		src:     src,
		pkgs:    map[string]*Package{},
		out:     &Program{exports: map[string]*Func{}},
		types:   map[string]*tinfo{},
		funcs:   map[string]*Func{},
		decls:   map[*types.Func]funcDecl{},
		externs: map[string]*extern{},
		natives: map[string]*native{},
		globals: map[*types.Var]*global{},
		nk:      map[uint64]int32{},
		sk:      map[string]int32{},
		typeIdx: map[*tinfo]int32{},
		tctx:    types.NewContext(),
	}
	defer func() {
		if r := recover(); r != nil {
			ce, ok := r.(compileError)
			if !ok {
				panic(r)
			}
			err = fmt.Errorf("vm: %s: %s", src.Fset.Position(ce.pos), ce.msg)
		}
	}()
	c.compile()
	return c.out, nil
}

func (c *compiler) compile() {
	for _, p := range c.src.Packages {
		c.pkgs[p.Path] = p
		for _, f := range p.Files {
			for _, d := range f.Decls {
				if fd, ok := d.(*ast.FuncDecl); ok {
					if fn, ok := p.Info.Defs[fd.Name].(*types.Func); ok {
						c.decls[fn] = funcDecl{pkg: p, decl: fd}
					}
				}
			}
		}
	}

	reached := map[*types.Package]bool{}
	var visit func(p *types.Package)
	visit = func(p *types.Package) {
		if reached[p] {
			return
		}
		reached[p] = true
		for _, imp := range p.Imports() {
			visit(imp)
		}
	}
	for name, fn := range c.src.Exports {
		visit(fn.Pkg())
		if _, ok := c.src.Entries[fn]; ok {
			// the host calls an entry point within the aspect
			c.out.exports[name] = c.hookFor(fn)
			continue
		}
		c.out.exports[name] = c.funcFor(fn, nil, nil)
	}

	// the package variables in the order of their initialization
	init := &Func{name: "$init", params: tuple(), results: tuple()}
	l := c.newFnLower(init, nil)
	for _, p := range c.src.Packages {
		if !p.User || !reached[p.Types] {
			continue
		}
		l.info = p.Info
		for _, ini := range p.Info.InitOrder {
			var lhs []lval
			for _, v := range ini.Lhs {
				g := c.global(v)
				lhs = append(lhs, &lGlobal{t: g.t, g: g})
			}
			var rhs []expr
			if len(ini.Lhs) == 1 {
				rhs = []expr{l.val(ini.Rhs, lhs[0].ltyp())}
			} else {
				rhs = []expr{l.expr(ini.Rhs)}
			}
			l.body = append(l.body, &sAssign{lhs: lhs, rhs: rhs})
		}
	}
	init.body = l.body
	init.locals = l.frame
	c.out.init = init

	for {
		c.hostVariants()
		if len(c.queue) == 0 {
			break
		}
		for len(c.queue) > 0 {
			next := c.queue[0]
			c.queue = c.queue[1:]
			next()
		}
	}
	for _, f := range c.later {
		f()
	}
	init.body = append(c.hostInits, init.body...)
	c.divideGlobals(init)
	c.inline()

	gens := append(c.out.funcs, init)
	if c.out.instInit != nil {
		gens = append(gens, c.out.instInit)
	}
	for _, f := range gens {
		g := &gen{c: c, f: f}
		g.function()
		// the program keeps only the bytecode
		f.body = nil
	}
	for i, e := range c.out.entries {
		h := e.hook
		h.frame = e.inner.frame.max(h.params.l).max(h.results.l)
		h.code = []instr{{op: opEntry, a: int32(i)}, {op: opRet}}
	}
	c.plan()
	for _, t := range c.types {
		t.t, t.sig = nil, nil
	}
}

// hostVariants adds the variants of the unions of the boundary, which the program uses, with their methods, also
// those, which the program never names: the host may create any of them, e.g. a ui.TBox as a ui.View.
func (c *compiler) hostVariants() {
	for name, u := range c.src.Unions {
		ut, ok := c.types[name]
		if !ok || ut.t == nil {
			continue
		}
		for _, v := range u.Variants {
			vt, ok := c.types[v]
			if !ok {
				t := variantType(ut.t, v)
				if t == nil {
					continue
				}
				vt = c.typeOf(t)
			}
			c.methodsFor(vt, ut.t, token.NoPos)
		}
	}
}

// variantType returns the variant of the sealed interface u by its full name, or nil. The variants are declared in
// the package of u, a generic one with the type arguments of u.
func variantType(u types.Type, name string) types.Type {
	n, ok := types.Unalias(u).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return nil
	}
	scope := n.Obj().Pkg().Scope()
	for _, tn := range scope.Names() {
		obj, ok := scope.Lookup(tn).(*types.TypeName)
		if !ok || obj.IsAlias() {
			continue
		}
		var t types.Type = obj.Type()
		if v, ok := t.(*types.Named); ok && v.TypeParams().Len() > 0 {
			if v.TypeParams().Len() != n.TypeArgs().Len() {
				continue
			}
			var targs []types.Type
			for a := range n.TypeArgs().Types() {
				targs = append(targs, a)
			}
			inst, err := types.Instantiate(nil, v, targs, true)
			if err != nil {
				continue
			}
			t = inst
		}
		if typeKey(t) == name {
			return t
		}
	}
	return nil
}

// plan prepares the conversion of the records and unions of the boundary.
func (c *compiler) plan() {
	c.out.byName = map[string]*tinfo{}
	for name, fields := range c.src.Records {
		ti, ok := c.types[name]
		if !ok {
			continue
		}
		c.out.byName[name] = ti
		ti.rec = make([]int, len(fields))
		ti.recFields = make([]field, len(fields))
		for i, f := range fields {
			for j, tf := range ti.fields {
				if tf.name == f {
					ti.rec[i], ti.recFields[i] = j, tf
				}
			}
		}
	}
	// the other structs cross the boundary with their fields in their order, like the generated code converts them,
	// e.g. the value of a variable of a header package
	for _, ti := range c.types {
		if (ti.kind == kStruct || ti.kind == kShared) && ti.rec == nil {
			ti.rec = make([]int, len(ti.fields))
			ti.recFields = make([]field, len(ti.fields))
			for i, f := range ti.fields {
				ti.rec[i], ti.recFields[i] = i, f
			}
		}
	}
	for name, u := range c.src.Unions {
		if ti, ok := c.types[name]; ok {
			ti.handles = u.Handles
		}
		for _, v := range u.Variants {
			if ti, ok := c.types[v]; ok {
				c.out.byName[v] = ti
			}
		}
	}
}

// hostVar initializes the global of a variable of a header package by the function of the host, which returns its
// value, before the variables of the program.
func (c *compiler) hostVar(v *types.Var, g *global, pos token.Pos) {
	key := v.Pkg().Path() + "." + v.Name()
	if _, ok := c.externs[key]; ok {
		return
	}
	if !c.src.HostVars[key] {
		c.fail(pos, "%s is no variable of a header package", key)
	}
	x := &extern{name: key, idx: len(c.out.externs), results: g.t, variable: true}
	c.externs[key] = x
	c.out.externs = append(c.out.externs, x)
	c.hostInits = append(c.hostInits, &sAssign{lhs: []lval{&lGlobal{t: g.t, g: g}}, rhs: []expr{&eHost{t: g.t, ext: x}}})
}

// global returns the slots of a package variable.
func (c *compiler) global(v *types.Var) *global {
	if g, ok := c.globals[v]; ok {
		return g
	}
	t := c.typeOf(v.Type())
	// the slots follow, when the variables are divided between the program and the instances, see divideGlobals
	g := &global{t: t}
	c.globals[v] = g
	c.globalList = append(c.globalList, g)
	return g
}

// funcFor returns the function of fn, a function or method of the program, for the type arguments of its receiver
// and its own, and queues its compilation. Each combination of type arguments is a function of its own.
func (c *compiler) funcFor(fn *types.Func, recvArgs, targs []types.Type) *Func {
	origin := fn.Origin()
	key := origin.FullName()
	if len(recvArgs)+len(targs) > 0 {
		key += "[" + typeListKey(recvArgs) + ";" + typeListKey(targs) + "]"
	}
	if f, ok := c.funcs[key]; ok {
		return f
	}
	d, ok := c.decls[origin]
	if !ok || d.decl.Body == nil {
		c.fail(fn.Pos(), "%s has no body", key)
	}
	sig := origin.Signature()
	subst := map[*types.TypeParam]types.Type{}
	for i := 0; i < sig.RecvTypeParams().Len() && i < len(recvArgs); i++ {
		subst[sig.RecvTypeParams().At(i)] = recvArgs[i]
	}
	for i := 0; i < sig.TypeParams().Len() && i < len(targs); i++ {
		subst[sig.TypeParams().At(i)] = targs[i]
	}
	l := c.newFnLower(nil, d.pkg.Info)
	l.subst = subst
	l.pkg = d.pkg.Types
	f := &Func{name: key, params: l.paramTuple(sig.Recv(), sig), results: l.tupleOf(sig.Results())}
	l.f = f
	f.closure = &Closure{fn: f}
	c.funcs[key] = f
	c.out.funcs = append(c.out.funcs, f)
	c.queue = append(c.queue, func() {
		decl := d.decl
		l.cells = cellsOf(d.pkg.Info, decl.Body)
		addressed(d.pkg.Info, decl.Body, l.cells)
		var names []*ast.Ident
		if decl.Recv != nil {
			names = append(names, fieldNames(decl.Recv)...)
		}
		names = append(names, fieldNames(decl.Type.Params)...)
		for i, id := range names {
			l.param(d.pkg.Info.Defs[id], id.Name, f.params.elems[i])
		}
		l.namedResults(decl.Type.Results)
		l.stmts(decl.Body.List)
		f.body = append(l.prologue, l.body...)
		f.locals = l.frame
	})
	return f
}

func typeListKey(ts []types.Type) string {
	var s string
	for i, t := range ts {
		if i > 0 {
			s += ","
		}
		s += typeKey(t)
	}
	return s
}

// fieldNames returns the names of the parameters, with a placeholder for each unnamed one.
func fieldNames(fl *ast.FieldList) []*ast.Ident {
	var res []*ast.Ident
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			res = append(res, &ast.Ident{Name: "_"})
			continue
		}
		res = append(res, f.Names...)
	}
	return res
}

// fnLower lowers the body of a function.
type fnLower struct {
	c     *compiler
	info  *types.Info
	f     *Func
	vars  map[types.Object]*local
	frame layout
	body  []stmt
	// labels maps the labels to their loops and switches
	labels map[string]stmt
	// named are the named results
	named []*local
	// pkg is the package of the function, whose calls of entry points of other packages go through the aspect
	pkg *types.Package
	// parent is the function, which encloses a function literal, whose variables it captures
	parent *fnLower
	// caps are the captured variables of a function literal
	caps []capture
	// subst maps the type parameters to the type arguments of the instance
	subst map[*types.TypeParam]types.Type
	// cells are the variables, which live in cells, see cellsOf
	cells map[types.Object]bool
	// rng is set for the body of a range over an iterator, see rangeFunc
	rng *rangeCtx
	// prologue initializes the cells of the ranges over iterators at the beginning of the function
	prologue []stmt
	// rr, rb and rv are the cells of a return and a branch out of the body of a range and of the results
	rr, rb    *types.Var
	rv        []*types.Var
	namedObjs []types.Object
	// breaks and conts are the innermost targets of break and continue
	breaks []stmt
	conts  []*sLoop
}

func (c *compiler) newFnLower(f *Func, info *types.Info) *fnLower {
	return &fnLower{c: c, info: info, f: f, vars: map[types.Object]*local{}, labels: map[string]stmt{}}
}

func (l *fnLower) newLocal(name string, t *tinfo) *local {
	v := &local{name: name, t: t, at: l.frame}
	l.frame = l.frame.add(t.l)
	return v
}

func (l *fnLower) fail(pos token.Pos, format string, args ...any) {
	l.c.fail(pos, format, args...)
}

func (l *fnLower) typeOf(x ast.Expr) *tinfo {
	return l.ty(l.info.TypeOf(x))
}

func (l *fnLower) emit(s stmt) {
	l.body = append(l.body, s)
}

// block lowers statements into a list of their own.
func (l *fnLower) block(list []ast.Stmt) []stmt {
	saved := l.body
	l.body = nil
	l.stmts(list)
	res := l.body
	l.body = saved
	return res
}

func (l *fnLower) stmts(list []ast.Stmt) {
	for _, s := range list {
		l.stmt(s, "")
	}
}

func (l *fnLower) stmt(s ast.Stmt, label string) {
	switch s := s.(type) {
	case *ast.ExprStmt:
		l.emit(&sExpr{x: l.expr(s.X)})
	case *ast.AssignStmt:
		l.assign(s)
	case *ast.IncDecStmt:
		op := token.ADD
		if s.Tok == token.DEC {
			op = token.SUB
		}
		t := l.typeOf(s.X)
		one := &eConst{t: t, n: 1}
		if t.kind == kF64 {
			one.n = math.Float64bits(1)
		}
		lv := l.lvalue(s.X, false)
		l.emit(&sAssign{lhs: []lval{lv}, rhs: []expr{l.arith(s.Pos(), op, l.read(lv), one, t)}})
	case *ast.DeclStmt:
		g := s.Decl.(*ast.GenDecl)
		if g.Tok != token.VAR {
			return // constants and types need no code
		}
		for _, spec := range g.Specs {
			vs := spec.(*ast.ValueSpec)
			l.define(vs.Names, vs.Values, vs.Pos())
		}
	case *ast.BlockStmt:
		l.stmts(s.List)
	case *ast.IfStmt:
		l.ifStmt(s)
	case *ast.ForStmt:
		l.forStmt(s, label)
	case *ast.RangeStmt:
		l.rangeStmt(s, label)
	case *ast.SwitchStmt:
		l.switchStmt(s, label)
	case *ast.TypeSwitchStmt:
		l.typeSwitch(s, label)
	case *ast.ReturnStmt:
		l.returnStmt(s)
	case *ast.BranchStmt:
		l.branch(s)
	case *ast.LabeledStmt:
		l.stmt(s.Stmt, s.Label.Name)
	case *ast.EmptyStmt:
	default:
		l.fail(s.Pos(), "statement %T is not supported yet", s)
	}
}

// define declares the variables with their values or zero values.
func (l *fnLower) define(names []*ast.Ident, values []ast.Expr, pos token.Pos) {
	var lhs []lval
	var cells []*local
	for _, id := range names {
		obj := l.info.Defs[id]
		t := l.ty(obj.Type())
		if id.Name == "_" {
			lhs = append(lhs, &lDiscard{t: t})
			continue
		}
		v, lv := l.declare(obj, id.Name, t)
		if v.cell != nil {
			cells = append(cells, v)
		}
		lhs = append(lhs, lv)
	}
	defer l.boxCells(cells)
	var rhs []expr
	switch {
	case len(values) == 0:
		for _, lv := range lhs {
			rhs = append(rhs, l.c.zeroOrNil(lv.ltyp()))
		}
	case len(values) == 1 && len(names) > 1:
		rhs = []expr{l.expr(values[0])}
	default:
		for i, x := range values {
			rhs = append(rhs, l.val(x, lhs[i].ltyp()))
		}
	}
	l.emit(&sAssign{lhs: lhs, rhs: rhs})
}

func (l *fnLower) assign(s *ast.AssignStmt) {
	if s.Tok == token.DEFINE {
		var lhs []lval
		var cells []*local
		for _, x := range s.Lhs {
			id := x.(*ast.Ident)
			if id.Name == "_" {
				lhs = append(lhs, &lDiscard{t: nil})
				continue
			}
			if obj := l.info.Defs[id]; obj != nil {
				v, lv := l.declare(obj, id.Name, l.ty(obj.Type()))
				if v.cell != nil {
					cells = append(cells, v)
				}
				lhs = append(lhs, lv)
				continue
			}
			lhs = append(lhs, l.lvalue(x, false))
		}
		l.assignValues(lhs, s.Rhs)
		l.boxCells(cells)
		return
	}
	if s.Tok != token.ASSIGN {
		// x op= y
		op := map[token.Token]token.Token{token.ADD_ASSIGN: token.ADD, token.SUB_ASSIGN: token.SUB, token.MUL_ASSIGN: token.MUL,
			token.QUO_ASSIGN: token.QUO, token.REM_ASSIGN: token.REM, token.AND_ASSIGN: token.AND, token.OR_ASSIGN: token.OR,
			token.XOR_ASSIGN: token.XOR, token.SHL_ASSIGN: token.SHL, token.SHR_ASSIGN: token.SHR, token.AND_NOT_ASSIGN: token.AND_NOT}[s.Tok]
		lv := l.lvalue(s.Lhs[0], true)
		t := lv.ltyp()
		var y expr
		if op == token.SHL || op == token.SHR {
			y = l.shiftCount(s.Rhs[0])
		} else {
			y = l.val(s.Rhs[0], t)
		}
		l.emit(&sAssign{lhs: []lval{lv}, rhs: []expr{l.arith(s.Pos(), op, l.read(lv), y, t)}})
		return
	}
	var lhs []lval
	for _, x := range s.Lhs {
		lhs = append(lhs, l.lvalue(x, false))
	}
	l.assignValues(lhs, s.Rhs)
}

func (l *fnLower) assignValues(lhs []lval, rhs []ast.Expr) {
	if len(rhs) == 1 && len(lhs) > 1 {
		x := l.expr(rhs[0])
		for i, lv := range lhs {
			if d, ok := lv.(*lDiscard); ok && d.t == nil {
				d.t = x.typ().elems[i]
			}
		}
		l.emit(&sAssign{lhs: lhs, rhs: []expr{x}})
		return
	}
	var vals []expr
	for i, x := range rhs {
		if d, ok := lhs[i].(*lDiscard); ok {
			if d.t == nil {
				d.t = l.typeOf(x)
			}
		}
		vals = append(vals, l.val(x, lhs[i].ltyp()))
	}
	l.emit(&sAssign{lhs: lhs, rhs: vals})
}

// lvalue returns the target of an assignment. If twice is set, the target is read and written, so the pointer of
// a store is kept in a variable, if its expression has side effects.
func (l *fnLower) lvalue(x ast.Expr, twice bool) lval {
	switch x := ast.Unparen(x).(type) {
	case *ast.Ident:
		if x.Name == "_" {
			return &lDiscard{}
		}
		obj := l.info.Uses[x]
		if obj == nil {
			obj = l.info.Defs[x]
		}
		if v, ok := l.lookup(obj); ok {
			return varLval(v)
		}
		l.fail(x.Pos(), "assignment to %s is not supported", x.Name)
	case *ast.SelectorExpr:
		sel := l.info.Selections[x]
		if sel == nil || sel.Kind() != types.FieldVal {
			l.fail(x.Pos(), "assignment to %s is not supported", x.Sel.Name)
		}
		base := l.expr(x.X)
		f := l.fieldOfType(base.typ(), x.Sel.Name)
		switch b := base.(type) {
		case *eLocal:
			if b.t.kind == kStruct {
				return &lLocal{t: f.t, v: b.v, off: b.off.add(f.off)}
			}
		case *eLoad:
			if b.t.kind == kStruct {
				return &lStore{t: f.t, p: l.stable(b.p, twice), off: b.off.add(f.off)}
			}
		}
		if base.typ().kind == kPtr {
			return &lStore{t: f.t, p: l.stable(base, twice), off: f.off}
		}
		l.fail(x.Pos(), "assignment to a field of %s is not supported", base.typ())
	case *ast.StarExpr:
		p := l.expr(x.X)
		return &lStore{t: p.typ().elem, p: l.stable(p, twice), off: layout{}}
	}
	l.fail(x.Pos(), "assignment to %T is not supported", x)
	return nil
}

// stable returns p, or a variable with the value of p, if p would be evaluated twice and has side effects.
func (l *fnLower) stable(p expr, twice bool) expr {
	if !twice || pure(p) {
		return p
	}
	v := l.newLocal("$p", p.typ())
	l.emit(&sAssign{lhs: []lval{&lLocal{t: v.t, v: v}}, rhs: []expr{p}})
	return &eLocal{t: v.t, v: v}
}

// pure tells, whether x reads only variables and fields.
func pure(x expr) bool {
	switch x := x.(type) {
	case *eLocal, *eGlobal, *eConst:
		return true
	case *eLoad:
		return pure(x.p)
	case *eField:
		return pure(x.x)
	}
	return false
}

// read returns the value of the target of an assignment.
func (l *fnLower) read(lv lval) expr {
	switch lv := lv.(type) {
	case *lLocal:
		return &eLocal{t: lv.t, v: lv.v, off: lv.off}
	case *lStore:
		return &eLoad{t: lv.t, p: lv.p, off: lv.off}
	case *lGlobal:
		return &eGlobal{t: lv.t, g: lv.g, off: lv.off}
	}
	panic("vm: unreadable target")
}

func (l *fnLower) ifStmt(s *ast.IfStmt) {
	if s.Init != nil {
		l.stmt(s.Init, "")
	}
	cond := l.expr(s.Cond)
	then := l.block(s.Body.List)
	var els []stmt
	if s.Else != nil {
		els = l.block([]ast.Stmt{s.Else})
	}
	l.emit(&sIf{cond: cond, then: then, els: els})
}

func (l *fnLower) loop(label string, loop *sLoop, body func()) {
	if label != "" {
		l.labels[label] = loop
	}
	l.breaks = append(l.breaks, loop)
	l.conts = append(l.conts, loop)
	body()
	l.breaks = l.breaks[:len(l.breaks)-1]
	l.conts = l.conts[:len(l.conts)-1]
	l.emit(loop)
}

func (l *fnLower) forStmt(s *ast.ForStmt, label string) {
	if s.Init != nil {
		l.stmt(s.Init, "")
	}
	// each iteration has its own variables of the init statement, which the next one copies before the post statement
	var cells []*local
	if init, ok := s.Init.(*ast.AssignStmt); ok && init.Tok == token.DEFINE {
		for _, x := range init.Lhs {
			if v, ok := l.vars[l.info.Defs[x.(*ast.Ident)]]; ok && v.cell != nil {
				cells = append(cells, v)
			}
		}
	}
	loop := &sLoop{pos: s.Pos()}
	if s.Cond != nil {
		loop.cond = l.expr(s.Cond)
	}
	l.loop(label, loop, func() {
		loop.body = l.block(s.Body.List)
		for _, v := range cells {
			loop.post = append(loop.post, l.initVar(v, varExpr(v)))
		}
		if s.Post != nil {
			loop.post = append(loop.post, l.block([]ast.Stmt{s.Post})...)
		}
	})
}

func (l *fnLower) rangeStmt(s *ast.RangeStmt, label string) {
	if rangeOverFunc(l.info, s) {
		l.rangeFunc(s, label)
		return
	}
	xt := l.typeOf(s.X)
	define := s.Tok == token.DEFINE
	bind := func(x ast.Expr, val expr) {
		if x == nil {
			return
		}
		if id, ok := x.(*ast.Ident); ok && id.Name == "_" {
			return
		}
		if define {
			obj := l.info.Defs[x.(*ast.Ident)]
			v := l.newVar(obj, obj.Name(), val.typ())
			l.emit(l.initVar(v, val))
			return
		}
		l.emit(&sAssign{lhs: []lval{l.lvalue(x, false)}, rhs: []expr{val}})
	}
	i32 := l.ty(types.Typ[types.Int32])
	counter := func() *local {
		i := l.newLocal("$i", i32)
		l.emit(&sAssign{lhs: []lval{&lLocal{t: i32, v: i}}, rhs: []expr{&eConst{t: i32}}})
		return i
	}
	inc := func(i *local) []stmt {
		return []stmt{&sAssign{lhs: []lval{&lLocal{t: i32, v: i}}, rhs: []expr{&eBinary{t: i32, op: opAddI, x: &eLocal{t: i32, v: i}, y: &eConst{t: i32, n: 1}}}}}
	}

	switch {
	case xt.kind == kI32 || xt.kind == kU32:
		n := l.newLocal("$n", xt)
		l.emit(&sAssign{lhs: []lval{&lLocal{t: xt, v: n}}, rhs: []expr{l.expr(s.X)}})
		i := counter()
		ltop := opLtI
		if xt.kind == kU32 {
			ltop = opLtU
		}
		loop := &sLoop{pos: s.Pos(), cond: &eBinary{t: l.c.boolType(), op: ltop, x: &eLocal{t: xt, v: i}, y: &eLocal{t: xt, v: n}}, post: inc(i)}
		l.loop(label, loop, func() {
			saved := l.body
			l.body = nil
			bind(s.Key, &eLocal{t: xt, v: i})
			l.stmts(s.Body.List)
			loop.body = l.body
			l.body = saved
		})
		return
	}

	call, ok := ast.Unparen(s.X).(*ast.CallExpr)
	if !ok {
		l.fail(s.Pos(), "range over %s is not supported yet", xt)
	}
	se, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		l.fail(s.Pos(), "range over %s is not supported yet", xt)
	}
	recvT := l.typeOf(se.X)
	if recvT.kind == kMap {
		l.rangeMap(s, label, se, recvT, bind, counter, inc)
		return
	}
	if recvT.kind != kList && recvT.kind != kOut {
		l.fail(s.Pos(), "range over %s.%s is not supported yet", recvT, se.Sel.Name)
	}
	lst := l.newLocal("$l", recvT)
	l.emit(&sAssign{lhs: []lval{&lLocal{t: recvT, v: lst}}, rhs: []expr{l.expr(se.X)}})
	i := counter()
	// the range visits the elements, which the list has, when it starts, and ends early, if it shrinks, like std/list
	length := l.nativeCall(l.c.listLen(), i32, []expr{&eLocal{t: recvT, v: lst}})
	n := l.newLocal("$n", i32)
	l.emit(&sAssign{lhs: []lval{&lLocal{t: i32, v: n}}, rhs: []expr{length}})
	bt := l.c.boolType()
	cond := &eAndOr{t: bt, and: true, x: &eBinary{t: bt, op: opLtI, x: &eLocal{t: i32, v: i}, y: &eLocal{t: i32, v: n}}, y: &eBinary{t: bt, op: opLtI, x: &eLocal{t: i32, v: i}, y: length}}
	loop := &sLoop{pos: s.Pos(), cond: cond, post: inc(i)}
	elem := &eListGet{t: recvT.elem, l: &eLocal{t: recvT, v: lst}, i: &eLocal{t: i32, v: i}, unchecked: true}
	l.loop(label, loop, func() {
		saved := l.body
		l.body = nil
		switch se.Sel.Name {
		case "All":
			bind(s.Key, &eLocal{t: i32, v: i})
			bind(s.Value, elem)
		case "Values":
			bind(s.Key, elem)
		default:
			l.fail(s.Pos(), "range over %s.%s is not supported yet", recvT, se.Sel.Name)
		}
		l.stmts(s.Body.List)
		loop.body = l.body
		l.body = saved
	})
}

// rangeMap ranges over the keys, which a map has, when the loop starts, and skips the deleted ones, like std/dict.
func (l *fnLower) rangeMap(s *ast.RangeStmt, label string, se *ast.SelectorExpr, mt *tinfo, bind func(ast.Expr, expr), counter func() *local, inc func(*local) []stmt) {
	i32 := l.ty(types.Typ[types.Int32])
	targs := []*tinfo{mt.key, mt.elem}
	mv := l.newLocal("$m", mt)
	l.emit(&sAssign{lhs: []lval{&lLocal{t: mt, v: mv}}, rhs: []expr{l.expr(se.X)}})
	keysT := l.c.listOf(mt.key.t)
	keys := l.newLocal("$keys", keysT)
	l.emit(&sAssign{lhs: []lval{&lLocal{t: keysT, v: keys}}, rhs: []expr{l.nativeCall(l.c.native(dictPath+".Map.$keys", targs, s.Pos()), keysT, []expr{&eLocal{t: mt, v: mv}})}})
	i := counter()
	length := l.nativeCall(l.c.listLen(), i32, []expr{&eLocal{t: keysT, v: keys}})
	loop := &sLoop{pos: s.Pos(), cond: &eBinary{t: l.c.boolType(), op: opLtI, x: &eLocal{t: i32, v: i}, y: length}, post: inc(i)}
	l.loop(label, loop, func() {
		saved := l.body
		l.body = nil
		k := l.newLocal("$k", mt.key)
		l.emit(&sAssign{lhs: []lval{&lLocal{t: mt.key, v: k}}, rhs: []expr{&eListGet{t: mt.key, l: &eLocal{t: keysT, v: keys}, i: &eLocal{t: i32, v: i}, unchecked: true}}})
		got := tuple(mt.elem, l.c.boolType())
		v := l.newLocal("$v", got)
		get := l.nativeCall(l.c.native(dictPath+".Map.Get", targs, s.Pos()), got, []expr{&eLocal{t: mt, v: mv}, &eLocal{t: mt.key, v: k}})
		l.emit(&sAssign{lhs: []lval{&lLocal{t: got, v: v}}, rhs: []expr{get}})
		ok := &eLocal{t: l.c.boolType(), v: v, off: got.offs[1]}
		l.emit(&sIf{cond: &eUnary{t: l.c.boolType(), op: opNotB, x: ok}, then: []stmt{&sContinue{target: loop}}})
		switch se.Sel.Name {
		case "All":
			bind(s.Key, &eLocal{t: mt.key, v: k})
			bind(s.Value, &eLocal{t: mt.elem, v: v})
		case "Keys":
			bind(s.Key, &eLocal{t: mt.key, v: k})
		default:
			l.fail(s.Pos(), "range over %s.%s is not supported yet", mt, se.Sel.Name)
		}
		l.stmts(s.Body.List)
		loop.body = l.body
		l.body = saved
	})
}

func (l *fnLower) switchStmt(s *ast.SwitchStmt, label string) {
	if s.Init != nil {
		l.stmt(s.Init, "")
	}
	var tag expr
	if s.Tag != nil {
		x := l.expr(s.Tag)
		v := l.newLocal("$tag", x.typ())
		l.emit(&sAssign{lhs: []lval{&lLocal{t: v.t, v: v}}, rhs: []expr{x}})
		tag = &eLocal{t: v.t, v: v}
	}
	block := &sBlock{}
	if label != "" {
		l.labels[label] = block
	}
	l.breaks = append(l.breaks, block)
	defer func() { l.breaks = l.breaks[:len(l.breaks)-1] }()

	// the cases become a chain of ifs, the default comes last
	var chain []stmt
	var dflt []stmt
	hasDefault := false
	var ifs []*sIf
	for _, cs := range s.Body.List {
		cc := cs.(*ast.CaseClause)
		body := l.block(cc.Body)
		if cc.List == nil {
			dflt, hasDefault = body, true
			continue
		}
		var cond expr
		for _, x := range cc.List {
			var c expr
			if tag == nil {
				c = l.expr(x)
			} else {
				c = l.compare(x.Pos(), token.EQL, tag, l.val(x, tag.typ()))
			}
			if cond == nil {
				cond = c
			} else {
				cond = &eAndOr{t: l.c.boolType(), x: cond, y: c}
			}
		}
		ifs = append(ifs, &sIf{cond: cond, then: body})
	}
	for i := len(ifs) - 1; i >= 0; i-- {
		if i == len(ifs)-1 {
			if hasDefault {
				ifs[i].els = dflt
			}
		} else {
			ifs[i].els = []stmt{ifs[i+1]}
		}
	}
	switch {
	case len(ifs) > 0:
		chain = []stmt{ifs[0]}
	case hasDefault:
		chain = dflt
	}
	block.body = chain
	l.emit(block)
}

func (l *fnLower) returnStmt(s *ast.ReturnStmt) {
	if l.rng != nil {
		l.rangeReturn(s)
		return
	}
	res := l.f.results
	if len(s.Results) == 0 && len(l.named) > 0 {
		var vals []expr
		for _, v := range l.named {
			vals = append(vals, varExpr(v))
		}
		l.emit(&sReturn{vals: vals})
		return
	}
	if len(s.Results) == 1 && res.kind == kTuple && len(res.elems) > 1 {
		l.emit(&sReturn{vals: []expr{l.expr(s.Results[0])}})
		return
	}
	var vals []expr
	for i, x := range s.Results {
		t := res
		if res.kind == kTuple {
			t = res.elems[i]
		}
		vals = append(vals, l.val(x, t))
	}
	l.emit(&sReturn{vals: vals})
}

func (l *fnLower) branch(s *ast.BranchStmt) {
	if s.Tok != token.BREAK && s.Tok != token.CONTINUE {
		l.fail(s.Pos(), "%s is not supported", s.Tok)
	}
	label := ""
	if s.Label != nil {
		label = s.Label.Name
	}
	l.branchTo(s.Tok, label, s.Pos())
}

func (c *compiler) boolType() *tinfo {
	return c.typeOf(types.Typ[types.Bool])
}

// zero returns the zero value of t.
func (c *compiler) zero(t *tinfo, pos token.Pos) expr {
	switch t.kind {
	case kBool, kI32, kU32, kF64, kStr, kError, kIface, kOpt:
		return &eConst{t: t}
	case kStruct:
		var fields []expr
		var order []int
		for i, f := range t.fields {
			fields = append(fields, c.zero(f.t, pos))
			order = append(order, i)
		}
		return &eStruct{t: t, fields: fields, order: order}
	case kShared:
		return &eConst{t: t, r: c.zeroObject(t, pos)}
	}
	c.fail(pos, "the zero value of %s is not supported", t)
	return nil
}

// zeroOrNil returns the zero value of t, which is nil for a reference, e.g. of a named result of type list.List,
// which a checked program assigns before it returns it.
func (c *compiler) zeroOrNil(t *tinfo) expr {
	switch t.kind {
	case kStruct, kShared, kBool, kI32, kU32, kF64, kStr:
		return c.zero(t, token.NoPos)
	}
	return &eConst{t: t}
}

// zeroObject returns the zero value of a shared struct, which all its zero values share.
func (c *compiler) zeroObject(t *tinfo, pos token.Pos) *object {
	if t.zero == nil {
		t.zero = newObject(t, t.obj)
		c.zeroSlotsInto(t, t.zero.r)
	}
	return t.zero
}

// zeroSlotsInto writes the shared structs of the zero value of the struct t into its references.
func (c *compiler) zeroSlotsInto(t *tinfo, r []any) {
	for _, f := range t.fields {
		switch f.t.kind {
		case kShared:
			r[f.off.r] = c.zeroObject(f.t, token.NoPos)
		case kStruct:
			c.zeroSlotsInto(f.t, r[f.off.r:])
		}
	}
}

// constant returns the value of a constant of type t.
func (c *compiler) constant(v constant.Value, t *tinfo) expr {
	switch t.kind {
	case kBool:
		if constant.BoolVal(v) {
			return &eConst{t: t, n: 1}
		}
		return &eConst{t: t}
	case kI32:
		n, _ := constant.Int64Val(constant.ToInt(v))
		return &eConst{t: t, n: uint64(n)}
	case kU32:
		n, _ := constant.Uint64Val(constant.ToInt(v))
		return &eConst{t: t, n: n}
	case kF64:
		f, _ := constant.Float64Val(constant.ToFloat(v))
		return &eConst{t: t, n: math.Float64bits(f)}
	case kStr:
		return &eConst{t: t, s: constant.StringVal(v)}
	}
	panic(fmt.Sprintf("vm: constant of %s", t))
}

// val returns the value of x as a value of type to: nil as the zero value and a concrete value boxed in an
// interface.
func (l *fnLower) val(x ast.Expr, to *tinfo) expr {
	if tv := l.info.Types[x]; tv.IsNil() {
		return &eConst{t: to}
	}
	e := l.expr(x)
	return l.convert(x.Pos(), e, to)
}

func (l *fnLower) convert(pos token.Pos, e expr, to *tinfo) expr {
	from := e.typ()
	if from == to || from.kind == kTuple {
		return e
	}
	if (to.kind == kIface || to.kind == kError) && from.kind != kIface && from.kind != kError {
		return l.box(pos, e, to)
	}
	return e
}

func (l *fnLower) expr(x ast.Expr) expr {
	tv := l.info.Types[x]
	if tv.Value != nil {
		return l.c.constant(tv.Value, l.ty(tv.Type))
	}
	switch x := x.(type) {
	case *ast.ParenExpr:
		return l.expr(x.X)
	case *ast.Ident:
		return l.ident(x)
	case *ast.SelectorExpr:
		return l.selector(x)
	case *ast.BinaryExpr:
		return l.binary(x)
	case *ast.UnaryExpr:
		return l.unary(x)
	case *ast.CallExpr:
		return l.call(x)
	case *ast.CompositeLit:
		return l.composite(x)
	case *ast.FuncLit:
		return l.funcLit(x)
	case *ast.StarExpr:
		p := l.expr(x.X)
		t := p.typ().elem
		if t.kind == kShared {
			// the pointer refers to the shared object itself
			return &eField{t: t, x: p}
		}
		return &eLoad{t: t, p: p}
	}
	l.fail(x.Pos(), "expression %T is not supported yet", x)
	return nil
}

func (l *fnLower) ident(id *ast.Ident) expr {
	obj := l.info.Uses[id]
	if obj == nil {
		obj = l.info.Defs[id]
	}
	switch obj := obj.(type) {
	case *types.Var:
		if v, ok := l.lookup(obj); ok {
			return varExpr(v)
		}
		if obj.Parent() == obj.Pkg().Scope() {
			g := l.c.global(obj)
			if p := l.c.pkgs[obj.Pkg().Path()]; p == nil || !p.User {
				l.c.hostVar(obj, g, id.Pos())
			}
			return &eGlobal{t: g.t, g: g}
		}
	case *types.Func:
		var targs []types.Type
		if inst, ok := l.info.Instances[id]; ok {
			for t := range inst.TypeArgs.Types() {
				targs = append(targs, l.sub(t))
			}
		}
		if !l.c.user(obj) {
			return l.libValue(id, obj, targs)
		}
		return &eFunc{t: l.typeOf(id), fn: l.funcOf(obj, nil, targs)}
	}
	l.fail(id.Pos(), "identifier %s is not supported", id.Name)
	return nil
}

func (l *fnLower) selector(x *ast.SelectorExpr) expr {
	sel := l.info.Selections[x]
	if sel == nil {
		// a qualified identifier of another package
		return l.ident(x.Sel)
	}
	if sel.Kind() == types.MethodVal {
		return l.methodValue(x, sel)
	}
	return l.field(l.expr(x.X), x.Sel.Name)
}

func (l *fnLower) fieldOfType(t *tinfo, name string) field {
	return t.fieldOf(name)
}

// field selects a field of a struct, a pointer or a shared struct.
func (l *fnLower) field(base expr, name string) expr {
	bt := base.typ()
	f := bt.fieldOf(name)
	if bt.kind == kPtr || bt.kind == kShared {
		return &eLoad{t: f.t, p: base, off: f.off}
	}
	switch b := base.(type) {
	case *eLocal:
		return &eLocal{t: f.t, v: b.v, off: b.off.add(f.off)}
	case *eGlobal:
		return &eGlobal{t: f.t, g: b.g, off: b.off.add(f.off)}
	case *eLoad:
		return &eLoad{t: f.t, p: b.p, off: b.off.add(f.off)}
	case *eField:
		return &eField{t: f.t, x: b.x, off: b.off.add(f.off)}
	}
	return &eField{t: f.t, x: base, off: f.off}
}

func (l *fnLower) binary(x *ast.BinaryExpr) expr {
	switch x.Op {
	case token.LAND, token.LOR:
		return &eAndOr{t: l.c.boolType(), and: x.Op == token.LAND, x: l.expr(x.X), y: l.expr(x.Y)}
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		xn, yn := l.info.Types[x.X].IsNil(), l.info.Types[x.Y].IsNil()
		if xn || yn {
			other := x.X
			if xn {
				other = x.Y
			}
			isNil := &eUnary{t: l.c.boolType(), op: opIsNil, x: l.expr(other)}
			if x.Op == token.NEQ {
				return &eUnary{t: l.c.boolType(), op: opNotB, x: isNil}
			}
			return isNil
		}
		a := l.expr(x.X)
		return l.compare(x.Pos(), x.Op, a, l.val(x.Y, a.typ()))
	case token.SHL, token.SHR:
		return l.arith(x.Pos(), x.Op, l.expr(x.X), l.shiftCount(x.Y), l.typeOf(x))
	}
	t := l.typeOf(x)
	return l.arith(x.Pos(), x.Op, l.val(x.X, t), l.val(x.Y, t), t)
}

// shiftCount returns the count of a shift, which the subset requires to be constant.
func (l *fnLower) shiftCount(x ast.Expr) expr {
	tv := l.info.Types[x]
	if tv.Value == nil {
		l.fail(x.Pos(), "shift by a variable count")
	}
	n, _ := constant.Uint64Val(constant.ToInt(tv.Value))
	return &eConst{t: l.ty(types.Typ[types.Uint32]), n: n}
}

// compare compares two values of the same type.
func (l *fnLower) compare(pos token.Pos, tok token.Token, a, b expr) expr {
	bt := l.c.boolType()
	t := a.typ()
	swap := false
	switch tok {
	case token.GTR:
		tok, swap = token.LSS, true
	case token.GEQ:
		tok, swap = token.LEQ, true
	}
	var o op
	switch t.kind {
	case kBool, kI32:
		o = map[token.Token]op{token.EQL: opEqI, token.NEQ: opNeI, token.LSS: opLtI, token.LEQ: opLeI}[tok]
	case kU32:
		o = map[token.Token]op{token.EQL: opEqI, token.NEQ: opNeI, token.LSS: opLtU, token.LEQ: opLeU}[tok]
	case kF64:
		o = map[token.Token]op{token.EQL: opEqF, token.NEQ: opNeF, token.LSS: opLtF, token.LEQ: opLeF}[tok]
	case kStr:
		o = map[token.Token]op{token.EQL: opEqS, token.NEQ: opNeS, token.LSS: opLtS, token.LEQ: opLeS}[tok]
	case kPtr, kHandle:
		o = map[token.Token]op{token.EQL: opEqR, token.NEQ: opNeR}[tok]
	case kStruct, kShared, kIface:
		eq := &eBinary{t: bt, op: opEqV, x: a, y: b}
		if tok == token.NEQ {
			return &eUnary{t: bt, op: opNotB, x: eq}
		}
		return eq
	default:
		l.fail(pos, "comparison of %s is not supported yet", t)
	}
	return &eBinary{t: bt, op: o, x: a, y: b, swap: swap}
}

func (l *fnLower) arith(pos token.Pos, tok token.Token, a, b expr, t *tinfo) expr {
	var o op
	switch t.kind {
	case kI32, kU32:
		signed := t.kind == kI32
		switch tok {
		case token.ADD:
			o = opAddI
		case token.SUB:
			o = opSubI
		case token.MUL:
			o = opMulI
		case token.QUO:
			o = map[bool]op{true: opDivI, false: opDivU}[signed]
		case token.REM:
			o = map[bool]op{true: opRemI, false: opRemU}[signed]
		case token.AND:
			o = opAndI
		case token.OR:
			o = opOrI
		case token.XOR:
			o = opXorI
		case token.AND_NOT:
			o = opAndNotI
		case token.SHL, token.SHR:
			cv, ok := b.(*eConst)
			if !ok {
				l.fail(pos, "shift by a variable count")
			}
			o = opShlI
			if tok == token.SHR {
				o = map[bool]op{true: opShrI, false: opShrU}[signed]
			}
			return &eBinary{t: t, op: o, x: a, y: &eConst{t: t, n: cv.n}, imm: true}
		}
	case kF64:
		o = map[token.Token]op{token.ADD: opAddF, token.SUB: opSubF, token.MUL: opMulF, token.QUO: opDivF}[tok]
	case kStr:
		if tok == token.ADD {
			o = opConcat
		}
	}
	if o == opNop {
		l.fail(pos, "operator %s on %s is not supported yet", tok, t)
	}
	return &eBinary{t: t, op: o, x: a, y: b}
}

func (l *fnLower) unary(x *ast.UnaryExpr) expr {
	t := l.typeOf(x)
	switch x.Op {
	case token.SUB:
		e := l.expr(x.X)
		if t.kind == kF64 {
			return &eUnary{t: t, op: opNegF, x: e}
		}
		return &eUnary{t: t, op: opNegI, x: e}
	case token.ADD:
		return l.expr(x.X)
	case token.NOT:
		return &eUnary{t: t, op: opNotB, x: l.expr(x.X)}
	case token.XOR:
		return &eUnary{t: t, op: opNotI, x: l.expr(x.X)}
	case token.AND:
		lit, ok := ast.Unparen(x.X).(*ast.CompositeLit)
		if !ok {
			l.fail(x.Pos(), "& of %T is not supported", x.X)
		}
		v := l.composite(lit)
		if v.typ().kind == kShared {
			// the shared object is the pointer
			return &eField{t: t, x: v}
		}
		return &eNew{t: t, x: v}
	}
	l.fail(x.Pos(), "operator %s is not supported", x.Op)
	return nil
}

func (l *fnLower) composite(lit *ast.CompositeLit) expr {
	t := l.typeOf(lit)
	if t.kind != kStruct && t.kind != kShared {
		l.fail(lit.Pos(), "composite literal of %s is not supported", t)
	}
	vals := make([]expr, len(t.fields))
	var order []int
	for i, elt := range lit.Elts {
		idx := i
		x := elt
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			name := kv.Key.(*ast.Ident).Name
			for j, f := range t.fields {
				if f.name == name {
					idx = j
				}
			}
			x = kv.Value
		}
		vals[idx] = l.val(x, t.fields[idx].t)
		order = append(order, idx)
	}
	for i, v := range vals {
		if v == nil {
			vals[i] = l.c.zero(t.fields[i].t, lit.Pos())
			order = append(order, i)
		}
	}
	return &eStruct{t: t, fields: vals, order: order}
}

func (l *fnLower) call(call *ast.CallExpr) expr {
	if tv, ok := l.info.Types[call.Fun]; ok && tv.IsType() {
		return l.conversion(call, l.ty(tv.Type))
	}
	fun := ast.Unparen(call.Fun)
	if id, ok := fun.(*ast.Ident); ok {
		if b, ok := l.info.Uses[id].(*types.Builtin); ok {
			return l.builtin(b, call)
		}
	}
	sig := l.sub(l.info.TypeOf(call.Fun)).Underlying().(*types.Signature)

	base := fun
	switch x := fun.(type) {
	case *ast.IndexExpr:
		base = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		base = ast.Unparen(x.X)
	}
	var id *ast.Ident
	switch x := base.(type) {
	case *ast.SelectorExpr:
		if sel := l.info.Selections[x]; sel != nil && sel.Kind() == types.MethodVal {
			return l.methodCall(x, sel, call)
		}
		if sel := l.info.Selections[x]; sel == nil {
			id = x.Sel
		}
	case *ast.Ident:
		id = x
	}
	if id != nil {
		if fn, ok := l.info.Uses[id].(*types.Func); ok {
			var targs []types.Type
			if inst, ok := l.info.Instances[id]; ok {
				for t := range inst.TypeArgs.Types() {
					targs = append(targs, l.sub(t))
				}
			}
			return l.funcCall(fn, nil, targs, call, sig, nil)
		}
	}
	f := l.expr(call.Fun)
	return &eCallValue{t: l.tupleOf(sig.Results()), f: f, args: l.args(call, sig, nil)}
}

// args lowers the arguments of a call, the variadic ones as a list.
func (l *fnLower) args(call *ast.CallExpr, sig *types.Signature, recv expr) []expr {
	var res []expr
	if recv != nil {
		res = append(res, recv)
	}
	params := sig.Params()
	if len(call.Args) == 1 && isTuple(l.info.TypeOf(call.Args[0])) {
		// f(g()): the results of g are laid out like the parameters of f, unless f is variadic and needs a list
		x := l.expr(call.Args[0])
		for i, et := range x.typ().elems {
			if sig.Variadic() && i >= params.Len()-1 {
				l.fail(call.Pos(), "f(g()) of a variadic f is not supported yet")
			}
			if pt := l.ty(params.At(i).Type()); pt != et && (pt.kind == kIface || pt.kind == kError) {
				l.fail(call.Pos(), "f(g()), which converts the results of g, is not supported yet")
			}
		}
		return append(res, x)
	}
	var rest []expr
	for i, a := range call.Args {
		if sig.Variadic() && i >= params.Len()-1 {
			elem := l.ty(params.At(params.Len() - 1).Type().(*types.Slice).Elem())
			rest = append(rest, l.val(a, elem))
			continue
		}
		res = append(res, l.val(a, l.ty(params.At(i).Type())))
	}
	if sig.Variadic() {
		lt := l.c.listOf(params.At(params.Len() - 1).Type().(*types.Slice).Elem())
		res = append(res, &eListOf{t: lt, elems: rest})
	}
	return res
}

func isTuple(t types.Type) bool {
	_, ok := t.(*types.Tuple)
	return ok
}

// listOf returns the type of a list of elements of type t, which a variadic parameter receives.
func (c *compiler) listOf(t types.Type) *tinfo {
	key := "$variadic " + typeKey(t)
	if ti, ok := c.types[key]; ok {
		return ti
	}
	ti := &tinfo{t: types.NewSlice(t), kind: kList, l: layout{r: 1}, elem: c.typeOf(t)}
	c.types[key] = ti
	return ti
}

func (l *fnLower) funcCall(fn *types.Func, recvArgs, targs []types.Type, call *ast.CallExpr, sig *types.Signature, recv expr) expr {
	if !l.c.user(fn) {
		if key := libKey(fn); (key == listPath+".List.Append" || key == msgPath+".Out.Send") && recv != nil {
			var elems []expr
			for _, a := range call.Args {
				elems = append(elems, l.val(a, recv.typ().elem))
			}
			return &eListAppend{t: l.tupleOf(sig.Results()), l: recv, elems: elems}
		}
		return l.libCall(fn, recvArgs, targs, sig, call.Pos(), recv, l.args(call, sig, recv))
	}
	return &eCall{t: l.tupleOf(sig.Results()), fn: l.funcOf(fn, recvArgs, targs), args: l.args(call, sig, recv)}
}

// user reports whether fn belongs to a package of the program, whose body l compiles, and not to the library or a
// header package.
func (c *compiler) user(fn *types.Func) bool {
	p := c.pkgs[fn.Pkg().Path()]
	return p != nil && p.User
}

// libKey names a function or method of the library or a header package, e.g. strings.TrimSpace.
func libKey(fn *types.Func) string {
	path := fn.Pkg().Path()
	if r := fn.Signature().Recv(); r != nil {
		return path + "." + recvName(r.Type()) + "." + fn.Name()
	}
	return path + "." + fn.Name()
}

// libCall calls fn of the library or a header package with the lowered arguments: its native or, for a function of a
// header package, the host.
func (l *fnLower) libCall(fn *types.Func, recvArgs, targs []types.Type, sig *types.Signature, pos token.Pos, recv expr, args []expr) expr {
	key := libKey(fn)
	res := l.tupleOf(sig.Results())
	var ntargs []*tinfo
	for _, t := range append(append([]types.Type(nil), recvArgs...), targs...) {
		ntargs = append(ntargs, l.ty(t))
	}
	if nat := l.c.native(key, ntargs, pos); nat != nil {
		if n := sig.Results().Len(); n > 0 && isErrorType(sig.Results().At(n-1).Type()) && l.c.src.ErrorPos != nil {
			// a function of the library, which creates an error, e.g. errors.New, keeps the position of the call
			at := res.l.r - 1
			nat = errorNative(nat, at, l.c.src.ErrorPos(pos))
		}
		return l.nativeCall(nat, res, args)
	}
	if l.c.pkgs[fn.Pkg().Path()] == nil {
		l.fail(pos, "%s is not supported yet", key)
	}
	if len(targs) > 0 {
		key += "[" + hostTypeArgs(targs) + "]"
	}
	return &eHost{t: res, ext: l.c.extern(key, recv, sig, pos), args: args}
}

// libValue returns a function of the library or a header package as a value, e.g. f := strings.TrimSpace, as a
// function literal, which calls it like a direct call does.
func (l *fnLower) libValue(id *ast.Ident, fn *types.Func, targs []types.Type) expr {
	sig := l.sub(l.info.TypeOf(id)).(*types.Signature)
	inner := l.child(sig, "lib")
	var args []expr
	for _, pt := range inner.f.params.elems {
		p := inner.newLocal("p", pt)
		args = append(args, &eLocal{t: pt, v: p})
	}
	inner.returnCall(inner.libCall(fn, nil, targs, sig, id.Pos(), nil, args))
	return inner.finish(l.typeOf(id))
}

// returnCall returns the results of call, the body of a function literal, which wraps a call.
func (l *fnLower) returnCall(call expr) {
	if res := call.typ(); res.l == (layout{}) && res.kind == kTuple && len(res.elems) == 0 {
		l.emit(&sExpr{x: call})
		l.emit(&sReturn{})
		return
	}
	l.emit(&sReturn{vals: []expr{call}})
}

// funcOf returns the function, which l calls for fn: the hook of an entry point of another package, see hookFor.
func (l *fnLower) funcOf(fn *types.Func, recvArgs, targs []types.Type) *Func {
	if _, ok := l.c.src.Entries[fn]; ok && fn.Pkg() != l.pkg {
		return l.c.hookFor(fn)
	}
	return l.c.funcFor(fn, recvArgs, targs)
}

// hostTypeArgs names the type arguments of an instance of a function of a header package like the bundle: a type
// argument by its core type, so that a named type like type ID string uses the function of string.
func hostTypeArgs(targs []types.Type) string {
	var parts []string
	for _, t := range targs {
		if !isHandle(t) {
			t = t.Underlying()
		}
		parts = append(parts, typeKey(t))
	}
	return strings.Join(parts, ",")
}

// nativeCall calls a native with the arguments, which are laid out like a tuple of their types.
func (l *fnLower) nativeCall(nat *native, res *tinfo, args []expr) expr {
	var ts []*tinfo
	for _, a := range args {
		ts = append(ts, a.typ())
	}
	return &eNative{t: res, nat: nat, params: tuple(ts...), args: args}
}

func isErrorType(t types.Type) bool {
	return types.Identical(t, errorType)
}

func recvName(t types.Type) string {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := types.Unalias(t).(*types.Named); ok {
		return n.Obj().Name()
	}
	return types.TypeString(t, nil)
}

func (l *fnLower) methodCall(se *ast.SelectorExpr, sel *types.Selection, call *ast.CallExpr) expr {
	fn := sel.Obj().(*types.Func)
	sig := l.sub(l.info.TypeOf(se)).(*types.Signature)
	recvT := l.typeOf(se.X)
	switch recvT.kind {
	case kIface, kError:
		return l.ifaceCall(se, call)
	}
	var recvArgs []types.Type
	if n, ok := types.Unalias(derefType(recvT.t)).(*types.Named); ok {
		for t := range n.TypeArgs().Types() {
			recvArgs = append(recvArgs, t)
		}
	}
	var targs []types.Type
	if inst, ok := l.info.Instances[se.Sel]; ok {
		for t := range inst.TypeArgs.Types() {
			targs = append(targs, l.sub(t))
		}
	}
	// the receiver as the method takes it
	var recv expr
	_, ptrRecv := fn.Signature().Recv().Type().(*types.Pointer)
	switch {
	case ptrRecv && recvT.kind != kPtr && recvT.kind != kHandle && recvT.kind != kList && recvT.kind != kMap && recvT.kind != kOut && recvT.kind != kOpt && recvT.kind != kBig:
		recv = l.addr(se.X)
	case !ptrRecv && recvT.kind == kPtr:
		recv = l.deref(l.expr(se.X))
	default:
		recv = l.expr(se.X)
	}
	return l.funcCall(fn, recvArgs, targs, call, sig, recv)
}

func derefType(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}

func (l *fnLower) builtin(b *types.Builtin, call *ast.CallExpr) expr {
	switch b.Name() {
	case "min", "max":
		t := l.typeOf(call)
		res := l.val(call.Args[0], t)
		for _, a := range call.Args[1:] {
			res = l.nativeCall(l.c.native(b.Name()+"."+kindName(t.kind), nil, call.Pos()), t, []expr{res, l.val(a, t)})
		}
		return res
	}
	l.fail(call.Pos(), "builtin %s is not supported", b.Name())
	return nil
}

func kindName(k kind) string {
	switch k {
	case kI32:
		return "int32"
	case kU32:
		return "uint32"
	case kF64:
		return "float64"
	case kStr:
		return "string"
	}
	return "?"
}

func (l *fnLower) conversion(call *ast.CallExpr, to *tinfo) expr {
	a := l.expr(call.Args[0])
	from := a.typ()
	switch {
	case to.kind == kIface || to.kind == kError:
		return l.convert(call.Pos(), a, to)
	case to.kind == kF64 && from.kind == kI32:
		return &eUnary{t: to, op: opI2F, x: a}
	case to.kind == kF64 && from.kind == kU32:
		return &eUnary{t: to, op: opU2F, x: a}
	case (to.kind == kI32 || to.kind == kU32) && from.kind == kF64:
		l.fail(call.Pos(), "conversion of float64 to %s", to)
	}
	// the representation stays, a 32 bit integer is read in the width of its type
	return retype(a, to)
}

// retype returns x as a value of the type t with the same slots.
func retype(x expr, t *tinfo) expr {
	switch x := x.(type) {
	case *eConst:
		return &eConst{t: t, n: x.n, s: x.s}
	case *eLocal:
		return &eLocal{t: t, v: x.v, off: x.off}
	case *eGlobal:
		return &eGlobal{t: t, g: x.g, off: x.off}
	case *eLoad:
		return &eLoad{t: t, p: x.p, off: x.off}
	}
	return &eField{t: t, x: x}
}

// extern returns the function of a header package, which the host implements.
// extern returns the function of a header package, which the host implements, with the parameters of the instance:
// the receiver, if any, and the parameters of sig, the variadic ones as a list.
func (c *compiler) extern(key string, recv expr, sig *types.Signature, pos token.Pos) *extern {
	if x, ok := c.externs[key]; ok {
		return x
	}
	x := &extern{name: key, idx: len(c.out.externs)}
	if !c.src.Hosted[key] {
		c.fail(pos, "%s is no function of a header package, which the host implements", key)
	}
	if recv != nil {
		x.params = append(x.params, recv.typ())
	}
	for i := range sig.Params().Len() {
		t := sig.Params().At(i).Type()
		if sig.Variadic() && i == sig.Params().Len()-1 {
			x.params = append(x.params, c.listOf(t.(*types.Slice).Elem()))
			continue
		}
		x.params = append(x.params, c.typeOf(t))
	}
	x.results = c.tupleOf(sig.Results())
	c.externs[key] = x
	c.out.externs = append(c.out.externs, x)
	return x
}

func isStd(path string) bool {
	return strings.HasPrefix(path, stdPrefix) || !strings.Contains(path, ".")
}
