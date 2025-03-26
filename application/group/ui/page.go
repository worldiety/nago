// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uigroup

import (
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/rcrud"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui/crud"
	"iter"
)

type Pages struct {
	Groups core.NavigationPath
}

func Groups(wnd core.Window, useCases group.UseCases) core.View {
	uc := rcrud.UseCasesFrom(&rcrud.Funcs[group.Group, group.ID]{
		PermFindByID:   group.PermFindByID,
		PermFindAll:    group.PermFindAll,
		PermDeleteByID: group.PermDelete,
		PermCreate:     group.PermCreate,
		PermUpdate:     group.PermUpdate,
		FindByID: func(subject auth.Subject, id group.ID) (std.Option[group.Group], error) {
			return useCases.FindByID(subject, id)
		},
		FindAll: func(subject auth.Subject) iter.Seq2[group.Group, error] {
			return useCases.FindAll(subject)
		},
		DeleteByID: func(subject auth.Subject, id group.ID) error {
			return useCases.Delete(subject, id)
		},
		Create: func(subject auth.Subject, e group.Group) (group.ID, error) {
			return useCases.Create(subject, e)
		},
		Update: func(subject auth.Subject, e group.Group) error {
			return useCases.Update(subject, e)
		},
		Upsert: func(subject auth.Subject, e group.Group) (group.ID, error) {
			return useCases.Upsert(subject, e)
		},
	})

	return crud.AutoRootView(crud.AutoRootViewOptions{Title: "Gruppen"}, uc)(wnd)
}
