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
// not a sealed interface.
func sealedVariants(t types.Type) ([]*types.Named, bool) {
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
	var variants []*types.Named
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok || tn.IsAlias() {
			continue
		}
		v, ok := tn.Type().(*types.Named)
		if !ok || v.TypeParams().Len() > 0 {
			continue
		}
		if _, isIface := v.Underlying().(*types.Interface); isIface {
			continue
		}
		if types.Implements(v, iface) {
			variants = append(variants, v)
		}
	}
	slices.SortFunc(variants, func(a, b *types.Named) int { return int(a.Obj().Pos() - b.Obj().Pos()) })
	return variants, true
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
