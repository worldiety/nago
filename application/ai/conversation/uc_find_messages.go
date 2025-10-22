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
	"iter"
	"os"

	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
)

func NewFindMessages(repoConv Repository, repoMsg message.Repository, idxConvMsg *data.CompositeIndex[ID, message.ID]) FindMessages {
	return func(subject auth.Subject, cid ID) iter.Seq2[message.Message, error] {
		return func(yield func(message.Message, error) bool) {
			if cid == "" {
				return
			}

			optConv, err := repoConv.FindByID(cid)
			if err != nil {
				yield(message.Message{}, err)
				return
			}

			if optConv.IsNone() {
				yield(message.Message{}, fmt.Errorf("conversation %q not found: %w", cid, os.ErrNotExist))
				return
			}

			conversation := optConv.Unwrap()
			if subject.ID() != conversation.CreatedBy && !subject.HasResourcePermission(repoConv.Name(), string(conversation.ID), PermFindAll) {
				yield(message.Message{}, subject.Audit(PermFindAll))
				return
			}

			for key, err := range idxConvMsg.AllByPrimary(context.Background(), cid) {
				if err != nil {
					yield(message.Message{}, fmt.Errorf("failed to iterate on conversation-msg-index: %w", err))
					return
				}

				optMsg, err := repoMsg.FindByID(key.Secondary)
				if err != nil {
					yield(message.Message{}, fmt.Errorf("failed to load message %q: %w", key.Secondary, err))
					return
				}

				if optMsg.IsNone() {
					yield(message.Message{}, fmt.Errorf("conversation-message-index referenced stale message: %q", key.Secondary))
					continue
				}

				if !yield(optMsg.Unwrap(), nil) {
					return
				}
			}
		}
	}
}
