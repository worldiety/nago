// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package vmcode

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Source is the checked program, which [Compile] compiles, as package gostrict passes it.
type Source struct {
	Fset *token.FileSet
	// Packages are all packages, each after the packages it imports.
	Packages []*Package
	// Exports are the functions, which the host calls, by their name at the boundary.
	Exports map[string]*types.Func
	// Hosted are the full names of the functions of header packages, which the host implements, e.g.
	// "example.com/ui.Text" or "example.com/ui.TText.Font" for a method.
	Hosted map[string]bool
	// Immutable tells, whether the program never changes a value of the struct type in place.
	Immutable func(types.Type) bool
	// Records are the names of the fields of the records of the boundary in their order by the full name of the
	// record.
	Records map[string][]string
	// Unions are the variants of the unions of the boundary by the full name of the union, and whether they are
	// handles.
	Unions map[string]Union
}

// Union is a sealed interface of the boundary.
type Union struct {
	Variants []string
	Handles  bool
}

// Package is a checked package.
type Package struct {
	Path string
	// User is set for a package of the subset, whose functions are compiled; the others are header packages and
	// the standard library, whose functions the host or the VM implement.
	User  bool
	Files []*ast.File
	Types *types.Package
	Info  *types.Info
}
