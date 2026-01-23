// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package flow

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

func NewHandleCommand(handler *evs.Handler[*Workspace, WorkspaceEvent, WorkspaceID]) HandleCommand {
	return func(subject auth.Subject, cmd WorkspaceCommand) error {
		return handler.Handle(subject, cmd.WorkspaceID(), cmd)
	}
}
