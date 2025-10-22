// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"fmt"
	"os"

	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/workspace"
	"go.wdy.de/nago/application/secret"
)

func delete(match secret.Match, findConvByID conversation.FindByID, findWSByID workspace.FindByID, syncConvRepo SyncConversationRepository, id conversation.ID) error {
	var syncConv SynchronizedConversation
	for synchronizedConversation, err := range syncConvRepo.All() {
		if err != nil {
			return err
		}

		if synchronizedConversation.ID == id {
			syncConv = synchronizedConversation
		}
	}

	if syncConv.Workspace == "" {
		return fmt.Errorf("cannot delete: workspace in synchronized conversation not found: %q: %w", id, os.ErrNotExist)
	}

	_, cl, err := createClientForWorkspace(match, findWSByID, syncConv.Workspace)
	if err != nil {
		return err
	}

	if err := cl.DeleteConversation(syncConv.CloudConversation); err != nil {
		return fmt.Errorf("error deleting mistral cloud conversation: %v", err)
	}

	if err := syncConvRepo.DeleteByID(syncConv.ID); err != nil {
		return fmt.Errorf("cannot delete conversation from sync repo: %w", err)
	}

	return nil
}
