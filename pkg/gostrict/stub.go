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
	"path"
	"regexp"
	"strconv"
)

// A stub of a package keeps its declarations and replaces the bodies of its functions by a panic. The checker
// needs only the declarations of a header package or a package of the standard library, while the bodies may use
// anything of Go. Removing the bodies removes their imports as well, so a stub depends only on the packages,
// which its declarations use.

// stubHeader turns a file of a header package into a stub. It keeps the exported functions, all methods, which
// may seal an interface, all types and constants and the exported variables, which have a binding, and drops the
// unexported functions and variables, which are implementation details.
func stubHeader(f *ast.File) {
	stub(f, func(d ast.Decl) bool {
		switch d := d.(type) {
		case *ast.FuncDecl:
			return d.Name.IsExported() || d.Recv != nil
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				return keepExportedVars(d)
			}
			return d.Tok == token.TYPE || d.Tok == token.CONST || d.Tok == token.IMPORT
		}
		return false
	})
}

// keepExportedVars removes the unexported variables of the declaration and tells, whether any are left.
func keepExportedVars(d *ast.GenDecl) bool {
	var specs []ast.Spec
	for _, s := range d.Specs {
		vs := s.(*ast.ValueSpec)
		exported := false
		for _, n := range vs.Names {
			exported = exported || n.IsExported()
		}
		if exported {
			if vs.Type != nil {
				// the program reads the binding, so the value of Go may use anything
				vs.Values = nil
			}
			specs = append(specs, vs)
		}
	}
	d.Specs = specs
	return len(specs) > 0
}

// stub keeps the declarations, for which keep returns true, replaces the bodies of the functions and removes
// the imports, which the remaining declarations do not use.
func stub(f *ast.File, keep func(ast.Decl) bool) {
	var decls []ast.Decl
	for _, d := range f.Decls {
		if !keep(d) {
			continue
		}
		if fd, ok := d.(*ast.FuncDecl); ok {
			fd.Body = stubBody(fd.Body)
		}
		decls = append(decls, d)
	}
	f.Decls = decls
	removeUnusedImports(f)
}

func stubBody(old *ast.BlockStmt) *ast.BlockStmt {
	pos := token.NoPos
	if old != nil {
		pos = old.Lbrace
	}
	return &ast.BlockStmt{
		Lbrace: pos,
		List: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
			Fun:  ast.NewIdent("panic"),
			Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"gostrict stub"`}},
		}}},
	}
}

var versionSuffix = regexp.MustCompile(`^v[0-9]+$`)

// importName returns the name, under which the file refers to the import.
func importName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	p, _ := strconv.Unquote(spec.Path.Value)
	name := path.Base(p)
	if versionSuffix.MatchString(name) {
		name = path.Base(path.Dir(p))
	}
	return name
}

func removeUnusedImports(f *ast.File) {
	used := map[string]bool{}
	for _, d := range f.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
			continue
		}
		ast.Inspect(d, func(n ast.Node) bool {
			if se, ok := n.(*ast.SelectorExpr); ok {
				if id, ok := se.X.(*ast.Ident); ok {
					used[id.Name] = true
				}
			}
			return true
		})
	}

	var decls []ast.Decl
	var imports []*ast.ImportSpec
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.IMPORT {
			decls = append(decls, d)
			continue
		}
		var specs []ast.Spec
		for _, s := range g.Specs {
			spec := s.(*ast.ImportSpec)
			if used[importName(spec)] {
				specs = append(specs, spec)
				imports = append(imports, spec)
			}
		}
		if len(specs) > 0 {
			g.Specs = specs
			decls = append(decls, g)
		}
	}
	f.Decls = decls
	f.Imports = imports
}
