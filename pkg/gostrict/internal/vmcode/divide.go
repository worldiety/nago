// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

// The package variables of a program never change after their initialization, so all instances share them, e.g. a
// table of list.Frozen. A variable, whose initialization calls a function of the host, e.g. three.NewScene, or reads
// such a variable, holds a value of the host, which belongs to the instance, like in a browser or in goja, where
// each instance initializes its variables. So it lives in the slots of each instance.

// divideGlobals divides the package variables between the program and the instances and splits the initialization
// into init, which runs once, and instInit, which runs for each instance.
func (c *compiler) divideGlobals(init *Func) {
	// the functions, which call the host, also through the functions they call or refer to
	hosts := map[*Func]bool{}
	refs := map[*Func][]*Func{}
	reads := map[*Func][]*global{}
	direct := func(x expr, refs *[]*Func, reads *[]*global) bool {
		found := false
		walk(x, func(e expr) {
			switch e := e.(type) {
			case *eHost:
				found = found || !e.ext.variable
			case *eCallIface:
				// the method of the variant is not known here
				found = true
			case *eCall:
				*refs = append(*refs, e.fn)
			case *eFunc:
				*refs = append(*refs, e.fn)
			case *eClosure:
				*refs = append(*refs, e.fn)
			case *eGlobal:
				*reads = append(*reads, e.g)
			}
		})
		return found
	}
	entry := map[*Func]bool{}
	for _, e := range c.out.entries {
		entry[e.hook] = true
	}
	for _, f := range c.out.funcs {
		var r []*Func
		var g []*global
		h := false
		walkStmts(f.body, func(x expr) { h = direct(x, &r, &g) || h })
		hosts[f], refs[f], reads[f] = h, r, g
	}
	// the variables, whose initialization calls the host
	type initStmt struct {
		s    *sAssign
		refs []*Func
		g    []*global
		host bool
	}
	var inits []initStmt
	for _, s := range init.body {
		a := s.(*sAssign)
		var r []*Func
		var g []*global
		h := false
		for _, x := range a.rhs {
			h = direct(x, &r, &g) || h
		}
		inits = append(inits, initStmt{s: a, refs: r, g: g, host: h})
	}
	for changed := true; changed; {
		changed = false
		calls := func(fs []*Func) bool {
			for _, f := range fs {
				if hosts[f] || entry[f] {
					return true
				}
			}
			return false
		}
		readsInst := func(gs []*global) bool {
			for _, g := range gs {
				if g.inst {
					return true
				}
			}
			return false
		}
		for _, f := range c.out.funcs {
			if !hosts[f] && (calls(refs[f]) || readsInst(reads[f])) {
				hosts[f], changed = true, true
			}
		}
		for _, in := range inits {
			if in.host || calls(in.refs) || readsInst(in.g) {
				for _, lv := range in.s.lhs {
					if lg, ok := lv.(*lGlobal); ok && !lg.g.inst {
						lg.g.inst, changed = true, true
					}
				}
			}
		}
	}

	// the slots of the variables and the two initializations
	for _, g := range c.globalList {
		if g.inst {
			g.off = c.out.instGlobals
			c.out.instGlobals = c.out.instGlobals.add(g.t.l)
		} else {
			g.off = c.out.globals
			c.out.globals = c.out.globals.add(g.t.l)
		}
	}
	var shared, own []stmt
	for _, in := range inits {
		inst := false
		for _, lv := range in.s.lhs {
			if lg, ok := lv.(*lGlobal); ok && lg.g.inst {
				inst = true
			}
		}
		if inst {
			own = append(own, in.s)
		} else {
			shared = append(shared, in.s)
		}
	}
	init.body = shared
	if len(own) > 0 {
		c.out.instInit = &Func{name: "$instInit", params: init.params, results: init.results, body: own, locals: init.locals}
	}
}

// walkStmts calls f for each expression of the statements.
func walkStmts(list []stmt, f func(expr)) {
	for _, s := range list {
		switch s := s.(type) {
		case *sAssign:
			for _, x := range s.rhs {
				f(x)
			}
			for _, lv := range s.lhs {
				if st, ok := lv.(*lStore); ok {
					f(st.p)
				}
			}
		case *sExpr:
			f(s.x)
		case *sIf:
			f(s.cond)
			walkStmts(s.then, f)
			walkStmts(s.els, f)
		case *sLoop:
			if s.cond != nil {
				f(s.cond)
			}
			walkStmts(s.body, f)
			walkStmts(s.post, f)
		case *sBlock:
			walkStmts(s.body, f)
		case *sReturn:
			for _, x := range s.vals {
				f(x)
			}
		}
	}
}
