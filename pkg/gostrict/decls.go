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
				res[fn] = funcSrc{pkg: p, decl: fd, js: jsBinding(fd), host: hasDirective(fd, hostDirective), consts: consts, constPos: pos}
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
