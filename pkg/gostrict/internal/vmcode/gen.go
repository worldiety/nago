// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"fmt"
	"math"
)

// gen translates the IR of a function into register bytecode. The parameters and variables occupy the first
// slots of the frame, the temporaries the slots above them, and the arguments of a call the slots above the
// temporaries, where the frame of the callee begins.
type gen struct {
	c    *compiler
	f    *Func
	code []instr
	top  layout
	max  layout
	// jumps are the jumps to patch for each loop and block
	breaks map[stmt][]int
	conts  map[stmt][]int
}

func (g *gen) function() {
	f := g.f
	g.breaks, g.conts = map[stmt][]int{}, map[stmt][]int{}
	g.top = f.locals.max(f.params.l)
	g.max = g.top.max(f.results.l)
	g.stmts(f.body)
	g.emit(opRet, 0, 0, 0)
	f.frame = g.max
	f.code = g.code
}

func (g *gen) emit(op op, a, b, c int32) int {
	g.code = append(g.code, instr{op: op, a: a, b: b, c: c})
	return len(g.code) - 1
}

func (g *gen) pc() int32 {
	return int32(len(g.code))
}

func (g *gen) patch(at int, target int32) {
	g.code[at].a = target
}

func (g *gen) alloc(l layout) layout {
	at := g.top
	g.top = g.top.add(l)
	g.max = g.max.max(g.top)
	return at
}

func (g *gen) stmts(list []stmt) {
	for _, s := range list {
		saved := g.top
		g.stmt(s)
		g.top = saved
	}
}

func (g *gen) stmt(s stmt) {
	switch s := s.(type) {
	case *sExpr:
		g.expr(s.x)
	case *sAssign:
		g.assign(s)
	case *sIf:
		g.ifStmt(s)
	case *sLoop:
		g.loop(s)
	case *sBlock:
		g.stmts(s.body)
		for _, at := range g.breaks[s] {
			g.patch(at, g.pc())
		}
	case *sBreak:
		g.breaks[s.target] = append(g.breaks[s.target], g.emit(opJmp, 0, 0, 0))
	case *sContinue:
		g.conts[s.target] = append(g.conts[s.target], g.emit(opJmp, 0, 0, 0))
	case *sReturn:
		g.ret(s)
	default:
		panic(fmt.Sprintf("vm: statement %T", s))
	}
}

func (g *gen) ifStmt(s *sIf) {
	toElse := g.condJump(s.cond, false)
	g.stmts(s.then)
	if len(s.els) == 0 {
		g.patchAll(toElse, g.pc())
		return
	}
	toEnd := g.emit(opJmp, 0, 0, 0)
	g.patchAll(toElse, g.pc())
	g.stmts(s.els)
	g.patch(toEnd, g.pc())
}

func (g *gen) patchAll(at []int, target int32) {
	for _, a := range at {
		g.patch(a, target)
	}
}

func (g *gen) loop(s *sLoop) {
	start := g.pc()
	var exits []int
	if s.cond != nil {
		exits = g.condJump(s.cond, false)
	}
	g.stmts(s.body)
	g.patchAll(g.conts[s], g.pc())
	g.stmts(s.post)
	g.emit(opLoop, start, 0, 0)
	g.patchAll(exits, g.pc())
	g.patchAll(g.breaks[s], g.pc())
}

// condJump emits the jumps, which are taken, if cond is the given value, and returns them for patching.
func (g *gen) condJump(cond expr, when bool) []int {
	switch e := cond.(type) {
	case *eUnary:
		if e.op == opNotB {
			return g.condJump(e.x, !when)
		}
	case *eAndOr:
		// a && b jumps, if false, when a or b is false; if true, when both are true
		if e.and != when {
			a := g.condJump(e.x, when)
			return append(a, g.condJump(e.y, when)...)
		}
		skip := g.condJump(e.x, !when)
		res := g.condJump(e.y, when)
		g.patchAll(skip, g.pc())
		return res
	case *eBinary:
		if !when {
			if fused, ok := fusedJump[e.op]; ok {
				saved := g.top
				x, y := g.expr(e.x), g.expr(e.y)
				g.top = saved
				a, b := x.n, y.n
				if e.swap {
					a, b = b, a
				}
				return []int{g.emit(fused, 0, a, b)}
			}
		}
	case *eConst:
		if (e.n != 0) == when {
			return []int{g.emit(opJmp, 0, 0, 0)}
		}
		return nil
	}
	saved := g.top
	v := g.expr(cond)
	g.top = saved
	if when {
		return []int{g.emit(opJmpT, 0, v.n, 0)}
	}
	return []int{g.emit(opJmpF, 0, v.n, 0)}
}

var fusedJump = map[op]op{opLtI: opJNltI, opLeI: opJNleI, opEqI: opJNeqI, opNeI: opJNneI, opLtF: opJNltF, opLeF: opJNleF}

func (g *gen) ret(s *sReturn) {
	res := g.f.results
	if len(s.vals) == 1 && (res.kind == kTuple || !hazard(s.vals[0], res)) {
		g.exprTo(s.vals[0], layout{})
		g.emit(opRet, 0, 0, 0)
		return
	}
	var at []layout
	for _, v := range s.vals {
		t := g.alloc(v.typ().l)
		g.exprTo(v, t)
		at = append(at, t)
	}
	for i, v := range s.vals {
		off := layout{}
		if res.kind == kTuple {
			off = res.offs[i]
		}
		g.move(off, at[i], v.typ().l)
	}
	g.emit(opRet, 0, 0, 0)
}

// hazard tells, whether evaluating x into slots may write them, before x has read all its operands, e.g. a
// struct, which is written field by field, or a && b, which writes a first. Then x must not be evaluated into the
// slots of a variable, which it reads.
func hazard(x expr, res *tinfo) bool {
	switch x := x.(type) {
	case *eConst, *eLocal, *eGlobal, *eCall, *eNative, *eHost, *eCallValue, *eBinary, *eUnary, *eFunc, *eBox, *eNew, *eListGet:
		return false
	case *eLoad:
		return x.t.l.size() > 1
	}
	return true
}

func (g *gen) assign(s *sAssign) {
	if len(s.lhs) == 1 && len(s.rhs) == 1 {
		switch lv := s.lhs[0].(type) {
		case *lLocal:
			if hazard(s.rhs[0], lv.t) && mentions(s.rhs[0], lv.v) {
				v := g.expr(s.rhs[0])
				g.move(lv.v.at.add(lv.off), v, lv.t.l)
				return
			}
			g.exprTo(s.rhs[0], lv.v.at.add(lv.off))
			return
		}
	}
	// the pointers of the targets, then the values, then the assignments
	ptrs := make([]layout, len(s.lhs))
	for i, lv := range s.lhs {
		if st, ok := lv.(*lStore); ok {
			ptrs[i] = g.expr(st.p)
		}
	}
	var vals []layout
	if len(s.rhs) == 1 && len(s.lhs) > 1 {
		t := s.rhs[0].typ()
		v := g.expr(s.rhs[0])
		for i := range s.lhs {
			vals = append(vals, v.add(t.offs[i]))
		}
	} else {
		for _, x := range s.rhs {
			if len(s.rhs) == 1 {
				vals = append(vals, g.expr(x))
				continue
			}
			t := g.alloc(x.typ().l)
			g.exprTo(x, t)
			vals = append(vals, t)
		}
	}
	for i, lv := range s.lhs {
		g.store(lv, ptrs[i], vals[i])
	}
}

// store writes the value at src into the target.
func (g *gen) store(lv lval, ptr, src layout) {
	switch lv := lv.(type) {
	case *lLocal:
		g.move(lv.v.at.add(lv.off), src, lv.t.l)
	case *lGlobal:
		l := lv.t.l
		for i := range l.n {
			g.emit(opStgN, lv.off.n+i, src.n+i, 0)
		}
		for i := range l.s {
			g.emit(opStgS, lv.off.s+i, src.s+i, 0)
		}
		for i := range l.r {
			g.emit(opStgR, lv.off.r+i, src.r+i, 0)
		}
	case *lStore:
		l := lv.t.l
		for i := range l.n {
			g.emit(opStoN, ptr.r, lv.off.n+i, src.n+i)
		}
		for i := range l.s {
			g.emit(opStoS, ptr.r, lv.off.s+i, src.s+i)
		}
		for i := range l.r {
			g.emit(opStoR, ptr.r, lv.off.r+i, src.r+i)
		}
	case *lDiscard:
	}
}

// mentions tells, whether x reads the variable v.
func mentions(x expr, v *local) bool {
	found := false
	walk(x, func(e expr) {
		if l, ok := e.(*eLocal); ok && l.v == v {
			found = true
		}
	})
	return found
}

func walk(x expr, f func(expr)) {
	f(x)
	switch x := x.(type) {
	case *eField:
		walk(x.x, f)
	case *eLoad:
		walk(x.p, f)
	case *eUnary:
		walk(x.x, f)
	case *eBinary:
		walk(x.x, f)
		walk(x.y, f)
	case *eAndOr:
		walk(x.x, f)
		walk(x.y, f)
	case *eCall:
		for _, a := range x.args {
			walk(a, f)
		}
	case *eCallValue:
		walk(x.f, f)
		for _, a := range x.args {
			walk(a, f)
		}
	case *eNative:
		for _, a := range x.args {
			walk(a, f)
		}
	case *eHost:
		for _, a := range x.args {
			walk(a, f)
		}
	case *eStruct:
		for _, a := range x.fields {
			walk(a, f)
		}
	case *eNew:
		walk(x.x, f)
	case *eBox:
		walk(x.x, f)
	case *eListOf:
		for _, a := range x.elems {
			walk(a, f)
		}
	case *eListGet:
		walk(x.l, f)
		walk(x.i, f)
	case *eTupleAt:
		walk(x.x, f)
	case *eListAppend:
		walk(x.l, f)
		for _, a := range x.elems {
			walk(a, f)
		}
	}
}

// move copies the slots of a value.
func (g *gen) move(dst, src layout, l layout) {
	if dst == src {
		return
	}
	for i := range l.n {
		g.emit(opMovN, dst.n+i, src.n+i, 0)
	}
	for i := range l.s {
		g.emit(opMovS, dst.s+i, src.s+i, 0)
	}
	for i := range l.r {
		g.emit(opMovR, dst.r+i, src.r+i, 0)
	}
}

// expr evaluates x and returns its slots: those of a variable, or temporaries.
func (g *gen) expr(x expr) layout {
	switch e := x.(type) {
	case *eLocal:
		return e.v.at.add(e.off)
	case *eNative:
		if e.nat.inline != nil {
			break
		}
		base := g.call(e)
		g.top = base.add(x.typ().l).max(g.top)
		return base
	case *eCall, *eHost, *eCallValue:
		base := g.call(e)
		g.top = base.add(x.typ().l).max(g.top)
		return base
	case *eTupleAt:
		v := g.expr(e.x)
		return v.add(e.x.typ().offs[e.i])
	case *eField:
		if _, ok := e.x.(*eLocal); ok {
			return g.expr(e.x).add(e.off)
		}
	}
	dst := g.alloc(x.typ().l)
	g.exprTo(x, dst)
	return dst
}

// exprTo evaluates x into the slots at dst.
func (g *gen) exprTo(x expr, dst layout) {
	saved := g.top
	defer func() { g.top = saved }()
	switch e := x.(type) {
	case *eConst:
		g.constant(e, dst)
	case *eLocal:
		g.move(dst, e.v.at.add(e.off), e.t.l)
	case *eGlobal:
		l := e.t.l
		for i := range l.n {
			g.emit(opLdgN, dst.n+i, e.off.n+i, 0)
		}
		for i := range l.s {
			g.emit(opLdgS, dst.s+i, e.off.s+i, 0)
		}
		for i := range l.r {
			g.emit(opLdgR, dst.r+i, e.off.r+i, 0)
		}
	case *eField:
		v := g.expr(e.x)
		g.move(dst, v.add(e.off), e.t.l)
	case *eTupleAt:
		v := g.expr(e.x)
		g.move(dst, v.add(e.x.typ().offs[e.i]), e.t.l)
	case *eLoad:
		p := g.expr(e.p)
		l := e.t.l
		for i := range l.n {
			g.emit(opLdoN, dst.n+i, p.r, e.off.n+i)
		}
		for i := range l.s {
			g.emit(opLdoS, dst.s+i, p.r, e.off.s+i)
		}
		for i := range l.r {
			g.emit(opLdoR, dst.r+i, p.r, e.off.r+i)
		}
	case *eUnary:
		v := g.expr(e.x)
		switch e.op {
		case opIsNil:
			g.emit(opIsNil, dst.n, v.r, 0)
		default:
			g.emit(e.op, dst.n, v.n, 0)
		}
	case *eBinary:
		g.binary(e, dst)
	case *eAndOr:
		g.exprTo(e.x, dst)
		op := opJmpF
		if !e.and {
			op = opJmpT
		}
		skip := g.emit(op, 0, dst.n, 0)
		g.exprTo(e.y, dst)
		g.patch(skip, g.pc())
	case *eNative:
		if e.nat.inline != nil {
			var args []layout
			for _, a := range e.args {
				args = append(args, g.expr(a))
			}
			e.nat.inline(g, dst, args)
			return
		}
		base := g.call(e)
		g.move(dst, base, x.typ().l)
	case *eCall, *eHost, *eCallValue:
		base := g.call(e)
		g.move(dst, base, x.typ().l)
	case *eListAppend:
		lst := g.expr(e.l)
		for _, el := range e.elems {
			saved := g.top
			g.appendElem(lst.r, e.l.typ().elem, g.expr(el))
			g.top = saved
		}
	case *eStruct:
		if e.t.kind == kShared {
			obj := g.alloc(layout{r: 1})
			g.newObject(obj.r, e.t)
			g.fill(obj.r, e)
			g.move(dst, obj, e.t.l)
			return
		}
		for _, i := range e.order {
			g.exprTo(e.fields[i], dst.add(e.t.fields[i].off))
		}
	case *eNew:
		obj := g.alloc(layout{r: 1})
		g.newObject(obj.r, e.t.elem)
		if st, ok := e.x.(*eStruct); ok {
			g.fill(obj.r, st)
		} else {
			v := g.expr(e.x)
			g.store(&lStore{t: e.t.elem}, obj, v)
		}
		g.move(dst, obj, e.t.l)
	case *eFunc:
		g.emit(opLdkR, dst.r, g.c.rconst(e.fn.closure), 0)
	case *eBox:
		v := g.expr(e.x)
		g.emit(opBox, g.aux(aux{t: e.v, dst: dst, src: v}), 0, 0)
	case *eListOf:
		lst := g.alloc(layout{r: 1})
		g.emit(opLNew, lst.r, 0, 0)
		for _, el := range e.elems {
			g.appendElem(lst.r, e.t.elem, g.expr(el))
		}
		g.move(dst, lst, e.t.l)
	case *eListGet:
		lst, i := g.expr(e.l), g.expr(e.i)
		if bank, ok := e.t.single(); ok {
			switch bank {
			case bankN:
				g.emit(opLGetN, dst.n, lst.r, i.n)
			case bankS:
				g.emit(opLGetS, dst.s, lst.r, i.n)
			case bankR:
				g.emit(opLGetR, dst.r, lst.r, i.n)
			}
			return
		}
		g.emit(opLGet, g.aux(aux{t: e.t, dst: dst, src: layout{n: i.n, r: lst.r}}), 0, 0)
	default:
		panic(fmt.Sprintf("vm: expression %T", x))
	}
}

// fill evaluates the fields of a struct into the object in r[obj].
func (g *gen) fill(obj int32, e *eStruct) {
	for _, i := range e.order {
		saved := g.top
		v := g.expr(e.fields[i])
		g.store(&lStore{t: e.fields[i].typ(), off: e.t.fields[i].off}, layout{r: obj}, v)
		g.top = saved
	}
}

func (g *gen) newObject(dst int32, t *tinfo) {
	g.emit(opNew, dst, g.c.typeIndex(t), 0)
}

func (g *gen) appendElem(lst int32, t *tinfo, v layout) {
	if bank, ok := t.single(); ok {
		switch bank {
		case bankN:
			g.emit(opLAppN, lst, v.n, 0)
		case bankS:
			g.emit(opLAppS, lst, v.s, 0)
		case bankR:
			g.emit(opLAppR, lst, v.r, 0)
		}
		return
	}
	g.emit(opLApp, g.aux(aux{t: t, dst: layout{r: lst}, src: v}), 0, 0)
}

func (g *gen) aux(a aux) int32 {
	g.f.aux = append(g.f.aux, a)
	return int32(len(g.f.aux) - 1)
}

func (g *gen) constant(e *eConst, dst layout) {
	switch {
	case e.t.l.n == 1:
		if v := int64(e.n); v >= math.MinInt32 && v <= math.MaxInt32 {
			g.emit(opLdiN, dst.n, int32(v), 0)
			return
		}
		g.emit(opLdkN, dst.n, g.c.nconst(e.n), 0)
	case e.t.l.s == 1:
		g.emit(opLdkS, dst.s, g.c.sconst(e.s), 0)
	case e.t.l.r == 1:
		if e.r != nil {
			g.emit(opLdkR, dst.r, g.c.rconst(e.r), 0)
			return
		}
		g.emit(opZeroR, dst.r, 0, 0)
	default:
		panic("vm: constant of " + e.t.String())
	}
}

func (g *gen) binary(e *eBinary, dst layout) {
	x := g.expr(e.x)
	if e.imm {
		g.emit(e.op, dst.n, x.n, int32(e.y.(*eConst).n))
		return
	}
	if c, ok := e.y.(*eConst); ok && e.op == opAddI && !e.swap {
		g.emit(opAddIK, dst.n, x.n, int32(c.n))
		return
	}
	y := g.expr(e.y)
	if e.swap {
		x, y = y, x
	}
	switch e.op {
	case opEqV:
		g.emit(opEqV, g.aux(aux{t: e.x.typ(), dst: dst, src: x, off: y}), 0, 0)
	case opEqR, opNeR:
		g.emit(e.op, dst.n, x.r, y.r)
	case opConcat:
		g.emit(opConcat, dst.s, x.s, y.s)
	case opEqS, opNeS, opLtS, opLeS:
		g.emit(e.op, dst.n, x.s, y.s)
	default:
		g.emit(e.op, dst.n, x.n, y.n)
	}
}

// call evaluates the arguments into the slots above the temporaries, calls the function and returns the slots of
// its results, which the callee writes over its arguments.
func (g *gen) call(x expr) layout {
	var args []expr
	var site callSite
	var op op
	var params *tinfo
	var fnReg layout
	switch e := x.(type) {
	case *eCall:
		args, site.fn, op, params = e.args, e.fn, opCall, e.fn.params
	case *eNative:
		args, site.nat, op, params = e.args, e.nat, opCallNative, e.params
	case *eHost:
		args, site.ext, op = e.args, e.ext, opCallHost
		params = tuple(e.ext.params...)
	case *eCallValue:
		// the function value is evaluated before the arguments
		fnReg = g.expr(e.f)
		args, op = e.args, opCallValue
		var ps []*tinfo
		for _, a := range e.args {
			ps = append(ps, a.typ())
		}
		params = tuple(ps...)
	}
	base := g.top
	site.base = base
	g.alloc(params.l.max(x.typ().l))
	for i, a := range args {
		off := layout{}
		if params.kind == kTuple {
			off = params.offs[i]
		}
		g.exprTo(a, base.add(off))
	}
	g.f.calls = append(g.f.calls, site)
	g.emit(op, int32(len(g.f.calls)-1), fnReg.r, 0)
	return base
}
