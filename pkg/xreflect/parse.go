// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xreflect

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"log/slog"
	"maps"
	"path"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
)

var modInfo *debug.BuildInfo
var importedPackages = map[string]*parsedPackage{}
var packages = map[string]*Package{}
var typeLookup = map[reflect.Type]Type{}
var mutex sync.Mutex
var typesInPackage = map[string][]Type{}

func init() {
	mi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("non-module builds are unsupported")
	}

	modInfo = mi
}

type parsedPackage struct {
	pkg    *types.Package
	astPkg *ast.Package
	info   *types.Info
}

func makeCommentString(groups ...*ast.CommentGroup) string {
	var tmp strings.Builder
	for _, group := range groups {
		tmp.WriteString(group.Text())
	}

	// TODO remove macro and third party lint notations
	return tmp.String()
}

func Import(importPrefix string, fsys fs.FS) error {
	pkgs, err := parse(importPrefix, fsys)
	if err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()
	for _, pkg := range pkgs {
		importedPackages[pkg.pkg.Path()] = pkg
	}

	// provide the complete tree before assembling
	for _, pkg := range pkgs {
		assembleTypes(pkg)
	}

	return nil
}

func ModName() string {
	return modInfo.Main.Path
}

// Parse reads all go Files from the given fs and resolves all import paths according to the running go module.
func parse(importPrefix string, fsys fs.FS) ([]*parsedPackage, error) {
	fset := token.NewFileSet()
	pkgs := make(map[string]*ast.Package)

	err := fs.WalkDir(fsys, ".", func(wpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(wpath) == "_test.go" {
			return nil
		}

		if filepath.Ext(wpath) == ".go" {
			buf, err := fs.ReadFile(fsys, wpath)
			if err != nil {
				return fmt.Errorf("cannot read %s: %w", wpath, err)
			}

			file, err := parser.ParseFile(fset, wpath, buf, parser.AllErrors|parser.ParseComments)
			if err != nil {
				return err
			}

			pkgName := file.Name.Name
			importPath := path.Join(importPrefix, path.Dir(wpath))
			if pkgs[importPath] == nil {
				p := &ast.Package{
					Name:  pkgName,
					Files: make(map[string]*ast.File),
				}
				pkgs[importPath] = p

			}
			pkgs[importPath].Files[path.Join(importPath, path.Base(wpath))] = file
		}
		return nil

	})

	if err != nil {
		return nil, fmt.Errorf("cannot parse %v: %w", fsys, err)
	}

	var res []*parsedPackage
	conf := types.Config{Importer: importer.Default()}
	for importPath, pkg := range pkgs {
		info := &types.Info{
			Types:        make(map[ast.Expr]types.TypeAndValue),
			Instances:    map[*ast.Ident]types.Instance{},
			Defs:         make(map[*ast.Ident]types.Object),
			Uses:         make(map[*ast.Ident]types.Object),
			Implicits:    map[ast.Node]types.Object{},
			Selections:   make(map[*ast.SelectorExpr]*types.Selection),
			Scopes:       make(map[ast.Node]*types.Scope),
			InitOrder:    []*types.Initializer{},
			FileVersions: map[*ast.File]string{},
		}

		tpkg, err := conf.Check(importPath, fset, slices.Collect(maps.Values(pkg.Files)), info)
		if err != nil {
			slog.Error("failed type checking", "err", err)
		}

		// either way, append that
		res = append(res, &parsedPackage{
			pkg:    tpkg,
			info:   info,
			astPkg: pkg,
		})

	}

	return res, nil
}
