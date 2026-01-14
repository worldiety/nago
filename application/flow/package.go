// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"
)

type PackageID string

type ImportPath string

func (p ImportPath) Validate() error {
	s := string(p)
	if s == "" {
		return fmt.Errorf("package path cannot be empty")
	}

	segments := strings.Split(s, "/")
	for i, segment := range segments {
		if segment == "" {
			return fmt.Errorf("empty segment at position %d", i)
		}
		if err := validateIdentifier(segment); err != nil {
			return fmt.Errorf("invalid segment %q at position %d: %w", segment, i, err)
		}
	}

	return nil
}

func validateIdentifier(s string) error {
	if s == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return fmt.Errorf("must start with letter or underscore")
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return fmt.Errorf("invalid character %q", r)
			}
		}
	}

	if token.Lookup(s).IsKeyword() {
		return fmt.Errorf("%q is a reserved keyword", s)
	}

	return nil
}

type PackageCreated struct {
	Workspace   WorkspaceID
	Package     PackageID
	Path        ImportPath
	Name        Ident
	Description string
}

func (e PackageCreated) WorkspaceID() WorkspaceID {
	return e.Workspace
}

type CreatePackageCmd struct {
	Package     PackageID
	Path        ImportPath
	Name        Ident
	Description string
}

type Package struct {
	Package     PackageID
	Path        ImportPath
	Name        Ident
	Description string
}
