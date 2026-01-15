// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

type WorkspaceEvent interface {
	WorkspaceID() WorkspaceID
	event()
}

type WorkspaceCommand[T any] interface {
	WorkspaceID() WorkspaceID
	WithWorkspaceID(id WorkspaceID) T
}
