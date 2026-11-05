// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import "go/types"

// signaturesFit tells, whether a function value of the signature call may hold a function of the signature value:
// in Go, a function value holds only functions, whose underlying signature is identical to its own, and the subset
// has no assertion, which bypasses this. The types of the parameters and results must be identical after a
// substitution of the type parameters of both sides, which makes them equal. The type parameters of the two sides
// are distinct, even if they are the same, because a generic function may run with other type arguments on each
// side. Names of parameters and receivers do not count. If the comparison is not decidable, e.g. for a generic
// signature, signatures fit with the same number of parameters and results, which arity tells.
func signaturesFit(call, value *types.Signature) (fit, arity bool) {
	if call.Params().Len() != value.Params().Len() || call.Results().Len() != value.Results().Len() || call.Variadic() != value.Variadic() {
		return false, false
	}
	u := unifier{subst: map[sideParam]sidedType{}}
	if !u.unify(sidedType{call, 0}, sidedType{value, 1}) {
		return false, false
	}
	return true, u.undecidable
}

// sidedType is a type of one side of a comparison, whose type parameters are variables of that side.
type sidedType struct {
	t    types.Type
	side int
}

type sideParam struct {
	p    *types.TypeParam
	side int
}

type unifier struct {
	subst map[sideParam]sidedType
	// undecidable tells, that a part of the types could not be compared and was assumed to be equal
	undecidable bool
}

// resolve returns the type, which a bound type parameter stands for.
func (u *unifier) resolve(x sidedType) sidedType {
	for {
		x.t = types.Unalias(x.t)
		p, ok := x.t.(*types.TypeParam)
		if !ok {
			return x
		}
		b, ok := u.subst[sideParam{p, x.side}]
		if !ok {
			return x
		}
		x = b
	}
}

func (u *unifier) unify(x, y sidedType) bool {
	x, y = u.resolve(x), u.resolve(y)
	if px, ok := x.t.(*types.TypeParam); ok {
		if py, ok := y.t.(*types.TypeParam); ok && px == py && x.side == y.side {
			return true
		}
		if u.occurs(sideParam{px, x.side}, y) {
			// no instance of the type parameter equals a type, which contains it
			return false
		}
		u.subst[sideParam{px, x.side}] = y
		return true
	}
	if _, ok := y.t.(*types.TypeParam); ok {
		return u.unify(y, x)
	}
	switch xt := x.t.(type) {
	case *types.Basic:
		yt, ok := y.t.(*types.Basic)
		return ok && xt.Kind() == yt.Kind()
	case *types.Named:
		yt, ok := y.t.(*types.Named)
		if !ok || xt.Obj() != yt.Obj() || xt.TypeArgs().Len() != yt.TypeArgs().Len() {
			return false
		}
		for i := range xt.TypeArgs().Len() {
			if !u.unify(sidedType{xt.TypeArgs().At(i), x.side}, sidedType{yt.TypeArgs().At(i), y.side}) {
				return false
			}
		}
		return true
	case *types.Pointer:
		yt, ok := y.t.(*types.Pointer)
		return ok && u.unify(sidedType{xt.Elem(), x.side}, sidedType{yt.Elem(), y.side})
	case *types.Slice:
		yt, ok := y.t.(*types.Slice)
		return ok && u.unify(sidedType{xt.Elem(), x.side}, sidedType{yt.Elem(), y.side})
	case *types.Array:
		yt, ok := y.t.(*types.Array)
		return ok && xt.Len() == yt.Len() && u.unify(sidedType{xt.Elem(), x.side}, sidedType{yt.Elem(), y.side})
	case *types.Map:
		yt, ok := y.t.(*types.Map)
		return ok && u.unify(sidedType{xt.Key(), x.side}, sidedType{yt.Key(), y.side}) && u.unify(sidedType{xt.Elem(), x.side}, sidedType{yt.Elem(), y.side})
	case *types.Chan:
		yt, ok := y.t.(*types.Chan)
		return ok && xt.Dir() == yt.Dir() && u.unify(sidedType{xt.Elem(), x.side}, sidedType{yt.Elem(), y.side})
	case *types.Signature:
		yt, ok := y.t.(*types.Signature)
		if !ok || xt.Variadic() != yt.Variadic() || xt.Params().Len() != yt.Params().Len() || xt.Results().Len() != yt.Results().Len() {
			return false
		}
		if xt.TypeParams().Len() > 0 || yt.TypeParams().Len() > 0 {
			u.undecidable = true
			return true
		}
		return u.unifyTuple(xt.Params(), yt.Params(), x.side, y.side) && u.unifyTuple(xt.Results(), yt.Results(), x.side, y.side)
	case *types.Struct:
		yt, ok := y.t.(*types.Struct)
		if !ok || xt.NumFields() != yt.NumFields() {
			return false
		}
		for i := range xt.NumFields() {
			fx, fy := xt.Field(i), yt.Field(i)
			if fx.Name() != fy.Name() || fx.Embedded() != fy.Embedded() || xt.Tag(i) != yt.Tag(i) || !fx.Exported() && fx.Pkg() != fy.Pkg() {
				return false
			}
			if !u.unify(sidedType{fx.Type(), x.side}, sidedType{fy.Type(), y.side}) {
				return false
			}
		}
		return true
	case *types.Interface:
		yt, ok := y.t.(*types.Interface)
		if !ok {
			return false
		}
		if !hasTypeParam(xt) && !hasTypeParam(yt) {
			return types.Identical(xt, yt)
		}
	}
	u.undecidable = true
	return true
}

func (u *unifier) unifyTuple(x, y *types.Tuple, sx, sy int) bool {
	for i := range x.Len() {
		if !u.unify(sidedType{x.At(i).Type(), sx}, sidedType{y.At(i).Type(), sy}) {
			return false
		}
	}
	return true
}

// occurs tells, whether the type contains the type parameter.
func (u *unifier) occurs(p sideParam, x sidedType) bool {
	x = u.resolve(x)
	switch t := x.t.(type) {
	case *types.TypeParam:
		return sideParam{t, x.side} == p
	case *types.Named:
		for a := range t.TypeArgs().Types() {
			if u.occurs(p, sidedType{a, x.side}) {
				return true
			}
		}
	case *types.Pointer:
		return u.occurs(p, sidedType{t.Elem(), x.side})
	case *types.Slice:
		return u.occurs(p, sidedType{t.Elem(), x.side})
	case *types.Array:
		return u.occurs(p, sidedType{t.Elem(), x.side})
	case *types.Map:
		return u.occurs(p, sidedType{t.Key(), x.side}) || u.occurs(p, sidedType{t.Elem(), x.side})
	case *types.Chan:
		return u.occurs(p, sidedType{t.Elem(), x.side})
	case *types.Signature:
		for _, tup := range []*types.Tuple{t.Params(), t.Results()} {
			for v := range tup.Variables() {
				if u.occurs(p, sidedType{v.Type(), x.side}) {
					return true
				}
			}
		}
	case *types.Struct:
		for f := range t.Fields() {
			if u.occurs(p, sidedType{f.Type(), x.side}) {
				return true
			}
		}
	}
	return false
}

// hasTypeParam tells, whether the type contains a type parameter, without looking into the underlying types of
// named types.
func hasTypeParam(t types.Type) bool {
	switch t := types.Unalias(t).(type) {
	case *types.TypeParam:
		return true
	case *types.Named:
		for a := range t.TypeArgs().Types() {
			if hasTypeParam(a) {
				return true
			}
		}
	case *types.Pointer:
		return hasTypeParam(t.Elem())
	case *types.Slice:
		return hasTypeParam(t.Elem())
	case *types.Array:
		return hasTypeParam(t.Elem())
	case *types.Map:
		return hasTypeParam(t.Key()) || hasTypeParam(t.Elem())
	case *types.Chan:
		return hasTypeParam(t.Elem())
	case *types.Signature:
		for _, tup := range []*types.Tuple{t.Params(), t.Results()} {
			for v := range tup.Variables() {
				if hasTypeParam(v.Type()) {
					return true
				}
			}
		}
	case *types.Struct:
		for f := range t.Fields() {
			if hasTypeParam(f.Type()) {
				return true
			}
		}
	case *types.Interface:
		for m := range t.Methods() {
			if hasTypeParam(m.Type()) {
				return true
			}
		}
		for e := range t.EmbeddedTypes() {
			if hasTypeParam(e) {
				return true
			}
		}
	}
	return false
}
