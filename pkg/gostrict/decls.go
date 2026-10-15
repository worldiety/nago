// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"go/ast"
	"go/types"
	"strings"
)

const jsDirective = "//gostrict:js "

// hostDirective marks a function of a header package, which has no JavaScript binding, because only a host in Go
// implements it, e.g. a function, which creates a view of nago on the server.
const hostDirective = "//gostrict:host"

// funcSrc is the declaration of a function or method.
type funcSrc struct {
	pkg  *Package
	decl *ast.FuncDecl
	// js is the JavaScript expression of a function of a header package, see the package documentation.
	js string
	// host is set for a function of a header package, which only the host implements, see hostDirective
	host bool
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
				res[fn] = funcSrc{pkg: p, decl: fd, js: jsBinding(fd), host: hasDirective(fd, hostDirective)}
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
