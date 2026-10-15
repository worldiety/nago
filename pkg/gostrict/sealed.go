// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/constant"
	gotoken "go/token"
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
	var variants []types.Type
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok || tn.IsAlias() {
			continue
		}
		v, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		if v.TypeParams().Len() > 0 {
			// a generic variant of a generic interface has its type arguments, e.g. Ok[T] of Result[T]
			if v.TypeParams().Len() != n.TypeArgs().Len() {
				continue
			}
			inst, err := types.Instantiate(nil, v, typeArgs(n.TypeArgs()), true)
			if err != nil {
				continue
			}
			v = inst.(*types.Named)
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
	return variants, true
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

// enumConstants returns the constants of an enum: a named type of int32, uint32 or string, for which its package
// declares constants of exactly this type, in the order of their declaration.
func enumConstants(t types.Type) ([]*types.Const, bool) {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return nil, false
	}
	switch basicKind(n) {
	case types.Int32, types.Uint32, types.String:
	default:
		return nil, false
	}

	scope := n.Obj().Pkg().Scope()
	var consts []*types.Const
	for _, name := range scope.Names() {
		if c, ok := scope.Lookup(name).(*types.Const); ok && types.Identical(c.Type(), n) {
			consts = append(consts, c)
		}
	}
	slices.SortFunc(consts, func(a, b *types.Const) int { return int(a.Pos() - b.Pos()) })
	return consts, len(consts) > 0
}

// missingConstants returns the names of the constants, whose values the covered values miss.
func missingConstants(consts []*types.Const, covered []constant.Value) []string {
	var missing []string
	seen := map[string]bool{}
	for _, c := range consts {
		key := c.Val().ExactString()
		if seen[key] {
			continue
		}
		found := false
		for _, v := range covered {
			found = found || constant.Compare(v, gotoken.EQL, c.Val())
		}
		if !found {
			missing = append(missing, c.Name())
			seen[key] = true
		}
	}
	return missing
}

func joinNames(names []string) string {
	return strings.Join(names, ", ")
}
