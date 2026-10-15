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

// funcSrc is the declaration of a function or method.
type funcSrc struct {
	pkg  *Package
	decl *ast.FuncDecl
	// js is the JavaScript expression of a function of a header package, see the package documentation.
	js string
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
				res[fn] = funcSrc{pkg: p, decl: fd, js: jsBinding(fd)}
			}
		}
	}
	return res
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
