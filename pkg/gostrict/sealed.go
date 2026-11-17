// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/constant"
	"go/types"
	"slices"
	"strings"
)

// A sealed interface is a sum type: an interface with an unexported method can only be implemented in its
// package, and the compiler sees the whole package, so it knows all variants. The semantics follow the proposal of
// union types for Go (golang/go#76920): a type switch over the variants must be exhaustive, a value is never nil
// and the boundary represents a value as its variant and its fields. Once Go has union types, they become a second
// notation of the same model.
//
//	type Shape interface{ isShape() }
//	type Circle struct{ R float64 }
//	func (Circle) isShape() {}

// sealedVariants returns the variants of a sealed interface in the order of their declaration, or false, if t is
// not a sealed interface. A type, whose pointer implements the interface, e.g. func (*Node) isTree(), is a
// variant as a pointer.
func sealedVariants(t types.Type) ([]types.Type, bool) {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil || isError(t) {
		return nil, false
	}
	iface, ok := n.Underlying().(*types.Interface)
	if !ok || iface.NumMethods() == 0 {
		return nil, false
	}
	sealed := false
	for m := range iface.Methods() {
		sealed = sealed || !m.Exported()
	}
	if !sealed {
		return nil, false
	}

	scope := n.Obj().Pkg().Scope()
	refs := variantCache.get(n, func() []variantRef {
		var variants []types.Type
		for _, name := range scope.Names() {
			v, ok := variantOf(scope, name, n)
			if !ok {
				continue
			}
			if _, isIface := v.Underlying().(*types.Interface); isIface {
				continue
			}
			switch {
			case types.Implements(v, iface):
				variants = append(variants, v)
			case types.Implements(types.NewPointer(v), iface):
				variants = append(variants, types.NewPointer(v))
			}
		}
		slices.SortFunc(variants, func(a, b types.Type) int { return int(variantObj(a).Pos() - variantObj(b).Pos()) })
		refs := make([]variantRef, 0, len(variants))
		for _, v := range variants {
			_, ptr := v.(*types.Pointer)
			refs = append(refs, variantRef{name: variantObj(v).Name(), ptr: ptr})
		}
		return refs
	})

	variants := make([]types.Type, 0, len(refs))
	for _, r := range refs {
		v, ok := variantOf(scope, r.name, n)
		if !ok {
			continue
		}
		if r.ptr {
			variants = append(variants, types.NewPointer(v))
		} else {
			variants = append(variants, v)
		}
	}
	return variants, true
}

// variantRef is a variant of a sealed interface by the name of its type in the package of the interface and whether
// its pointer is the variant, see variantCache.
type variantRef struct {
	name string
	ptr  bool
}

// variantCache memoizes the variants of a sealed interface by their names, see [typeCache].
var variantCache typeCache[[]variantRef]

// variantOf returns the named type of the name in the scope of the sealed interface n, a generic one with the type
// arguments of n, e.g. Ok[T] of Result[T], or false, if the name is no such type.
func variantOf(scope *types.Scope, name string, n *types.Named) (*types.Named, bool) {
	tn, ok := scope.Lookup(name).(*types.TypeName)
	if !ok || tn.IsAlias() {
		return nil, false
	}
	v, ok := tn.Type().(*types.Named)
	if !ok {
		return nil, false
	}
	if v.TypeParams().Len() > 0 {
		if v.TypeParams().Len() != n.TypeArgs().Len() {
			return nil, false
		}
		inst, err := types.Instantiate(nil, v, typeArgs(n.TypeArgs()), true)
		if err != nil {
			return nil, false
		}
		v = inst.(*types.Named)
	}
	return v, true
}

// variantObj returns the declaration of the type of a variant.
func variantObj(t types.Type) *types.TypeName {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	return types.Unalias(t).(*types.Named).Obj()
}

// variantName returns the name of a variant, e.g. Circle or *Node.
func variantName(t types.Type) string {
	if _, ok := t.(*types.Pointer); ok {
		return "*" + variantObj(t).Name()
	}
	return variantObj(t).Name()
}

// unionComparable tells, whether the interfaces in t are sealed interfaces, whose variants are all comparable, so
// that a comparison never panics, as golang/go#76920 proposes for union types.
func unionComparable(t types.Type, seen map[types.Type]bool) bool {
	if seen[t] {
		return true
	}
	seen[t] = true
	if isOpaque(t) {
		return true
	}
	switch u := t.Underlying().(type) {
	case *types.Interface:
		variants, ok := sealedVariants(t)
		if !ok {
			return false
		}
		for _, v := range variants {
			if !types.Comparable(v) || !unionComparable(v, seen) {
				return false
			}
		}
	case *types.Struct:
		for f := range u.Fields() {
			if !unionComparable(f.Type(), seen) {
				return false
			}
		}
	}
	return true
}

// isSealed tells, whether t is a sealed interface.
func isSealed(t types.Type) bool {
	_, ok := sealedVariants(t)
	return ok
}

// enumConst is a constant of an enum by its name and value, see enumConstants.
type enumConst struct {
	name string
	val  constant.Value
}

// enumCache memoizes the constants of an enum by their names and values, see [typeCache].
var enumCache typeCache[[]enumConst]

// enumConstants returns the constants of an enum: a named type of an integer or string, for which its package
// declares constants of exactly this type, in the order of their declaration.
func enumConstants(t types.Type) ([]enumConst, bool) {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return nil, false
	}
	switch basicKind(n) {
	case types.Int32, types.Uint32, types.Int64, types.Uint64, types.String:
	default:
		return nil, false
	}

	consts := enumCache.get(n, func() []enumConst {
		scope := n.Obj().Pkg().Scope()
		var objs []*types.Const
		for _, name := range scope.Names() {
			if c, ok := scope.Lookup(name).(*types.Const); ok && types.Identical(c.Type(), n) {
				objs = append(objs, c)
			}
		}
		slices.SortFunc(objs, func(a, b *types.Const) int { return int(a.Pos() - b.Pos()) })
		consts := make([]enumConst, 0, len(objs))
		for _, c := range objs {
			consts = append(consts, enumConst{name: c.Name(), val: c.Val()})
		}
		return consts
	})
	return consts, len(consts) > 0
}

// missingConstants returns the names of the constants, whose values the covered values miss.
func missingConstants(consts []enumConst, covered []constant.Value) []string {
	have := make(map[string]bool, len(covered))
	for _, v := range covered {
		have[constKey(v)] = true
	}
	var missing []string
	seen := map[string]bool{}
	for _, c := range consts {
		key := constKey(c.val)
		if have[key] || seen[key] {
			continue
		}
		missing = append(missing, c.name)
		seen[key] = true
	}
	return missing
}

// constKey identifies the value of a constant: two values are equal, if their keys are, also an integer, which a
// float denotes exactly.
func constKey(v constant.Value) string {
	if v.Kind() == constant.Float {
		if i := constant.ToInt(v); i.Kind() == constant.Int {
			v = i
		}
	}
	return v.ExactString()
}

func joinNames(names []string) string {
	return strings.Join(names, ", ")
}
