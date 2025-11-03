// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package libsync

import (
	"go.wdy.de/nago/application/ai/library"
	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/permission"
)

var (
	PermSynchronize = permission.DeclareSync[Synchronize]("nago.ai.libsync.synchronize", "AI Library Sync")
)

var Permissions = ent.DeclarePermissions[Job, library.ID]("nago.ai.synclib.job", "AI Sync Job")
