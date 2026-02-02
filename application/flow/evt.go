// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"slices"

	"go.wdy.de/nago/application/evs"
)

type WorkspaceEvent interface {
	evs.Evt[*Workspace]
	WorkspaceID() WorkspaceID
}

var DefaultEvents = slices.Values([]WorkspaceEvent{
	TypeFieldAppended{},
	BoolFieldAppended{},
	FormCreated{},
	FormTextAdded{},
	PackageCreated{},
	PrimaryKeySelected{},
	RepositoryAssigned{},
	StringEnumCaseAdded{},
	StringFieldAppended{},
	StringTypeCreated{},
	StructTypeCreated{},
	WorkspaceCreated{},
	FormVStackAdded{},
	FormViewDeleted{},
	FormDeleted{},
	FormCheckboxAdded{},
	FormAlignmentUpdated{},
	FormVisibleExprUpdated{},
	FormButtonAdded{},
	ButtonStyleUpdated{},
	FormActionExprUpdated{},
	FormEnableExprUpdated{},
	FormGapUpdated{},
	FormHStackAdded{},
	FormHLineAdded{},
	FormBackgroundColorUpdated{},
})
