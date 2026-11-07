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
	"slices"
	"strings"
)

const jsDirective = "//gostrict:js "

// hostDirective marks a function of a header package, which has no JavaScript binding, because only a host in Go
// implements it, e.g. a function, which creates a view of nago on the server.
const hostDirective = "//gostrict:host"

// constDirective names the parameters of a function of a header package, whose arguments must be constants, e.g.
// //gostrict:const name for func Activity(name string), because the host reads them from the source at compile
// time. Such a function can only be called, because a call of a function value would bypass the rule.
const constDirective = "//gostrict:const"

// callsDirective tells, when a function of a header package calls the functions, which it takes: //gostrict:calls now
// during the call, e.g. a builder, which changes a frame with a function, //gostrict:calls later only after the call
// returned, e.g. the action of a button. The check of recursion follows a call, which calls now, and assumes now
// without the directive, see [HeaderCallbacks].
const callsDirective = "//gostrict:calls"

// zeroDirective marks a function of a header package, whose last result is a bool or an error, which returns the zero
// value of each result before it, if the bool is false or the error is not nil, e.g. text.ParseInt. The fix of GS033
// may then replace x, _ := f() by x := opt.OrZero(f()) or opt.OrZeroErr(f()), which returns the same value. A function
// of a package in the subset needs no directive: the checker reads its returns.
const zeroDirective = "//gostrict:zero"

// CallTime tells, when a function of a header package calls the functions, which it takes, see [HeaderCallbacks].
type CallTime string

const (
	// CallsNow calls them during the call, e.g. ui.TBox.WithFrame or list.SortFunc.
	CallsNow CallTime = "now"
	// CallsLater calls them only after the call returned, e.g. the host calls the action of ui.PrimaryButton, when
	// the user clicks the button.
	CallsLater CallTime = "later"
)

// funcSrc is the declaration of a function or method.
type funcSrc struct {
	pkg  *Package
	decl *ast.FuncDecl
	// js is the JavaScript expression of a function of a header package, see the package documentation.
	js string
	// host is set for a function of a header package, which only the host implements, see hostDirective
	host bool
	// consts are the parameters, which the constDirective names, and constPos is the position of the directive or
	// token.NoPos without one.
	consts   []string
	constPos token.Pos
	// calls is the value of the callsDirective or "" without one, and callsPos the position of the directive.
	calls    CallTime
	callsPos token.Pos
	// zero is set by the zeroDirective, whose position is zeroPos, or token.NoPos without one.
	zero    bool
	zeroPos token.Pos
}

func funcDecls(prog *Program) map[*types.Func]funcSrc {
	res := map[*types.Func]funcSrc{}
	for _, p := range prog.Packages {
		for _, f := range p.Files {
			for _, d := range f.Decls {
				fd, ok := d.(*ast.FuncDecl)
				if !ok {
					continue
				}
				fn, ok := p.Info.Defs[fd.Name].(*types.Func)
				if !ok {
					continue
				}
				consts, pos := constNames(fd)
				calls, callsPos := callsTime(fd)
				zeroPos := directivePos(fd, zeroDirective)
				res[fn] = funcSrc{pkg: p, decl: fd, js: jsBinding(fd), host: hasDirective(fd, hostDirective), consts: consts, constPos: pos, calls: calls, callsPos: callsPos, zero: zeroPos.IsValid(), zeroPos: zeroPos}
			}
		}
	}
	return res
}

// hasDirective tells, whether the documentation of the function has the directive on a line of its own.
func hasDirective(fd *ast.FuncDecl, directive string) bool {
	if fd.Doc == nil {
		return false
	}
	for _, c := range fd.Doc.List {
		if strings.TrimSpace(c.Text) == directive {
			return true
		}
	}
	return false
}

// directivePos returns the position of the directive on a line of its own in the documentation of the function, or
// token.NoPos without one.
func directivePos(fd *ast.FuncDecl, directive string) token.Pos {
	if fd.Doc == nil {
		return token.NoPos
	}
	for _, c := range fd.Doc.List {
		if strings.TrimSpace(c.Text) == directive {
			return c.Pos()
		}
	}
	return token.NoPos
}

// constNames returns the parameters, which the constDirective in the documentation of the function names, e.g. a
// and b for //gostrict:const a, b, and the position of the directive, or token.NoPos without one.
func constNames(fd *ast.FuncDecl) ([]string, token.Pos) {
	if fd.Doc == nil {
		return nil, token.NoPos
	}
	for _, c := range fd.Doc.List {
		list, ok := strings.CutPrefix(c.Text, constDirective)
		if !ok || list != "" && list[0] != ' ' && list[0] != '\t' {
			continue
		}
		// a comment may follow the names
		list, _, _ = strings.Cut(list, "//")
		var names []string
		for _, name := range strings.Split(list, ",") {
			if name = strings.TrimSpace(name); name != "" && !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
		return names, c.Pos()
	}
	return nil, token.NoPos
}

// callsTime returns the value of the callsDirective in the documentation of the function and its position, or ""
// and token.NoPos without one.
func callsTime(fd *ast.FuncDecl) (CallTime, token.Pos) {
	if fd.Doc == nil {
		return "", token.NoPos
	}
	for _, c := range fd.Doc.List {
		rest, ok := strings.CutPrefix(c.Text, callsDirective)
		if !ok || rest != "" && rest[0] != ' ' && rest[0] != '\t' {
			continue
		}
		return CallTime(strings.TrimSpace(rest)), c.Pos()
	}
	return "", token.NoPos
}

// varBinding returns the binding of a variable of a header package in the documentation of its declaration.
func varBinding(decl *ast.GenDecl, vs *ast.ValueSpec) string {
	for _, doc := range []*ast.CommentGroup{vs.Doc, decl.Doc} {
		if doc == nil {
			continue
		}
		for _, c := range doc.List {
			if expr, ok := strings.CutPrefix(c.Text, jsDirective); ok {
				return strings.TrimSpace(expr)
			}
		}
	}
	return ""
}

func jsBinding(fd *ast.FuncDecl) string {
	if fd.Doc == nil {
		return ""
	}
	for _, c := range fd.Doc.List {
		if expr, ok := strings.CutPrefix(c.Text, jsDirective); ok {
			return strings.TrimSpace(expr)
		}
	}
	return ""
}
