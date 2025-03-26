// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package xreflect

import (
	"go/ast"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type Package struct {
	doc           string
	parsedPackage *parsedPackage
}

func (p *Package) Doc() string {
	return p.doc
}

func (p *Package) Name() string {
	return p.parsedPackage.pkg.Name()
}

func (p *Package) Path() string {
	return p.parsedPackage.pkg.Path()
}

func (p *Package) Types() []Type {
	return slices.Clone(typesInPackage[p.parsedPackage.pkg.Path()])
}

// Packages returns a slice of all known package import paths which have been imported.
func Packages() []*Package {
	mutex.Lock()
	defer mutex.Unlock()

	pkgs := slices.Collect(maps.Values(packages))
	slices.SortFunc(pkgs, func(a, b *Package) int {
		return strings.Compare(a.Path(), b.Path())
	})

	return pkgs
}

type Type interface {
	Doc() string
	// Path returns the import path for the type.
	Path() string
	// Name of the type
	Name() string
}

func Types(path string) []Type {
	mutex.Lock()
	defer mutex.Unlock()

	return typesInPackage[path]
}

func assembleTypes(pkg *parsedPackage) {
	var allFileComments []*ast.CommentGroup
	for _, file := range pkg.astPkg.Files {
		if file.Doc != nil {
			allFileComments = append(allFileComments, file.Doc)
		}
	}

	packages[pkg.pkg.Path()] = &Package{
		doc:           makeCommentString(allFileComments...),
		parsedPackage: pkg,
	}

	var res []Type
	for _, file := range pkg.astPkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			// these are just normal functions, may be associated to types
			case *ast.FuncDecl:
				if !decl.Name.IsExported() && !(decl.Name.Name == "main" && pkg.pkg.Name() == "main") {
					continue
				}

				if decl.Recv == nil || len(decl.Recv.List) == 0 {
					// this is free package-level function
					// TODO this is likely a factory/constructor, main or helper or just a typeless usecase
				}

			// these are new types
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if !spec.Name.IsExported() {
							continue
						}
						switch typ := spec.Type.(type) {

						case *ast.FuncType:
							if !spec.Name.IsExported() {
								continue
							}
							res = append(res, &Func{
								doc:  makeCommentString(decl.Doc, spec.Doc),
								path: pkg.pkg.Path(),
								name: spec.Name.Name,
							})
						case *ast.StructType:
							if !spec.Name.IsExported() {
								continue
							}

							var fields []*Field
							if typ.Fields != nil {
								for _, field := range typ.Fields.List {
									tag := ""
									if field.Tag != nil {
										t, err := strconv.Unquote(field.Tag.Value)
										if err != nil {
											tag = field.Tag.Value
										} else {
											tag = t
										}
									}
									for _, name := range field.Names {
										if !name.IsExported() {
											continue
										}

										fields = append(fields, &Field{
											doc:  makeCommentString(field.Doc),
											name: name.Name,
											tag:  tag,
										})
									}

								}
							}
							res = append(res, &Struct{
								doc:    makeCommentString(decl.Doc, spec.Doc),
								name:   spec.Name.Name,
								path:   pkg.pkg.Path(),
								fields: fields,
							})
						default:
							res = append(res, &Incomplete{
								doc:    makeCommentString(decl.Doc),
								path:   pkg.pkg.Path(),
								name:   spec.Name.Name,
								origin: typ,
							})
						}
					}
				}
			default:
				slog.Error("unsupported top level decl type", "t", reflect.TypeOf(decl))
			}
		}
	}

	slices.SortFunc(res, func(a, b Type) int {
		return strings.Compare(a.Name(), b.Name())
	})
	typesInPackage[pkg.pkg.Path()] = res
}

type Incomplete struct {
	doc    string
	path   string
	name   string
	origin ast.Expr
	error  error
}

func (t *Incomplete) Doc() string {
	return t.doc
}

func (t *Incomplete) Path() string {
	return t.path
}

func (t *Incomplete) Name() string {
	return t.name
}

type Func struct {
	doc  string
	path string
	name string
}

func (f *Func) Doc() string {
	return f.doc
}

func (f *Func) Path() string {
	return f.path
}

func (f *Func) Name() string {
	return f.name
}

type Struct struct {
	doc    string
	path   string
	name   string
	fields []*Field
}

type Field struct {
	doc  string
	name string
	tag  string
}

func (f Field) Doc() string {
	return f.doc
}

func (f Field) Name() string {
	return f.name
}

func (f Field) Tag() reflect.StructTag {
	return reflect.StructTag(f.tag)
}

func (s *Struct) Doc() string {
	return s.doc
}

func (s *Struct) Path() string {
	return s.path
}

func (s *Struct) Name() string {
	return s.name
}

func (s *Struct) Fields() []*Field {
	return s.fields
}
