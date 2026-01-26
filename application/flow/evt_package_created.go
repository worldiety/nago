// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"context"
	"fmt"
	"strings"

	"go.wdy.de/nago/application/evs"
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

type PackageCreated struct {
	Workspace   WorkspaceID `json:"workspace,omitempty"`
	Package     PackageID   `json:"package,omitempty"`
	Path        ImportPath  `json:"path,omitempty"`
	Name        Ident       `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
}

func (evt PackageCreated) Evolve(ctx context.Context, ws *Workspace) error {
	pkg := NewPackage(evt.Package, evt.Path, evt.Name)
	pkg.Description = evt.Description

	ws.Packages.AddPackage(pkg)
	return nil
}

func (evt PackageCreated) Discriminator() evs.Discriminator {
	return "PackageCreated"
}

func (evt PackageCreated) WorkspaceID() WorkspaceID {
	return evt.Workspace
}

func (evt PackageCreated) event() {}
