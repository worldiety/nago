// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"fmt"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/auth"
)

func NewDeleteByID(repoWS Repository, repoAgents agent.Repository) DeleteByID {
	return func(subject auth.Subject, id ID) error {
		if err := subject.AuditResource(repoWS.Name(), string(id), PermDeleteByID); err != nil {
			return err
		}

		optWs, err := repoWS.FindByID(id)
		if err != nil {
			return err
		}

		if optWs.IsNone() {
			return nil
		}

		ws := optWs.Unwrap()
		for _, aid := range ws.Agents {
			if err := repoAgents.DeleteByID(aid); err != nil {
				return fmt.Errorf("failed to delete referenced agent '%s': %w", aid, err)
			}
		}

		if err := repoWS.DeleteByID(id); err != nil {
			return err
		}

		return nil
	}
}
