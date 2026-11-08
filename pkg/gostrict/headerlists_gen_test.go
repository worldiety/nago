// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict_test

import (
	"fmt"
	"go/types"
	"sort"
	"strconv"
	"strings"
)

// The generator of the observation of R-SCRIPT-GS-HEADER-LISTS: for each enumerated function a block of a program,
// which calls it with values of its types, holds every list and dict, which it hands over or gets back, and records,
// whether a function of the program runs during the call. hlGen writes the values from the types of go/types; a
// handle comes from a function of its package, which returns it (the one with the fewest parameters), a ui.Window
// is the window of the program.

// hlGen writes the blocks of one program.
type hlGen struct {
	pkgs  map[string]*types.Package // by path
	ctors map[string]*types.Func    // the constructor of a type by its full name
	imps  map[string]bool           // the imports of the program
	n     int                       // a counter for fresh names
	// fixed are expressions for types by their full name, which the generator does not write itself
	fixed map[string]string
}

func (g *hlGen) name(prefix string) string {
	g.n++
	return prefix + strconv.Itoa(g.n)
}

// qual writes a type with the names of the packages, which the program imports.
func (g *hlGen) qual(p *types.Package) string {
	g.imps[p.Path()] = true
	return p.Name()
}

func (g *hlGen) typ(t types.Type) string {
	return types.TypeString(t, g.qual)
}

// hlFull is the full name of a named type without type arguments.
func hlFull(t types.Type) string {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return ""
	}
	return n.Obj().Pkg().Path() + "." + n.Obj().Name()
}

// hlPre are the statements before an expression, hlExpr the expression.
type hlVal struct {
	pre  []string
	expr string
	// held are the names of lists and dicts, which the program holds, with their length
	held []hlHeld
	// sig is the signature of the method of a receiver with type arguments
	sig *types.Signature
}

type hlHeld struct {
	name string
	n    int
}

// value writes a value of t. depth bounds the constructors of handles.
func (g *hlGen) value(t types.Type, depth int) (hlVal, error) {
	if depth > 6 {
		return hlVal{}, fmt.Errorf("no value of %s within 6 constructors", t)
	}
	if e, ok := g.fixed[hlFull(t)]; ok {
		return hlVal{expr: e}, nil
	}
	if types.Identical(t, types.Universe.Lookup("error").Type()) {
		g.imps["errors"] = true
		return hlVal{expr: `errors.New("x")`}, nil
	}
	switch tt := types.Unalias(t).(type) {
	case *types.TypeParam:
		return g.value(types.Typ[types.Int32], depth)
	case *types.Basic:
		switch {
		case tt.Info()&types.IsBoolean != 0:
			return hlVal{expr: "true"}, nil
		case tt.Info()&types.IsString != 0:
			return hlVal{expr: `"x"`}, nil
		case tt.Kind() == types.Float64:
			return hlVal{expr: "1.0"}, nil
		case tt.Kind() == types.Int32:
			return hlVal{expr: "int32(1)"}, nil
		case tt.Kind() == types.Uint32:
			return hlVal{expr: "uint32(1)"}, nil
		}
		return hlVal{}, fmt.Errorf("basic type %s", tt)
	case *types.Slice:
		// the parameter of a variadic function: one element
		return g.value(tt.Elem(), depth)
	case *types.Signature:
		return g.recorder(tt, "own")
	case *types.Named:
		full := hlFull(tt)
		switch full {
		case "go.wdy.de/nago/pkg/gostrict/std/list.List":
			v, err := g.value(tt.TypeArgs().At(0), depth+1)
			if err != nil {
				return hlVal{}, err
			}
			n := g.name("l")
			v.pre = append(v.pre, n+" := list.Of["+g.typ(tt.TypeArgs().At(0))+"]("+v.expr+", "+v.expr+")")
			g.imps["go.wdy.de/nago/pkg/gostrict/std/list"] = true
			v.held = append(v.held, hlHeld{n, 2})
			v.expr = n
			return v, nil
		case "go.wdy.de/nago/pkg/gostrict/std/list.Frozen":
			v, err := g.value(tt.TypeArgs().At(0), depth+1)
			if err != nil {
				return hlVal{}, err
			}
			n := g.name("f")
			v.pre = append(v.pre, n+" := list.Freeze(list.Of["+g.typ(tt.TypeArgs().At(0))+"]("+v.expr+", "+v.expr+"))")
			g.imps["go.wdy.de/nago/pkg/gostrict/std/list"] = true
			v.held = append(v.held, hlHeld{n, 2})
			v.expr = n
			return v, nil
		case "go.wdy.de/nago/pkg/gostrict/std/dict.Map":
			k, err := g.value(tt.TypeArgs().At(0), depth+1)
			if err != nil {
				return hlVal{}, err
			}
			e, err := g.value(tt.TypeArgs().At(1), depth+1)
			if err != nil {
				return hlVal{}, err
			}
			n := g.name("d")
			pre := append(append(k.pre, e.pre...), n+" := dict.New["+g.typ(tt.TypeArgs().At(0))+", "+g.typ(tt.TypeArgs().At(1))+"]()", n+".Set("+k.expr+", "+e.expr+")")
			g.imps["go.wdy.de/nago/pkg/gostrict/std/dict"] = true
			return hlVal{pre: pre, expr: n, held: append(append(k.held, e.held...), hlHeld{n, 1})}, nil
		case "go.wdy.de/nago/pkg/gostrict/std/opt.Option":
			v, err := g.value(tt.TypeArgs().At(0), depth+1)
			if err != nil {
				return hlVal{}, err
			}
			g.imps["go.wdy.de/nago/pkg/gostrict/std/opt"] = true
			v.expr = "opt.Some[" + g.typ(tt.TypeArgs().At(0)) + "](" + v.expr + ")"
			return v, nil
		case "go.wdy.de/nago/pkg/gostrict/std/big.Int":
			g.imps["go.wdy.de/nago/pkg/gostrict/std/big"] = true
			return hlVal{expr: "big.FromInt32(1)"}, nil
		}
		if _, ok := tt.Underlying().(*types.Basic); ok {
			v, err := g.value(tt.Underlying(), depth)
			if err != nil {
				return hlVal{}, err
			}
			if v.expr == "true" {
				return hlVal{expr: g.typ(tt) + "(true)"}, nil
			}
			if v.expr == `"x"` {
				return hlVal{expr: g.typ(tt) + `("")`}, nil
			}
			return hlVal{expr: g.typ(tt) + "(0)"}, nil
		}
		if c, ok := g.ctors[full]; ok {
			return g.call(c, nil, depth+1, "")
		}
		if s, ok := tt.Underlying().(*types.Struct); ok && !hlIsHandle(tt) {
			// a record: each exported field, that needs a value (a list, a dict, a handle)
			var v hlVal
			var fields []string
			for i := range s.NumFields() {
				f := s.Field(i)
				if !f.Exported() {
					continue
				}
				if !hlReaches(f.Type(), map[types.Type]bool{}) && !hlIsHandle(f.Type()) {
					if _, isIface := f.Type().Underlying().(*types.Interface); !isIface {
						if _, isSig := f.Type().Underlying().(*types.Signature); !isSig {
							continue
						}
					}
				}
				fv, err := g.value(f.Type(), depth+1)
				if err != nil {
					return hlVal{}, err
				}
				v.pre, v.held = append(v.pre, fv.pre...), append(v.held, fv.held...)
				fields = append(fields, f.Name()+": "+fv.expr)
			}
			v.expr = g.typ(tt) + "{" + strings.Join(fields, ", ") + "}"
			return v, nil
		}
		if _, ok := tt.Underlying().(*types.Interface); ok {
			if e, ok := g.fixed[full]; ok {
				return hlVal{expr: e}, nil
			}
		}
		return hlVal{}, fmt.Errorf("no constructor of %s", full)
	case *types.Interface:
		if types.Identical(t, types.Universe.Lookup("error").Type()) {
			g.imps["errors"] = true
			return hlVal{expr: `errors.New("x")`}, nil
		}
		return hlVal{}, fmt.Errorf("interface %s", tt)
	case *types.Pointer:
		return hlVal{}, fmt.Errorf("pointer %s", tt)
	}
	return hlVal{}, fmt.Errorf("type %s", t)
}

// zero writes the zero value of t for the result of a recorder.
func (g *hlGen) zero(t types.Type) (hlVal, error) {
	switch tt := types.Unalias(t).Underlying().(type) {
	case *types.Basic:
		if hlFull(t) == "" {
			switch {
			case tt.Info()&types.IsBoolean != 0:
				return hlVal{expr: "false"}, nil
			case tt.Info()&types.IsString != 0:
				return hlVal{expr: `""`}, nil
			case tt.Kind() == types.Float64:
				return hlVal{expr: "0.0"}, nil
			default:
				return hlVal{expr: g.typ(t) + "(0)"}, nil
			}
		}
	case *types.Interface:
		if types.Identical(t, types.Universe.Lookup("error").Type()) {
			return hlVal{expr: "nil"}, nil
		}
	}
	return g.value(t, 1)
}

// recorder writes a function of the type sig, which counts its calls in the variable counter.
func (g *hlGen) recorder(sig *types.Signature, counter string) (hlVal, error) {
	var params []string
	for i := range sig.Params().Len() {
		params = append(params, "_ "+g.typ(sig.Params().At(i).Type()))
	}
	var res []string
	var pre []string
	var held []hlHeld
	for i := range sig.Results().Len() {
		z, err := g.zero(sig.Results().At(i).Type())
		if err != nil {
			return hlVal{}, err
		}
		pre, held = append(pre, z.pre...), append(held, z.held...)
		res = append(res, z.expr)
	}
	body := counter + "++"
	if len(res) > 0 {
		body += "\n\t\treturn " + strings.Join(res, ", ")
	}
	// the values of the results are built outside of the recorder, so that its call creates nothing
	n := g.name("r")
	pre = append(pre, n+" := func("+strings.Join(params, ", ")+") "+hlResults(g, sig)+" {\n\t\t"+body+"\n\t}")
	return hlVal{pre: pre, expr: n, held: held}, nil
}

func hlResults(g *hlGen, sig *types.Signature) string {
	switch sig.Results().Len() {
	case 0:
		return ""
	case 1:
		return g.typ(sig.Results().At(0).Type())
	}
	var r []string
	for i := range sig.Results().Len() {
		r = append(r, g.typ(sig.Results().At(i).Type()))
	}
	return "(" + strings.Join(r, ", ") + ")"
}

// targ finds a type argument for the type parameter p: int32, string or a named type of the packages, which
// satisfies its constraint and has a value.
func (g *hlGen) targ(p *types.TypeParam) (types.Type, error) {
	for _, c := range []types.Type{types.Typ[types.Int32], types.Typ[types.String]} {
		if types.Satisfies(c, p.Constraint().Underlying().(*types.Interface)) {
			return c, nil
		}
	}
	var names []string
	for n := range g.ctors {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		rt := g.ctors[n].Type().(*types.Signature).Results().At(0).Type()
		if types.Satisfies(rt, p.Constraint().Underlying().(*types.Interface)) {
			return rt, nil
		}
	}
	return nil, fmt.Errorf("no type argument for %s %s", p, p.Constraint())
}

// instance returns the signature of f with type arguments and their names for the call.
func (g *hlGen) instance(f *types.Func) (*types.Signature, string, error) {
	sig := f.Type().(*types.Signature)
	if sig.TypeParams().Len() == 0 {
		return sig, "", nil
	}
	var targs []types.Type
	var names []string
	for i := range sig.TypeParams().Len() {
		a, err := g.targ(sig.TypeParams().At(i))
		if err != nil {
			return nil, "", err
		}
		targs = append(targs, a)
		names = append(names, g.typ(a))
	}
	inst, err := types.Instantiate(nil, sig, targs, true)
	if err != nil {
		return nil, "", err
	}
	return inst.(*types.Signature), "[" + strings.Join(names, ", ") + "]", nil
}

// call writes a call of f with values of its parameters; recv is the receiver, if f is a method. A function
// parameter is a recorder, which counts into counter ("" = a fresh counter, which nobody reads).
func (g *hlGen) call(f *types.Func, recv *hlVal, depth int, counter string) (hlVal, error) {
	sig, targs, err := g.instance(f)
	if err != nil {
		return hlVal{}, err
	}
	if recv != nil && recv.sig != nil {
		sig = recv.sig
	}
	var v hlVal
	var args []string
	for i := range sig.Params().Len() {
		pt := sig.Params().At(i).Type()
		if fs, ok := types.Unalias(pt).Underlying().(*types.Signature); ok {
			c := counter
			if c == "" {
				c = g.name("x")
				v.pre = append(v.pre, c+" := int32(0)", "_ = "+c)
			}
			r, err := g.recorder(fs, c)
			if err != nil {
				return hlVal{}, err
			}
			v.pre, v.held = append(v.pre, r.pre...), append(v.held, r.held...)
			args = append(args, r.expr)
			continue
		}
		a, err := g.value(pt, depth)
		if err != nil {
			return hlVal{}, fmt.Errorf("%s: %w", f.FullName(), err)
		}
		v.pre, v.held = append(v.pre, a.pre...), append(v.held, a.held...)
		args = append(args, a.expr)
	}
	fn := g.qual(f.Pkg()) + "." + f.Name() + targs
	if recv != nil {
		v.pre, v.held = append(recv.pre, v.pre...), append(recv.held, v.held...)
		fn = recv.expr + "." + f.Name()
	}
	v.expr = fn + "(" + strings.Join(args, ", ") + ")"
	return v, nil
}

// hlCtors finds for each named type of the packages the exported function, which returns exactly one value of it,
// with the fewest parameters and without a parameter of its own type.
func hlCtors(pkgs []*types.Package) map[string]*types.Func {
	res := map[string]*types.Func{}
	cost := map[string]int{}
	for _, p := range pkgs {
		sc := p.Scope()
		names := sc.Names()
		sort.Strings(names)
		for _, n := range names {
			f, ok := sc.Lookup(n).(*types.Func)
			if !ok || !f.Exported() {
				continue
			}
			sig := f.Type().(*types.Signature)
			if sig.Results().Len() != 1 {
				continue
			}
			full := hlFull(sig.Results().At(0).Type())
			if full == "" {
				continue
			}
			self := false
			for i := range sig.Params().Len() {
				pt := sig.Params().At(i).Type()
				if sl, ok := pt.(*types.Slice); ok {
					pt = sl.Elem()
				}
				if hlFull(pt) == full {
					self = true
				}
			}
			if self {
				continue
			}
			c := sig.Params().Len()
			if sig.Variadic() {
				c--
			}
			if old, ok := cost[full]; !ok || c < old {
				res[full], cost[full] = f, c
			}
		}
	}
	return res
}
