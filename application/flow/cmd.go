// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import "go.wdy.de/nago/application/evs"

type WorkspaceCommand interface {
	evs.Cmd[*Workspace, WorkspaceEvent]
	WorkspaceID() WorkspaceID
}
