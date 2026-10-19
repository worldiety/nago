// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// Machine executes the bytecode of an instance. Its banks are stacks, on which each call has its frame. It is not
// safe for concurrent use.
type Machine struct {
	prog *Program
	host Host
	n    []uint64
	s    []string
	r    []any
	// hi is the end of the frames in use, where a callback from the host begins its frame
	hi       layout
	depth    int
	maxDepth int
	fuel     int
	ctx      context.Context
	// peak is the end of the frames since the outermost call, whose references it releases
	peak layout
	// calls are the fast calls, the first fast of them are in progress, see [Machine.BeginFast]
	calls []*fastCall
	fast  int
}

// vmError ends a call of the instance, e.g. because the context ended.
type vmError struct {
	err error
}

// fuelChunk is the number of loops and calls between two checks of the context.
const fuelChunk = 1 << 14

var errDepth = errors.New("vm: the calls are nested too deeply")

// ensure grows the banks to hold a frame.
func (m *Machine) ensure(n, s, r int) {
	if n > len(m.n) {
		m.n = append(m.n, make([]uint64, max(n-len(m.n), len(m.n)))...)
	}
	if s > len(m.s) {
		m.s = append(m.s, make([]string, max(s-len(m.s), len(m.s)))...)
	}
	if r > len(m.r) {
		m.r = append(m.r, make([]any, max(r-len(m.r), len(m.r)))...)
	}
}

func (m *Machine) burn() {
	m.fuel = fuelChunk
	if m.ctx != nil {
		select {
		case <-m.ctx.Done():
			panic(&vmError{err: m.ctx.Err()})
		default:
		}
	}
}

// run executes f with its frame at the given offsets of the banks, where its arguments are. Its results replace
// them.
func (m *Machine) run(f *Func, bn, bs, br int) {
	fr := f.frame
	m.ensure(bn+int(fr.n), bs+int(fr.s), br+int(fr.r))
	hi := m.hi
	m.hi = layout{int32(bn) + fr.n, int32(bs) + fr.s, int32(br) + fr.r}
	if m.hi.s > m.peak.s || m.hi.r > m.peak.r {
		m.peak = m.peak.max(m.hi)
	}
	m.exec(f, bn, bs, br)
	m.hi = hi
}

// callIface calls the method of the interface value v: it copies the receiver from the box and the arguments
// behind it into the frame of the method at the base of the call site.
func (m *Machine) callIface(cs *callSite, v any, bn, bs, br int) {
	base := layout{int32(bn) + cs.base.n, int32(bs) + cs.base.s, int32(br) + cs.base.r}
	o, ok := v.(*object)
	if !ok {
		// an error of errors.New
		if err, ok := v.(error); ok && cs.method == "Error" {
			m.s[base.s] = err.Error()
			return
		}
		panic(fmt.Sprintf("vm: %T has no method %s", v, cs.method))
	}
	mr := o.t.methods[cs.method]
	if mr == nil {
		panic(fmt.Sprintf("vm: %s has no method %s", o.t.name, cs.method))
	}
	rn, rs, rr := o.n, o.s, o.r
	if mr.deref {
		elem := o.t.elem
		p := o.r[0].(*object)
		if elem.kind == kShared {
			rn, rs, rr = nil, nil, []any{p}
		} else {
			rn, rs, rr = p.n, p.s, p.r
		}
	}
	at := base.add(layout{int32(len(rn)), int32(len(rs)), int32(len(rr))})
	fn := mr.fn
	m.ensure(int(max(base.n+fn.frame.n, at.n+cs.argsL.n)), int(max(base.s+fn.frame.s, at.s+cs.argsL.s)), int(max(base.r+fn.frame.r, at.r+cs.argsL.r)))
	args := layout{int32(bn) + cs.args.n, int32(bs) + cs.args.s, int32(br) + cs.args.r}
	copy(m.n[at.n:at.n+cs.argsL.n], m.n[args.n:])
	copy(m.s[at.s:at.s+cs.argsL.s], m.s[args.s:])
	copy(m.r[at.r:at.r+cs.argsL.r], m.r[args.r:])
	copy(m.n[base.n:], rn)
	copy(m.s[base.s:], rs)
	copy(m.r[base.r:], rr)
	m.enter()
	m.run(fn, int(base.n), int(base.s), int(base.r))
	m.depth--
}

// errorMessage calls the method Error of a value of the program, which is an error.
func (m *Machine) errorMessage(o *object) string {
	base := m.hi
	cs := &callSite{method: "Error", base: layout{}}
	m.ensure(int(base.n+1), int(base.s+1), int(base.r+1))
	m.callIface(cs, o, int(base.n), int(base.s), int(base.r))
	return m.s[base.s]
}

// runClosure runs the function of a closure with the captured variables in its frame.
func (m *Machine) runClosure(cl *Closure, bn, bs, br int) {
	f := cl.fn
	if len(f.caps) == 0 {
		m.run(f, bn, bs, br)
		return
	}
	fr := f.frame
	m.ensure(bn+int(fr.n), bs+int(fr.s), br+int(fr.r))
	var at layout
	for _, c := range f.caps {
		copy(m.n[bn+int(c.at.n):bn+int(c.at.n+c.l.n)], cl.n[at.n:])
		copy(m.s[bs+int(c.at.s):bs+int(c.at.s+c.l.s)], cl.s[at.s:])
		copy(m.r[br+int(c.at.r):br+int(c.at.r+c.l.r)], cl.r[at.r:])
		at = at.add(c.l)
	}
	m.run(f, bn, bs, br)
}

func (m *Machine) exec(f *Func, bn, bs, br int) {
	fr := f.frame
	n := m.n[bn : bn+int(fr.n)]
	s := m.s[bs : bs+int(fr.s)]
	r := m.r[br : br+int(fr.r)]
	code := f.code
	pc := 0
	for {
		in := code[pc]
		pc++
		switch in.op {
		case opNop:
		case opMovN:
			n[in.a] = n[in.b]
		case opMovS:
			s[in.a] = s[in.b]
		case opMovR:
			r[in.a] = r[in.b]
		case opLdiN:
			n[in.a] = uint64(int64(in.b))
		case opLdkN:
			n[in.a] = m.prog.nconst[in.b]
		case opLdkS:
			s[in.a] = m.prog.sconst[in.b]
		case opLdkR:
			r[in.a] = m.prog.rconst[in.b]
		case opZeroR:
			r[in.a] = nil
		case opLdgN:
			n[in.a] = m.prog.gn[in.b]
		case opLdgS:
			s[in.a] = m.prog.gs[in.b]
		case opLdgR:
			r[in.a] = m.prog.gr[in.b]
		case opStgN:
			m.prog.gn[in.a] = n[in.b]
		case opStgS:
			m.prog.gs[in.a] = s[in.b]
		case opStgR:
			m.prog.gr[in.a] = r[in.b]
		case opLdoN:
			n[in.a] = r[in.b].(*object).n[in.c]
		case opLdoS:
			s[in.a] = r[in.b].(*object).s[in.c]
		case opLdoR:
			r[in.a] = r[in.b].(*object).r[in.c]
		case opStoN:
			r[in.a].(*object).n[in.b] = n[in.c]
		case opStoS:
			r[in.a].(*object).s[in.b] = s[in.c]
		case opStoR:
			r[in.a].(*object).r[in.b] = r[in.c]
		case opNew:
			t := m.prog.types[in.b]
			r[in.a] = newObject(t, t.obj)

		case opAddI:
			n[in.a] = fromI32(int32(n[in.b]) + int32(n[in.c]))
		case opAddIK:
			n[in.a] = fromI32(int32(n[in.b]) + in.c)
		case opMulIK:
			n[in.a] = fromI32(int32(n[in.b]) * in.c)
		case opSubI:
			n[in.a] = fromI32(int32(n[in.b]) - int32(n[in.c]))
		case opMulI:
			n[in.a] = fromI32(int32(n[in.b]) * int32(n[in.c]))
		case opDivI:
			n[in.a] = fromI32(int32(n[in.b]) / int32(n[in.c]))
		case opRemI:
			n[in.a] = fromI32(int32(n[in.b]) % int32(n[in.c]))
		case opDivU:
			n[in.a] = uint64(uint32(n[in.b]) / uint32(n[in.c]))
		case opRemU:
			n[in.a] = uint64(uint32(n[in.b]) % uint32(n[in.c]))
		case opAndI:
			n[in.a] = n[in.b] & n[in.c]
		case opOrI:
			n[in.a] = n[in.b] | n[in.c]
		case opXorI:
			n[in.a] = n[in.b] ^ n[in.c]
		case opAndNotI:
			n[in.a] = n[in.b] &^ n[in.c]
		case opShlI:
			n[in.a] = fromI32(int32(n[in.b]) << in.c)
		case opShrI:
			n[in.a] = fromI32(int32(n[in.b]) >> in.c)
		case opShrU:
			n[in.a] = uint64(uint32(n[in.b]) >> in.c)
		case opNegI:
			n[in.a] = fromI32(-int32(n[in.b]))
		case opNotI:
			n[in.a] = fromI32(^int32(n[in.b]))
		case opEqI:
			n[in.a] = fromBool(int32(n[in.b]) == int32(n[in.c]))
		case opNeI:
			n[in.a] = fromBool(int32(n[in.b]) != int32(n[in.c]))
		case opLtI:
			n[in.a] = fromBool(int32(n[in.b]) < int32(n[in.c]))
		case opLeI:
			n[in.a] = fromBool(int32(n[in.b]) <= int32(n[in.c]))
		case opLtU:
			n[in.a] = fromBool(uint32(n[in.b]) < uint32(n[in.c]))
		case opLeU:
			n[in.a] = fromBool(uint32(n[in.b]) <= uint32(n[in.c]))

		case opAddF:
			n[in.a] = math.Float64bits(math.Float64frombits(n[in.b]) + math.Float64frombits(n[in.c]))
		case opSubF:
			n[in.a] = math.Float64bits(math.Float64frombits(n[in.b]) - math.Float64frombits(n[in.c]))
		case opMulF:
			n[in.a] = math.Float64bits(math.Float64frombits(n[in.b]) * math.Float64frombits(n[in.c]))
		case opDivF:
			n[in.a] = math.Float64bits(math.Float64frombits(n[in.b]) / math.Float64frombits(n[in.c]))
		case opNegF:
			n[in.a] = math.Float64bits(-math.Float64frombits(n[in.b]))
		case opEqF:
			n[in.a] = fromBool(math.Float64frombits(n[in.b]) == math.Float64frombits(n[in.c]))
		case opNeF:
			n[in.a] = fromBool(math.Float64frombits(n[in.b]) != math.Float64frombits(n[in.c]))
		case opLtF:
			n[in.a] = fromBool(math.Float64frombits(n[in.b]) < math.Float64frombits(n[in.c]))
		case opLeF:
			n[in.a] = fromBool(math.Float64frombits(n[in.b]) <= math.Float64frombits(n[in.c]))
		case opI2F:
			n[in.a] = math.Float64bits(float64(int32(n[in.b])))
		case opU2F:
			n[in.a] = math.Float64bits(float64(uint32(n[in.b])))

		case opNotB:
			n[in.a] = n[in.b] ^ 1
		case opIsNil:
			n[in.a] = fromBool(r[in.b] == nil)
		case opEqR:
			n[in.a] = fromBool(r[in.b] == r[in.c])
		case opNeR:
			n[in.a] = fromBool(r[in.b] != r[in.c])
		case opEqV:
			x := &f.aux[in.a]
			n[x.dst.n] = fromBool(equal(x.t, n[x.src.n:], s[x.src.s:], r[x.src.r:], n[x.off.n:], s[x.off.s:], r[x.off.r:]))
		case opConcat:
			s[in.a] = s[in.b] + s[in.c]
		case opEqS:
			n[in.a] = fromBool(s[in.b] == s[in.c])
		case opNeS:
			n[in.a] = fromBool(s[in.b] != s[in.c])
		case opLtS:
			n[in.a] = fromBool(s[in.b] < s[in.c])
		case opLeS:
			n[in.a] = fromBool(s[in.b] <= s[in.c])

		case opJmp:
			pc = int(in.a)
		case opLoop:
			pc = int(in.a)
			if m.fuel--; m.fuel <= 0 {
				m.burn()
			}
		case opJmpT:
			if n[in.b] != 0 {
				pc = int(in.a)
			}
		case opJmpF:
			if n[in.b] == 0 {
				pc = int(in.a)
			}
		case opJNltI:
			if !(int32(n[in.b]) < int32(n[in.c])) {
				pc = int(in.a)
			}
		case opJNleI:
			if !(int32(n[in.b]) <= int32(n[in.c])) {
				pc = int(in.a)
			}
		case opJNeqI:
			if int32(n[in.b]) != int32(n[in.c]) {
				pc = int(in.a)
			}
		case opJNneI:
			if int32(n[in.b]) == int32(n[in.c]) {
				pc = int(in.a)
			}
		case opJEqIK:
			if int32(n[in.b]) == in.c {
				pc = int(in.a)
			}
		case opJNeIK:
			if int32(n[in.b]) != in.c {
				pc = int(in.a)
			}
		case opJNltF:
			if !(math.Float64frombits(n[in.b]) < math.Float64frombits(n[in.c])) {
				pc = int(in.a)
			}
		case opJNleF:
			if !(math.Float64frombits(n[in.b]) <= math.Float64frombits(n[in.c])) {
				pc = int(in.a)
			}

		case opCall:
			cs := &f.calls[in.a]
			m.enter()
			m.run(cs.fn, bn+int(cs.base.n), bs+int(cs.base.s), br+int(cs.base.r))
			m.depth--
			n, s, r = m.n[bn:bn+int(fr.n)], m.s[bs:bs+int(fr.s)], m.r[br:br+int(fr.r)]
		case opCallValue:
			cs := &f.calls[in.a]
			cl := r[in.b].(*Closure)
			m.enter()
			m.runClosure(cl, bn+int(cs.base.n), bs+int(cs.base.s), br+int(cs.base.r))
			m.depth--
			n, s, r = m.n[bn:bn+int(fr.n)], m.s[bs:bs+int(fr.s)], m.r[br:br+int(fr.r)]
		case opCallNative:
			cs := &f.calls[in.a]
			cs.nat.fn(m, n[cs.base.n:], s[cs.base.s:], r[cs.base.r:])
			// a native may call the program, e.g. list.SortFunc, which grows the banks
			n, s, r = m.n[bn:bn+int(fr.n)], m.s[bs:bs+int(fr.s)], m.r[br:br+int(fr.r)]
		case opCallHost:
			cs := &f.calls[in.a]
			m.callHost(cs.ext, layout{int32(bn) + cs.base.n, int32(bs) + cs.base.s, int32(br) + cs.base.r})
			n, s, r = m.n[bn:bn+int(fr.n)], m.s[bs:bs+int(fr.s)], m.r[br:br+int(fr.r)]
		case opRet:
			return

		case opLLen:
			n[in.a] = fromI32(r[in.b].(*list).len)
		case opLGetN:
			n[in.a] = r[in.b].(*list).n[n[in.c]]
		case opLGetS:
			s[in.a] = r[in.b].(*list).s[n[in.c]]
		case opLGetR:
			r[in.a] = r[in.b].(*list).r[n[in.c]]
		case opLGet:
			x := &f.aux[in.a]
			r[x.src.r].(*list).get(int(n[x.src.n]), x.t.l, n[x.dst.n:], s[x.dst.s:], r[x.dst.r:])
		case opLAppN:
			l := r[in.a].(*list)
			l.n = append(l.n, n[in.b])
			l.len++
		case opLAppS:
			l := r[in.a].(*list)
			l.s = append(l.s, s[in.b])
			l.len++
		case opLAppR:
			l := r[in.a].(*list)
			l.r = append(l.r, r[in.b])
			l.len++
		case opLApp:
			x := &f.aux[in.a]
			r[x.dst.r].(*list).add(x.t.l, n[x.src.n:], s[x.src.s:], r[x.src.r:])
		case opLNew:
			r[in.a] = &list{}
		case opCallIface:
			m.callIface(&f.calls[in.a], r[in.b], bn, bs, br)
			n, s, r = m.n[bn:bn+int(fr.n)], m.s[bs:bs+int(fr.s)], m.r[br:br+int(fr.r)]
		case opTypeIs:
			o, ok := r[in.b].(*object)
			n[in.a] = fromBool(ok && o.t == m.prog.types[in.c])
		case opView:
			x := &f.aux[in.a]
			o := r[x.src.r].(*object)
			l := x.t.l
			r[x.dst.r] = &object{t: x.t, n: o.n[x.off.n : x.off.n+l.n], s: o.s[x.off.s : x.off.s+l.s], r: o.r[x.off.r : x.off.r+l.r]}
		case opClosure:
			x := &f.closures[in.a]
			cl := &Closure{fn: x.fn}
			for i, at := range x.caps {
				l := x.fn.caps[i].l
				cl.n = append(cl.n, n[at.n:at.n+l.n]...)
				cl.s = append(cl.s, s[at.s:at.s+l.s]...)
				cl.r = append(cl.r, r[at.r:at.r+l.r]...)
			}
			r[x.dst] = cl
		case opBox:
			x := &f.aux[in.a]
			o := newObject(x.t, x.t.l)
			copy(o.n, n[x.src.n:])
			copy(o.s, s[x.src.s:])
			copy(o.r, r[x.src.r:])
			r[x.dst.r] = o
		default:
			panic(fmt.Sprintf("vm: unknown instruction %d", in.op))
		}
	}
}

// equal compares two values of type t like ==.
func equal(t *tinfo, an []uint64, as []string, ar []any, bn []uint64, bs []string, br []any) bool {
	switch t.kind {
	case kBool, kI32, kU32:
		return int32(an[0]) == int32(bn[0])
	case kF64:
		return math.Float64frombits(an[0]) == math.Float64frombits(bn[0])
	case kStr:
		return as[0] == bs[0]
	case kPtr, kHandle, kFunc:
		return ar[0] == br[0]
	case kIface, kError:
		ao, aok := ar[0].(*object)
		bo, bok := br[0].(*object)
		if !aok || !bok {
			return ar[0] == br[0]
		}
		return ao.t == bo.t && equal(ao.t, ao.n, ao.s, ao.r, bo.n, bo.s, bo.r)
	case kStruct:
		for _, f := range t.fields {
			o := f.off
			if !equal(f.t, an[o.n:], as[o.s:], ar[o.r:], bn[o.n:], bs[o.s:], br[o.r:]) {
				return false
			}
		}
		return true
	case kShared:
		ao, bo := ar[0].(*object), br[0].(*object)
		if ao == bo {
			return true
		}
		for _, f := range t.fields {
			o := f.off
			if !equal(f.t, ao.n[o.n:], ao.s[o.s:], ao.r[o.r:], bo.n[o.n:], bo.s[o.s:], bo.r[o.r:]) {
				return false
			}
		}
		return true
	}
	panic(fmt.Sprintf("vm: comparison of %s", t.name))
}

// callFunc calls fn above the frames in use with the arguments, which args writes into the windows of its
// frame, e.g. the comparison of list.SortFunc, and returns the base of its results.
func (m *Machine) callFunc(cl *Closure, args func(n []uint64, s []string, r []any)) layout {
	base := m.hi
	fn := cl.fn
	m.ensure(int(base.n+fn.frame.n), int(base.s+fn.frame.s), int(base.r+fn.frame.r))
	args(m.n[base.n:], m.s[base.s:], m.r[base.r:])
	m.enter()
	m.runClosure(cl, int(base.n), int(base.s), int(base.r))
	m.depth--
	return base
}

func (m *Machine) enter() {
	m.depth++
	if m.depth > m.maxDepth {
		panic(&vmError{err: errDepth})
	}
	if m.fuel--; m.fuel <= 0 {
		m.burn()
	}
}
