// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/auth"
)

type ID string

type Conversation struct {
	ID string
}

type StartOptions struct {
	Agent agent.ID

	Name        string
	Description string

	// CloudStore indicates if the conversation should be stored and retrievable if the provider uses a cloud
	// backend.
	CloudStore bool
}

type Start func(subject auth.Subject) (ID, error)
