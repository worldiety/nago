// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/constant"
	"go/types"
)

// checkTypeSwitch allows a type switch only over a sealed interface, whose variants it must cover, unless it has
// a default case.
func (c *checker) checkTypeSwitch(n *ast.TypeSwitchStmt) {
	var x ast.Expr
	switch a := n.Assign.(type) {
	case *ast.ExprStmt:
		x = a.X.(*ast.TypeAssertExpr).X
	case *ast.AssignStmt:
		x = a.Rhs[0].(*ast.TypeAssertExpr).X
	}
	t := c.info.TypeOf(x)
	variants, ok := sealedVariants(t)
	if !ok {
		c.report(n, CodeTypeSwitch, "type switches are only allowed over sealed interfaces", "only the variants of a sealed interface are known to the compiler, which checks, that a switch covers them", "make the interface sealed with an unexported method, e.g. interface{ isShape() }, and implement it by each variant, or declare a method in the interface")
		return
	}

	covered := map[types.Type]bool{}
	hasDefault := false
	for _, st := range n.Body.List {
		cc := st.(*ast.CaseClause)
		if cc.List == nil {
			hasDefault = true
			continue
		}
		for _, e := range cc.List {
			ct := types.Unalias(c.info.TypeOf(e))
			idx := -1
			for i, v := range variants {
				if ct != nil && types.Identical(v, ct) {
					idx = i
				}
			}
			if idx < 0 {
				c.report(e, CodeTypeSwitch, types.ExprString(e)+" is not a variant of "+t.String(), "the variants of a sealed interface are the types of its package, whose values implement it", "list only these types: "+variantNames(variants))
				continue
			}
			covered[variants[idx]] = true
		}
	}

	if hasDefault {
		return
	}
	var missing []string
	for _, v := range variants {
		if !covered[v] {
			missing = append(missing, variantName(v))
		}
	}
	if len(missing) > 0 {
		c.report(n, CodeExhaustive, "the switch misses the variants "+joinNames(missing)+" of "+t.String(), "a switch over a sealed interface handles each variant, so that a new variant shows every place, which must handle it", "add a case for each missing variant")
	}
}

func variantNames(variants []types.Type) string {
	var names []string
	for _, v := range variants {
		names = append(names, variantName(v))
	}
	return joinNames(names)
}

// checkEnumSwitch requires a switch over an enum to cover all its constants, unless it has a default case.
func (c *checker) checkEnumSwitch(n *ast.SwitchStmt) {
	if n.Tag == nil {
		return
	}
	consts, ok := enumConstants(c.info.TypeOf(n.Tag))
	if !ok {
		return
	}

	var covered []constant.Value
	for _, st := range n.Body.List {
		cc := st.(*ast.CaseClause)
		if cc.List == nil {
			return
		}
		for _, e := range cc.List {
			tv, ok := c.info.Types[e]
			if !ok || tv.Value == nil {
				// a case, which is not a constant, could cover anything
				return
			}
			covered = append(covered, tv.Value)
		}
	}

	if missing := missingConstants(consts, covered); len(missing) > 0 {
		c.report(n, CodeExhaustive, "the switch misses the constants "+joinNames(missing)+" of "+c.info.TypeOf(n.Tag).String(), "a switch over an enum handles each constant, so that a new constant shows every place, which must handle it", "add a case for each missing constant")
	}
}
