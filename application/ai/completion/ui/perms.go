// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"go.wdy.de/nago/application/ai"
	"go.wdy.de/nago/application/ai/provider/cache"
	"go.wdy.de/nago/application/ai/session"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// RequiredPermissions are the framework permissions a subject needs so that a [Chat] or [ChatButton] works
// for them: list the configured providers, list the models of the chosen provider, and start a session.
//
// Everything beyond that is granted per instance when a session is created (see session.NewCreate), so the
// user can continue, rename and delete their own conversations without holding anything globally. And the
// tools an agent offers are bounded by the acting subject anyway, so this list grants no domain access
// whatsoever - it only makes the assistant appear.
//
// Do not enumerate these by hand in an application. Assign the system role declared by cfgai instead
// (see cfgai.RoleAssistantUser), which is kept in sync with this list.
func RequiredPermissions() []permission.ID {
	return []permission.ID{
		ai.PermFindAllProvider,
		cache.PermFindAllModel,
		session.PermCreate,
	}
}

// MissingPermissions reports which of the [RequiredPermissions] the subject does not hold. An empty result
// means the assistant can start.
//
// It exists because the failure mode is otherwise invisible: without these permissions the chat button simply
// does not render, and an operator has no way to tell an unconfigured provider from a missing role.
func MissingPermissions(subject auth.Subject) []permission.ID {
	if subject == nil || !subject.Valid() {
		return RequiredPermissions()
	}

	var missing []permission.ID
	for _, pid := range RequiredPermissions() {
		if !subject.HasPermission(pid) {
			missing = append(missing, pid)
		}
	}

	return missing
}
