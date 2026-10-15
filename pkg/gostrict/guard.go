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
	"go/token"
	"go/types"
)

// A guard is a condition, which proves a property of a variable, e.g. b != 0 before a / b. The checker accepts
// a guard, if it dominates the use: the use is in the branch of an if, whose condition implies the property, or
// after an if without else, whose body leaves the block and whose condition is the negation. The variable must
// not change between the guard and the use: it is not assigned in between, not in a loop around the use, which
// does not contain the guard, and not in a function literal at all.

type assignment struct {
	pos token.Pos
	// captured tells, that a function literal assigns a variable declared outside of it, which may happen at any
	// time.
	captured bool
}

// collectAssignments returns the assignments of all variables of the package, which follow their definition.
func collectAssignments(info *types.Info, files []*ast.File) map[types.Object][]assignment {
	res := map[types.Object][]assignment{}
	var stack []ast.Node

	mark := func(e ast.Expr) {
		id, ok := ast.Unparen(e).(*ast.Ident)
		if !ok {
			return
		}
		obj := info.Uses[id]
		if obj == nil {
			return
		}
		captured := false
		for _, n := range stack {
			if l, ok := n.(*ast.FuncLit); ok && (obj.Pos() < l.Pos() || obj.Pos() >= l.End()) {
				captured = true
			}
		}
		res[obj] = append(res[obj], assignment{pos: id.Pos(), captured: captured})
	}

	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return false
			}
			stack = append(stack, n)
			switch n := n.(type) {
			case *ast.AssignStmt:
				for _, l := range n.Lhs {
					mark(l)
				}
			case *ast.IncDecStmt:
				mark(n.X)
			case *ast.RangeStmt:
				if n.Tok == token.ASSIGN {
					mark(n.Key)
					if n.Value != nil {
						mark(n.Value)
					}
				}
			}
			return true
		})
	}
	return res
}

// atom tells, whether the condition proves the property, if it evaluates to positive.
type atom func(cond ast.Expr, positive bool) bool

func implies(cond ast.Expr, positive bool, a atom) bool {
	switch e := ast.Unparen(cond).(type) {
	case *ast.UnaryExpr:
		if e.Op == token.NOT {
			return implies(e.X, !positive, a)
		}
	case *ast.BinaryExpr:
		switch e.Op {
		case token.LAND:
			return positive && (implies(e.X, true, a) || implies(e.Y, true, a))
		case token.LOR:
			return !positive && (implies(e.X, false, a) || implies(e.Y, false, a))
		}
	}
	return a(ast.Unparen(cond), positive)
}

// guardFor finds the guard of the node on top of the stack. It returns the index of the guarding node in the
// stack and the position, from which on the variable must not change.
func (c *checker) guardFor(a atom) (depth int, from token.Pos, ok bool) {
	for i := len(c.stack) - 1; i > 0; i-- {
		child, parent := c.stack[i], c.stack[i-1]
		switch p := parent.(type) {
		case *ast.IfStmt:
			if child == ast.Node(p.Body) && implies(p.Cond, true, a) {
				return i - 1, p.Cond.Pos(), true
			}
			if p.Else != nil && child == ast.Node(p.Else) && implies(p.Cond, false, a) {
				return i - 1, p.Cond.Pos(), true
			}
		case *ast.BinaryExpr:
			if child == ast.Node(p.Y) {
				if (p.Op == token.LAND && implies(p.X, true, a)) || (p.Op == token.LOR && implies(p.X, false, a)) {
					return i - 1, p.X.Pos(), true
				}
			}
		case *ast.BlockStmt:
			if pos, ok := earlyExit(p.List, child, a); ok {
				return i - 1, pos, true
			}
		case *ast.CaseClause:
			if pos, ok := earlyExit(p.Body, child, a); ok {
				return i - 1, pos, true
			}
		}
	}
	return 0, token.NoPos, false
}

// earlyExit finds an if before child, which leaves the block unless the property holds.
func earlyExit(list []ast.Stmt, child ast.Node, a atom) (token.Pos, bool) {
	for _, s := range list {
		if ast.Node(s) == child {
			return token.NoPos, false
		}
		is, ok := s.(*ast.IfStmt)
		if !ok || is.Init != nil || is.Else != nil || !terminates(is.Body) {
			continue
		}
		if implies(is.Cond, false, a) {
			return is.Cond.Pos(), true
		}
	}
	return token.NoPos, false
}

func terminates(b *ast.BlockStmt) bool {
	if len(b.List) == 0 {
		return false
	}
	switch s := b.List[len(b.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return s.Tok == token.BREAK || s.Tok == token.CONTINUE
	}
	return false
}

// stable tells, whether obj keeps its value from the position from up to the use on top of the stack.
func (c *checker) stable(obj types.Object, from token.Pos) bool {
	use := c.stack[len(c.stack)-1].Pos()
	for _, a := range c.assigns[obj] {
		if a.captured {
			return false
		}
		if a.pos > from && a.pos < use {
			return false
		}
		// an assignment in a loop around the use, which was entered after the guard, happens before the next use
		for _, n := range c.stack {
			switch n.(type) {
			case *ast.ForStmt, *ast.RangeStmt:
				if !(n.Pos() <= from && from < n.End()) && n.Pos() <= a.pos && a.pos < n.End() {
					return false
				}
			}
		}
	}
	return true
}

func (c *checker) refersTo(e ast.Expr, obj types.Object) bool {
	id, ok := ast.Unparen(e).(*ast.Ident)
	return ok && c.info.Uses[id] == obj
}

func (c *checker) isNilExpr(e ast.Expr) bool {
	tv, ok := c.info.Types[e]
	return ok && tv.IsNil()
}

// nonNil is the property x != nil.
func (c *checker) nonNil(obj types.Object) atom {
	return func(cond ast.Expr, positive bool) bool {
		b, ok := cond.(*ast.BinaryExpr)
		if !ok || (b.Op != token.NEQ && b.Op != token.EQL) {
			return false
		}
		matches := (c.refersTo(b.X, obj) && c.isNilExpr(b.Y)) || (c.refersTo(b.Y, obj) && c.isNilExpr(b.X))
		return matches && (b.Op == token.NEQ) == positive
	}
}

// truthy is the property, that the bool variable is true.
func (c *checker) truthy(obj types.Object) atom {
	return func(cond ast.Expr, positive bool) bool {
		return positive && c.refersTo(cond, obj)
	}
}

// nonZero is the property x != 0, which a comparison with a constant proves, e.g. x > 0 or x != 0.
func (c *checker) nonZero(obj types.Object) atom {
	return func(cond ast.Expr, positive bool) bool {
		b, ok := cond.(*ast.BinaryExpr)
		if !ok {
			return false
		}
		op := b.Op
		var other ast.Expr
		switch {
		case c.refersTo(b.X, obj):
			other = b.Y
		case c.refersTo(b.Y, obj):
			other = b.X
			// mirror the comparison, so that obj is on the left
			op = map[token.Token]token.Token{token.LSS: token.GTR, token.GTR: token.LSS, token.LEQ: token.GEQ, token.GEQ: token.LEQ, token.EQL: token.EQL, token.NEQ: token.NEQ}[op]
		default:
			return false
		}

		tv, ok := c.info.Types[other]
		if !ok || tv.Value == nil {
			return false
		}
		sign := constant.Sign(tv.Value)
		if !positive {
			op = map[token.Token]token.Token{token.LSS: token.GEQ, token.GTR: token.LEQ, token.LEQ: token.GTR, token.GEQ: token.LSS, token.EQL: token.NEQ, token.NEQ: token.EQL}[op]
		}
		switch op {
		case token.NEQ:
			return sign == 0
		case token.GTR:
			return sign >= 0
		case token.LSS:
			return sign <= 0
		case token.GEQ:
			return sign > 0
		case token.LEQ:
			return sign < 0
		}
		return false
	}
}
