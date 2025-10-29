// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cache

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"os"

	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Conversation = (*cacheConversation)(nil)

type cacheConversation struct {
	parent *Provider
	id     conversation.ID
}

func (c cacheConversation) Identity() conversation.ID {
	return c.id
}

func (c cacheConversation) All(subject auth.Subject) iter.Seq2[message.Message, error] {
	return func(yield func(message.Message, error) bool) {
		optConv, err := c.parent.repoConversations.FindByID(c.id)
		if err != nil {
			yield(message.Message{}, err)
			return
		}

		if optConv.IsNone() {
			return
		}

		conv := optConv.Unwrap()

		for key := range c.parent.idxMsg.AllByPrimary(context.Background(), c.id) {
			optMsg, err := c.parent.repoMessages.FindByID(key.Secondary)
			if err != nil {
				yield(message.Message{}, err)
				return
			}

			if optMsg.IsNone() {
				slog.Info("found stale message id in conversation/message index", "conversation", c.id, "message", key.Secondary)
				continue
			}

			msg := optMsg.Unwrap()

			if conv.CreatedBy == subject.ID() || msg.CreatedBy == subject.ID() || subject.HasResourcePermission(c.parent.repoConversations.Name(), string(c.id), PermMessageFindAll) {
				if !yield(msg, nil) {
					return
				}
			}
		}

	}
}

func (c cacheConversation) Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error) {
	optConv, err := c.parent.repoConversations.FindByID(c.id)
	if err != nil {
		return nil, err
	}

	if optConv.IsNone() {
		return []message.Message{}, fmt.Errorf("conversation to append message to does not exists: %s: %w", c.id, os.ErrNotExist)
	}

	conv := optConv.Unwrap()
	if conv.CreatedBy != subject.ID() && !subject.HasResourcePermission(c.parent.repoConversations.Name(), string(c.id), PermMessageAppend) {
		return nil, subject.Audit(PermMessageAppend)
	}

	msgs, err := c.parent.prov.Conversations().Unwrap().Conversation(subject, c.id).Append(subject, opts)
	if err != nil {
		return nil, err
	}

	// TODO fix me: what about transient input messages???
	for _, msg := range msgs {
		if msg.CreatedAt == 0 {
			msg.CreatedAt = xtime.Now()
		}

		msg.CreatedBy = subject.ID()
		if err := c.parent.repoMessages.Save(msg); err != nil {
			return nil, err
		}

		if err := c.parent.idxMsg.Put(conv.ID, msg.ID); err != nil {
			return nil, fmt.Errorf("failed to put conversation/message tuple into index: %w", err)
		}
	}

	return msgs, nil
}
