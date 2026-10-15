// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/token"
	"go/types"
)

// A struct has value semantics, so the translated code copies it, wherever Go copies it. The copy is only
// observable, if the program changes a value in place: by assigning a field, also through a pointer or a chain
// of fields like a.b.c = v, or by assigning the value of a pointer like *p = v. A struct type, whose values the
// whole program never changes, is immutable: its values are shared instead of copied, e.g. a style in a package
// variable, which each cell of a table refers to. So the values of immutable types also keep their identity, when
// they cross the boundary, which an engine uses to convert a shared value once.

// mutation is the result of the analysis of the whole program.
type mutation struct {
	// changed are the struct types, whose values the program changes, by the origin of generic types
	changed map[*types.TypeName]bool
	// all is set, if the program changes a value of a type parameter, which may be any struct
	all bool
}

// analyzeMutation finds the struct types, whose values the packages of the subset change in place.
func analyzeMutation(prog *Program) *mutation {
	m := &mutation{changed: map[*types.TypeName]bool{}}
	for _, p := range prog.Packages {
		if p.Kind != KindUser {
			continue
		}
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.AssignStmt:
					for _, lhs := range n.Lhs {
						// := assigns the variables, which it does not define
						if id, ok := lhs.(*ast.Ident); ok && n.Tok == token.DEFINE && (p.Info.Defs[id] != nil || id.Name == "_") {
							continue
						}
						m.target(p.Info, lhs)
					}
				case *ast.IncDecStmt:
					m.target(p.Info, n.X)
				case *ast.RangeStmt:
					if n.Tok == token.ASSIGN {
						for _, x := range []ast.Expr{n.Key, n.Value} {
							if x != nil {
								m.target(p.Info, x)
							}
						}
					}
				}
				return true
			})
		}
	}
	return m
}

// target marks the struct types along the target of an assignment: each operand of a field or a pointer, whose
// value changes with it, and the value of a pointer, which other pointers see. A variable or field of an immutable
// type is replaced by the new value instead, so assigning it does not change a value.
func (m *mutation) target(info *types.Info, x ast.Expr) {
	_, star := ast.Unparen(x).(*ast.StarExpr)
	for first := true; ; first = false {
		if !first || star {
			m.mark(info.TypeOf(x))
		}
		switch e := ast.Unparen(x).(type) {
		case *ast.SelectorExpr:
			if sel := info.Selections[e]; sel == nil || sel.Kind() != types.FieldVal {
				return
			}
			x = e.X
		case *ast.StarExpr:
			x = e.X
		default:
			return
		}
	}
}

func (m *mutation) mark(t types.Type) {
	if t == nil {
		return
	}
	if p, ok := t.Underlying().(*types.Pointer); ok {
		t = p.Elem()
	}
	if _, ok := t.(*types.TypeParam); ok {
		m.all = true
		return
	}
	if n, ok := types.Unalias(t).(*types.Named); ok {
		if _, ok := n.Underlying().(*types.Struct); ok {
			m.changed[n.Origin().Obj()] = true
		}
	}
}

// immutable tells, whether t is a struct type, whose values the program never changes, so that they need no copy.
func (m *mutation) immutable(t types.Type) bool {
	if m == nil || m.all || !isStruct(t) {
		return false
	}
	n, ok := types.Unalias(t).(*types.Named)
	return ok && !m.changed[n.Origin().Obj()]
}
