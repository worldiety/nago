// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import "go.wdy.de/nago/application/permission"

var (
	PermSync = permission.DeclareSync[Sync]("nago.ai.mistral.sync", "Mistral AI Workspace")
)
