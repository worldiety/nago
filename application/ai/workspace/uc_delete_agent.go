// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package workspace

import (
	"fmt"
	"os"
	"slices"
	"sync"

	"go.wdy.de/nago/application/ai/agent"
	"go.wdy.de/nago/auth"
)

func NewDeleteAgent(mutex *sync.Mutex, repoWS Repository, repoAgents agent.Repository) DeleteAgent {
	return func(subject auth.Subject, parent ID, aid agent.ID) error {
		if err := subject.AuditResource(repoWS.Name(), string(parent), PermDeleteAgent); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		optParent, err := repoWS.FindByID(parent)
		if err != nil {
			return err
		}

		if optParent.IsNone() {
			return fmt.Errorf("parent workspace is gone: %w", os.ErrNotExist)
		}

		ws := optParent.Unwrap()

		ws.Agents = slices.DeleteFunc(ws.Agents, func(id agent.ID) bool {
			return id == aid
		})

		if err := repoAgents.DeleteByID(aid); err != nil {
			return fmt.Errorf("failed to delete agent: %w", err)
		}

		if err := repoWS.Save(ws); err != nil {
			return fmt.Errorf("failed to save workspace: %w", err)
		}

		return nil
	}
}
