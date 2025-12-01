// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package aam provides an abstract-application-model
package aam

import (
	"go.wdy.de/nago/app/builder/environment"
	"go.wdy.de/nago/pkg/data/mem"
	"go.wdy.de/nago/presentation/core"
)

type Struct struct {
	Namespace environment.Ident
	Name      environment.Ident
}

func (s *Struct) Identity() environment.Ident {
	return s.Name
}

type App struct {
	ID         core.ApplicationID
	GitRepoURL core.URI
	Namespaces mem.Repository[*Namespace, environment.Ident]
}

type Namespace struct {
	Name    environment.Ident
	Structs mem.Repository[*Struct, environment.Ident]
}

func (n *Namespace) Identity() environment.Ident {
	return n.Name
}
