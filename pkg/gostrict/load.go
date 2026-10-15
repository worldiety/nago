// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gostrict

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// The sources of the library of the subset. Its packages are header packages, so the loader uses their
// declarations only.
//
//go:embed std
var stdFS embed.FS

// The stubs of the allowed packages of the standard library, see TestStubs.
//
//go:embed all:_stubs
var stubFS embed.FS

// The sources of the libraries of hosts, e.g. lib/three. They are header packages like the library of the
// subset, so that a program, which an application loads from untrusted modules, may use them.
//
//go:embed lib
var libFS embed.FS

const (
	stdModule = "go.wdy.de/nago/pkg/gostrict/std"
	libModule = "go.wdy.de/nago/pkg/gostrict/lib"
)

// Kind tells, how a package is translated.
type Kind int

const (
	// KindUser is a package in the subset, which is checked and translated.
	KindUser Kind = iota
	// KindHeader is a trusted package, whose functions are bound to JavaScript, see the package documentation.
	KindHeader
	// KindStd is an allowed package of the standard library, whose allowed functions are built in.
	KindStd
)

// Package is a loaded and type checked package.
type Package struct {
	Path  string
	Name  string
	Kind  Kind
	Files []*ast.File
	Types *types.Package
	Info  *types.Info
}

// Program is a package with all packages it depends on.
type Program struct {
	Fset *token.FileSet
	Main *Package
	// Packages contains each package after the packages it imports.
	Packages []*Package
	byPath   map[string]*Package
	// Diagnostics contains the import and type errors.
	Diagnostics Diagnostics
	sources     map[string][]byte
	// Terminating requires the terminating profile of all packages in the subset, see the directive
	// //gostrict:terminating.
	Terminating bool
	// importNames maps the name of a file to its import path and name, e.g. example.com/app/main.go, which a
	// source map shows instead of the path on disk
	importNames map[string]string
}

// Package returns the loaded package or nil.
func (p *Program) Package(path string) *Package {
	return p.byPath[path]
}

// Module is a tree with the sources of a Go module, e.g. an [embed.FS], a directory on disk or files, which a
// language model has written.
type Module struct {
	// Path is the module path, e.g. "example.com/app". The packages with this path or below it are read from FS.
	Path string
	FS   fs.FS
	// Dir is put before the paths of the files in diagnostics, e.g. the directory of FS on disk. If empty, the
	// module path is used.
	Dir string
	// Trusted allows header packages in the module, whose bodies are not checked and whose JavaScript bindings
	// run with all rights of the host, e.g. in a browser. A module, which a language model or a user has written,
	// must not be trusted, so its header packages are reported. The libraries std and lib are always trusted.
	Trusted bool
}

const headerDirective = "//gostrict:header"

// Load loads the package in dir and the packages, which it imports, from the module on disk, which contains dir.
// Problems of the source code are returned as [Program.Diagnostics], an error means, that the packages could not
// be read at all.
func Load(dir string) (*Program, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	modDir, modPath, err := findModule(dir)
	if err != nil {
		return nil, err
	}

	rel, err := filepath.Rel(modDir, dir)
	if err != nil {
		return nil, err
	}

	importPath := modPath
	if rel != "." {
		importPath = path.Join(modPath, filepath.ToSlash(rel))
	}

	// the sources on disk belong to the developer, so they are trusted
	return LoadModules(importPath, Module{Path: modPath, FS: os.DirFS(modDir), Dir: modDir, Trusted: true})
}

// LoadModules loads the package main and the packages, which it imports, from the modules. It needs neither a Go
// installation nor the go command: the library of the subset and the allowed packages of the standard library
// are built in. A package is read from the module with the longest matching path, so a module may provide its
// own header packages, if it is trusted, see [Module.Trusted].
func LoadModules(main string, modules ...Module) (*Program, error) {
	std, err := fs.Sub(stdFS, "std")
	if err != nil {
		return nil, err
	}
	lib, err := fs.Sub(libFS, "lib")
	if err != nil {
		return nil, err
	}
	modules = append(slices.Clone(modules), Module{Path: stdModule, FS: std, Trusted: true}, Module{Path: libModule, FS: lib, Trusted: true})
	// the longest path wins and a trusted module wins over an untrusted one with the same path, so that an
	// untrusted module can not replace a library
	slices.SortStableFunc(modules, func(a, b Module) int {
		if len(a.Path) != len(b.Path) {
			return len(b.Path) - len(a.Path)
		}
		switch {
		case a.Trusted && !b.Trusted:
			return -1
		case b.Trusted && !a.Trusted:
			return 1
		}
		return 0
	})

	fset := token.NewFileSet()
	l := &loader{
		fset:    fset,
		prog:    &Program{Fset: fset, byPath: map[string]*Package{}, sources: map[string][]byte{}, importNames: map[string]string{}},
		modules: modules,
		loading: map[string]bool{},
	}
	l.diag = &diagnoser{fset: fset, sources: l.prog.sources}

	p, err := l.load(main)
	if err != nil {
		return nil, err
	}

	l.prog.Main = p
	l.prog.Diagnostics = l.diag.sorted()
	return l.prog, nil
}

type loader struct {
	fset    *token.FileSet
	prog    *Program
	modules []Module
	loading map[string]bool
	diag    *diagnoser
	// importer is the package, whose imports are resolved
	importer *Package
}

// isStdPath tells, whether the path belongs to the standard library, whose first element has no dot.
func isStdPath(p string) bool {
	first, _, _ := strings.Cut(p, "/")
	return !strings.Contains(first, ".")
}

func (l *loader) Import(importPath string) (*types.Package, error) {
	if p := l.prog.byPath[importPath]; p != nil {
		return p.Types, nil
	}
	if importPath == "unsafe" {
		return types.Unsafe, nil
	}

	if isStdPath(importPath) {
		if _, ok := stdAllowed[importPath]; !ok {
			// the checker reports the import, an empty package avoids a cascade of follow-up errors
			return types.NewPackage(importPath, path.Base(importPath)), nil
		}
	}

	importer := l.importer
	p, err := l.load(importPath)
	l.importer = importer
	if err != nil {
		return nil, err
	}
	return p.Types, nil
}

// find returns the module and the directory of the package.
func (l *loader) find(importPath string) (Module, string, bool) {
	if isStdPath(importPath) {
		sub, err := fs.Sub(stubFS, "_stubs")
		if err != nil {
			return Module{}, "", false
		}
		return Module{FS: sub, Dir: "go", Trusted: true}, importPath, true
	}
	for _, m := range l.modules {
		if importPath == m.Path {
			return m, ".", true
		}
		if rest, ok := strings.CutPrefix(importPath, m.Path+"/"); ok {
			return m, rest, true
		}
	}
	return Module{}, "", false
}

func (l *loader) fileName(m Module, dir, name string) string {
	switch {
	case m.Dir != "":
		return filepath.Join(m.Dir, filepath.FromSlash(path.Join(dir, name)))
	case m.Path != "":
		return path.Join(m.Path, dir, name)
	}
	return path.Join(dir, name)
}

func (l *loader) load(importPath string) (*Package, error) {
	if l.loading[importPath] {
		return nil, fmt.Errorf("import cycle through %s", importPath)
	}
	l.loading[importPath] = true
	defer delete(l.loading, importPath)

	m, dir, ok := l.find(importPath)
	if !ok {
		return nil, fmt.Errorf("package %s is not in any module", importPath)
	}

	entries, err := fs.ReadDir(m.FS, dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read package %s: %w", importPath, err)
	}

	p := &Package{Path: importPath, Kind: KindUser}
	if isStdPath(importPath) {
		p.Kind = KindStd
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		src, err := fs.ReadFile(m.FS, path.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if !included(src) {
			continue
		}

		fname := l.fileName(m, dir, name)
		l.prog.sources[fname] = src
		l.prog.importNames[fname] = path.Join(importPath, name)
		f, err := parser.ParseFile(l.fset, fname, src, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			l.reportParseError(err)
			if f == nil {
				continue
			}
		}
		p.Files = append(p.Files, f)
		if p.Kind == KindUser {
			if pos, ok := headerDirectivePos(f); ok {
				if m.Trusted {
					p.Kind = KindHeader
				} else {
					l.diag.report(pos, pos+token.Pos(len(headerDirective)), CodeUntrusted, "the package "+importPath+" declares itself a header package, but its module is not trusted", "a header package is not checked and its JavaScript bindings run with all rights of the host, so only the application may provide it", "remove "+headerDirective+" and write the package in the subset, or use a library like go.wdy.de/nago/pkg/gostrict/lib/three")
				}
			}
		}
		if p.Name == "" {
			p.Name = f.Name.Name
		}
	}

	if len(p.Files) == 0 {
		return nil, fmt.Errorf("package %s has no Go files", importPath)
	}

	if p.Kind == KindHeader {
		// the checker needs the declarations only, which keeps the bodies free to use anything of Go
		for _, f := range p.Files {
			stubHeader(f)
		}
	}

	p.Info = &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Instances:  map[*ast.Ident]types.Instance{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Implicits:  map[ast.Node]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Scopes:     map[ast.Node]*types.Scope{},
	}

	conf := types.Config{
		Importer:  l,
		GoVersion: "go1.27",
		Error: func(err error) {
			if te, ok := err.(types.Error); ok && !l.isFollowUp(te) {
				l.diag.report(te.Pos, token.NoPos, CodeTypeError, te.Msg, "the code is not valid Go", "")
			}
		},
	}

	l.importer = p
	p.Types, _ = conf.Check(importPath, l.fset, p.Files, p.Info)
	if p.Types == nil {
		p.Types = types.NewPackage(importPath, p.Name)
	}

	l.prog.byPath[importPath] = p
	l.prog.Packages = append(l.prog.Packages, p)
	return p, nil
}

// included evaluates the build constraint of the file with the tag gostrict, so a file may be excluded from the
// subset with //go:build !gostrict, e.g. a generated registry, which uses the whole of Go.
func included(src []byte) bool {
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || (strings.HasPrefix(line, "//") && !constraint.IsGoBuild(line)) {
			continue
		}
		if !constraint.IsGoBuild(line) {
			return true
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			return true
		}
		return expr.Eval(func(tag string) bool { return tag == "gostrict" })
	}
	return true
}

// isFollowUp tells, whether the error is caused by a package of the standard library, which the checker reports
// as not importable, or by one of its functions, which is not in the subset.
func (l *loader) isFollowUp(err types.Error) bool {
	if l.importer != nil && l.importer.Kind != KindUser {
		// a header package must not depend on such a package, which its check reports
		return false
	}
	for imp := range l.forbiddenImports(err.Pos) {
		if err.Msg == "undefined: "+imp || strings.Contains(err.Msg, "undefined: "+imp+".") || strings.Contains(err.Msg, "\""+imp+"\"") {
			return true
		}
	}
	return false
}

// forbiddenImports returns the names of the packages, which are not allowed in the file at pos.
func (l *loader) forbiddenImports(pos token.Pos) map[string]bool {
	res := map[string]bool{}
	file := l.fset.File(pos)
	if file == nil {
		return res
	}
	f, err := parser.ParseFile(token.NewFileSet(), file.Name(), l.prog.sources[file.Name()], parser.ImportsOnly)
	if err != nil {
		return res
	}
	for _, imp := range f.Imports {
		p := strings.Trim(imp.Path.Value, `"`)
		if isStdPath(p) {
			// a function of an allowed package, which is not in the subset, is missing in its stub
			res[importName(imp)] = true
		}
	}
	return res
}

func (l *loader) reportParseError(err error) {
	var list scanner.ErrorList
	if !errors.As(err, &list) {
		l.diag.diags = append(l.diag.diags, Diagnostic{Code: CodeTypeError, Msg: err.Error(), Why: "the code is not valid Go"})
		return
	}

	for _, e := range list {
		l.diag.diags = append(l.diag.diags, Diagnostic{
			Code:   CodeTypeError,
			File:   e.Pos.Filename,
			Line:   e.Pos.Line,
			Col:    e.Pos.Column,
			Msg:    e.Msg,
			Why:    "the code is not valid Go",
			Source: l.diag.line(e.Pos.Filename, e.Pos.Line),
		})
	}
}

// headerDirectivePos returns the position of the directive //gostrict:header in the documentation of the file.
func headerDirectivePos(f *ast.File) (token.Pos, bool) {
	if f.Doc == nil {
		return token.NoPos, false
	}
	for _, c := range f.Doc.List {
		if strings.TrimSpace(c.Text) == headerDirective {
			return c.Slash, true
		}
	}
	return token.NoPos, false
}

func findModule(dir string) (modDir, modPath string, err error) {
	for d := dir; ; {
		f, err := os.Open(filepath.Join(d, "go.mod"))
		if err == nil {
			defer f.Close()
			s := bufio.NewScanner(f)
			for s.Scan() {
				if rest, ok := strings.CutPrefix(strings.TrimSpace(s.Text()), "module "); ok {
					return d, strings.Trim(strings.TrimSpace(rest), `"`), nil
				}
			}
			return "", "", fmt.Errorf("%s/go.mod has no module path", d)
		}

		parent := filepath.Dir(d)
		if parent == d {
			return "", "", fmt.Errorf("no go.mod found for %s", dir)
		}
		d = parent
	}
}
