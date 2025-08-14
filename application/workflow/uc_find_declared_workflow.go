// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workflow

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindDeclaredWorkflow(declarations *concurrent.RWMap[ID, *workflow]) FindDeclaredWorkflow {
	return func(subject user.Subject, id ID) (option.Opt[DeclareOptions], error) {
		if err := subject.AuditResource(RepositoryNameDeclaredWorkflows, string(id), PermFindDeclaredWorkflows); err != nil {
			return option.Opt[DeclareOptions]{}, err
		}

		decl, ok := declarations.Get(id)
		if !ok {
			return option.Opt[DeclareOptions]{}, nil
		}

		return option.Some(decl.opts), nil
	}
}
