// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/types"
	"strings"
)

// handleField returns the field of type js.Ref of a handle type, see package std/js, or nil.
func handleField(t types.Type) *types.Var {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil {
		return nil
	}
	st, ok := n.Underlying().(*types.Struct)
	if !ok || st.NumFields() != 1 || st.Field(0).Exported() {
		return nil
	}
	f := st.Field(0)
	ft, ok := types.Unalias(f.Type()).(*types.Named)
	if !ok || ft.Obj().Pkg() == nil || ft.Obj().Pkg().Path() != jsPath || ft.Obj().Name() != "Ref" {
		return nil
	}
	return f
}

// isHandle tells, whether t is a handle of a header package, which is an object of the host in JavaScript.
func isHandle(t types.Type) bool {
	return handleField(t) != nil
}

// isBig tells, whether t is the big.Int of package std/big, a handle, which is a BigInt in JavaScript and crosses
// the boundary as a value.
func isBig(t types.Type) bool {
	n, ok := types.Unalias(t).(*types.Named)
	return ok && n.Obj().Pkg() != nil && n.Obj().Pkg().Path() == bigPath && n.Obj().Name() == "Int"
}

// boundaryTerms returns the types of the constraint of a type parameter of a header function, which must be a
// union of types, which cross the boundary, or false.
func boundaryTerms(tp *types.TypeParam) ([]types.Type, bool) {
	iface, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok || iface.NumMethods() > 0 || iface.NumEmbeddeds() != 1 {
		return nil, false
	}

	var terms []types.Type
	switch e := iface.EmbeddedType(0).(type) {
	case *types.Union:
		for i := 0; i < e.Len(); i++ {
			terms = append(terms, e.Term(i).Type())
		}
	default:
		terms = append(terms, e)
	}

	for _, t := range terms {
		k := valueKind(t)
		if k == ValueInvalid || k == ValueError {
			return nil, false
		}
	}
	return terms, true
}

// externKey returns the name of the instance of a function of a header package, e.g. "example.com/three.Add" or
// "example.com/three.Add[example.com/three.Mesh]". A type argument is named by its core type, so that a named
// type like type ID string uses the function of string.
func externKey(fn *types.Func, targs []types.Type) string {
	name := fn.Pkg().Path() + "." + fn.Name()
	if len(targs) == 0 {
		return name
	}
	var parts []string
	for _, t := range targs {
		if !isHandle(t) {
			t = t.Underlying()
		}
		parts = append(parts, typeKey(t))
	}
	return name + "[" + strings.Join(parts, ",") + "]"
}
