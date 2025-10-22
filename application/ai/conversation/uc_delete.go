// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"context"
	"fmt"

	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
)

func NewDelete(bus events.Bus, repo Repository, repoMsg message.Repository, idxConvMsg *data.CompositeIndex[ID, message.ID]) Delete {
	return func(subject auth.Subject, id ID) error {
		optConv, err := repo.FindByID(id)
		if err != nil {
			return err
		}

		if optConv.IsNone() {
			return nil
		}

		conversation := optConv.Unwrap()

		if subject.ID() != conversation.CreatedBy && !subject.HasResourcePermission(repo.Name(), string(conversation.ID), PermDelete) {
			return subject.Audit(PermFindAll)
		}

		if err := repo.DeleteByID(conversation.ID); err != nil {
			return err
		}

		for key, err := range idxConvMsg.AllByPrimary(context.Background(), conversation.ID) {
			if err != nil {
				return err
			}

			if err := repoMsg.DeleteByID(key.Secondary); err != nil {
				return fmt.Errorf("failed to delete msg: %w", err)
			}
		}

		if err := idxConvMsg.DeleteAllPrimary(context.Background(), conversation.ID); err != nil {
			return fmt.Errorf("failed to delete all index conv-msg entries: %w", err)
		}

		bus.Publish(Deleted{Conversation: conversation.ID})

		return nil
	}
}
