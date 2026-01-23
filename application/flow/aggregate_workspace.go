// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/pkg/cloner"
)

type Workspace struct {
	ID           WorkspaceID
	Packages     *Packages
	Repositories *Repositories
	Forms        *Forms
	Name         Ident
	Description  string
	// Extensions can be used by custom events and commands to introduce additional functionality.
	// We cannot know the according types, thus this has to be type-unsafe.
	Extensions map[string]cloner.Cloneable
}

func (a *Workspace) Clone() *Workspace {
	xClone := map[string]cloner.Cloneable{}
	for k, v := range a.Extensions {
		xClone[k] = v.Clone()
	}

	return &Workspace{
		ID:           a.ID,
		Packages:     a.Packages.Clone(),
		Repositories: a.Repositories.Clone(),
		Forms:        a.Forms.Clone(),
		Name:         a.Name,
		Description:  a.Description,
		Extensions:   xClone,
	}
}

func (a *Workspace) Identity() WorkspaceID {
	return a.ID
}
