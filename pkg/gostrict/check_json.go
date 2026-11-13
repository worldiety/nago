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
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"
)

// jsonFix is the fix of a type, which json.Decode or json.Encode can not read or write.
const jsonFix = "use bool, string, int32, uint32, int64, uint64, float64, big.Int, decimal.Dec, bin.Bytes, list.List, list.Frozen, dict.Map with string keys, opt.Option and structs of them"

// jsonDemand tells, which call of json reads or writes a type parameter, see jsonDemands.
type jsonDemand struct {
	// call is the call of json, e.g. json.Decode[T] at item.go:12
	call string
	// verb is read or write
	verb string
}

// jsonDemands are the type parameters, which a call of json.Decode or json.Encode reads or writes. Each engine
// translates a generic function or type for each of its instantiations, so the type is known there: checkJSONInstances
// checks the type arguments of every instantiation of a type parameter, which a call of json reads or writes, also
// through other generic functions. A type parameter of the receiver of a method stands for the type parameter of its
// type, so that every instantiation of the type is checked.
type jsonDemands struct {
	by map[*types.TypeParam]jsonDemand
}

func newJSONDemands() *jsonDemands {
	return &jsonDemands{by: map[*types.TypeParam]jsonDemand{}}
}

// checkJSON rejects a type of json.Decode or json.Encode, which it can not read or write, because each engine reads
// and writes the type, which the program names, without a value in between. A type parameter in the type is checked
// at the instantiations, see jsonDemands.
func (c *checker) checkJSON(n *ast.CallExpr) {
	fun := ast.Unparen(n.Fun)
	if ix, ok := fun.(*ast.IndexExpr); ok {
		fun = ix.X
	}
	var id *ast.Ident
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		id = f.Sel
	case *ast.Ident:
		id = f
	default:
		return
	}
	fn, ok := c.info.Uses[id].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Pkg().Path() != jsonPath || (fn.Name() != "Decode" && fn.Name() != "Encode") {
		return
	}
	inst, ok := c.info.Instances[id]
	if !ok || inst.TypeArgs.Len() != 1 {
		return
	}
	verb := "read"
	if fn.Name() == "Encode" {
		verb = "write"
	}

	t := inst.TypeArgs.At(0)
	params := map[*types.TypeParam]bool{}
	if why := jsonProblem(t, map[types.Type]bool{}, params); why != "" {
		c.report(n, CodeJSON, "json."+fn.Name()+" can not "+verb+" "+types.TypeString(t, types.RelativeTo(c.pkg.Types)), why, jsonFix+", and name the type at the call or as the type argument of a generic function")
		return
	}

	pos := c.prog.Fset.Position(n.Pos())
	short := func(p *types.Package) string {
		if p == c.pkg.Types {
			return ""
		}
		return p.Name()
	}
	call := fmt.Sprintf("json.%s[%s] at %s:%d", fn.Name(), types.TypeString(t, short), filepath.Base(pos.Filename), pos.Line)
	for p := range params {
		if _, ok := c.json.by[p]; !ok {
			c.json.by[p] = jsonDemand{call: call, verb: verb}
		}
	}
}

// jsonProblem returns, why json.Decode can not read t, or "". It collects the type parameters of t into params,
// which an instantiation decides.
func jsonProblem(t types.Type, seen map[types.Type]bool, params map[*types.TypeParam]bool) string {
	if p, ok := types.Unalias(t).(*types.TypeParam); ok {
		params[p] = true
		return ""
	}
	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() != nil {
		targs := n.TypeArgs()
		switch n.Obj().Pkg().Path() + "." + n.Obj().Name() {
		case listPath + ".List", listPath + ".Frozen", optPath + ".Option":
			return jsonProblem(targs.At(0), seen, params)
		case dictPath + ".Map":
			if basicKind(targs.At(0)) != types.String {
				return "a key of an object is a string, but " + t.String() + " has other keys"
			}
			return jsonProblem(targs.At(1), seen, params)
		case binPath + ".Bytes", bigPath + ".Int", decimalPath + ".Dec":
			return ""
		}
	}
	if isStruct(t) {
		if seen[t] {
			return ""
		}
		seen[t] = true
		st := t.Underlying().(*types.Struct)
		for i := range st.NumFields() {
			if why := jsonProblem(st.Field(i).Type(), seen, params); why != "" {
				return why
			}
		}
		return ""
	}
	switch basicKind(t) {
	case types.Bool, types.Int32, types.Uint32, types.Int64, types.Uint64, types.Float64, types.String:
		return ""
	}
	return t.String() + " has no form in JSON"
}

// jsonInstance is the instantiation of a generic function or type at id in pkg.
type jsonInstance struct {
	pkg    *Package
	id     *ast.Ident
	name   string
	params *types.TypeParamList
	args   *types.TypeList
}

// checkJSONInstances checks every instantiation of a type parameter, which json reads or writes, see jsonDemands. A type
// argument, which is itself a type parameter of a generic function or type, passes the demand on, until nothing
// changes.
func checkJSONInstances(prog *Program, d *diagnoser, demands *jsonDemands) {
	if len(demands.by) == 0 {
		return
	}

	// the type parameters of the receivers of the methods of generic types stand for the ones of their types
	canon := map[*types.TypeParam]*types.TypeParam{}
	var instances []jsonInstance
	for _, p := range prog.Packages {
		if p.Kind != KindUser {
			continue
		}
		for _, name := range p.Types.Scope().Names() {
			tn, ok := p.Types.Scope().Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok || named.TypeParams().Len() == 0 {
				continue
			}
			for i := range named.NumMethods() {
				rtp := named.Method(i).Signature().RecvTypeParams()
				for j := 0; j < rtp.Len() && j < named.TypeParams().Len(); j++ {
					canon[rtp.At(j)] = named.TypeParams().At(j)
				}
			}
		}

		for id, inst := range p.Info.Instances {
			var params *types.TypeParamList
			switch obj := p.Info.Uses[id].(type) {
			case *types.Func:
				params = obj.Origin().Signature().TypeParams()
			case *types.TypeName:
				if named, ok := obj.Type().(*types.Named); ok {
					params = named.Origin().TypeParams()
				}
			}
			if params == nil || params.Len() != inst.TypeArgs.Len() {
				continue
			}
			instances = append(instances, jsonInstance{pkg: p, id: id, name: id.Name, params: params, args: inst.TypeArgs})
		}
	}

	demand := func(p *types.TypeParam) (jsonDemand, bool) {
		if c, ok := canon[p]; ok {
			p = c
		}
		dm, ok := demands.by[p]
		return dm, ok
	}
	for p, dm := range demands.by {
		if c, ok := canon[p]; ok {
			if _, ok := demands.by[c]; !ok {
				demands.by[c] = dm
			}
		}
	}

	// the order of the instances is the one of the source, so that the reports do not depend on the order of a map
	slices.SortFunc(instances, func(a, b jsonInstance) int {
		pa, pb := prog.Fset.Position(a.id.Pos()), prog.Fset.Position(b.id.Pos())
		if c := strings.Compare(pa.Filename, pb.Filename); c != 0 {
			return c
		}
		return pa.Offset - pb.Offset
	})

	reported := map[token.Pos]bool{}
	for changed := true; changed; {
		changed = false
		for _, in := range instances {
			for i := range in.params.Len() {
				dm, ok := demand(in.params.At(i))
				if !ok {
					continue
				}
				arg := in.args.At(i)
				params := map[*types.TypeParam]bool{}
				if why := jsonProblem(arg, map[types.Type]bool{}, params); why != "" {
					if !reported[in.id.Pos()] {
						reported[in.id.Pos()] = true
						typ := types.TypeString(arg, types.RelativeTo(in.pkg.Types))
						d.report(in.id.Pos(), in.id.End(), CodeJSON,
							fmt.Sprintf("%s can not be instantiated with %s, because json can not %s it", in.name, typ, dm.verb),
							fmt.Sprintf("the type parameter %s reaches %s: %s", in.params.At(i).Obj().Name(), dm.call, why),
							jsonFix+" as the type argument")
					}
					continue
				}
				for p := range params {
					if c, ok := canon[p]; ok {
						p = c
					}
					if _, ok := demands.by[p]; !ok {
						demands.by[p] = dm
						changed = true
					}
				}
			}
		}
	}
}
