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
	"slices"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Conversations = (*cacheConversations)(nil)

type cacheConversations struct {
	parent *Provider
}

func (c *cacheConversations) All(subject auth.Subject) iter.Seq2[conversation.Conversation, error] {
	return func(yield func(conversation.Conversation, error) bool) {
		var tmp []conversation.Conversation
		for key, err := range c.parent.idxProvConversations.AllByPrimary(context.Background(), c.parent.Identity()) {
			if err != nil {
				if !yield(conversation.Conversation{}, err) {
					return
				}

				continue
			}

			optConv, err := c.parent.repoConversations.FindByID(key.Secondary)
			if err != nil {
				if !yield(conversation.Conversation{}, err) {
					return
				}

				continue
			}

			if optConv.IsNone() {
				continue // stale ref
			}

			m := optConv.Unwrap()

			if m.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoConversations.Name()), rebac.Instance(m.ID), PermConversationFindAll) {
				continue
			}

			tmp = append(tmp, m)
		}

		slices.SortFunc(tmp, func(a, b conversation.Conversation) int {
			return int(b.CreatedAt - a.CreatedAt)
		})

		for _, t := range tmp {
			if !yield(t, nil) {
				return
			}
		}
	}
}

func (c *cacheConversations) FindByID(subject auth.Subject, id conversation.ID) (option.Opt[conversation.Conversation], error) {
	optConv, err := c.parent.repoConversations.FindByID(id)
	if err != nil {
		return option.Opt[conversation.Conversation]{}, err
	}

	if optConv.IsNone() {
		return optConv, nil
	}

	conv := optConv.Unwrap()
	if conv.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoModels.Name()), rebac.Instance(conv.ID), PermConversationFindByID) {
		return option.Opt[conversation.Conversation]{}, subject.Audit(PermConversationFindByID)
	}

	return option.Some(conv), nil
}

func (c *cacheConversations) Delete(subject auth.Subject, id conversation.ID) error {
	optLib, err := c.parent.repoConversations.FindByID(id)
	if err != nil {
		return err
	}

	if optLib.IsNone() {
		return nil
	}

	lib := optLib.Unwrap()
	if lib.CreatedBy != subject.ID() && !subject.HasResourcePermission(rebac.Namespace(c.parent.repoModels.Name()), rebac.Instance(lib.ID), PermConversationDelete) {
		return subject.Audit(PermConversationDelete)
	}

	if err := c.parent.prov.Conversations().Unwrap().Delete(subject, id); err != nil {
		return err
	}

	if err := c.parent.repoConversations.DeleteByID(id); err != nil {
		return err
	}

	if err := c.parent.idxMsg.DeleteAllPrimary(context.Background(), id); err != nil {
		return err
	}

	if err := c.parent.idxProvConversations.Delete(context.Background(), c.parent.Identity(), id); err != nil {
		return err
	}

	return nil
}

func (c *cacheConversations) Create(subject auth.Subject, opts conversation.CreateOptions) (conversation.Conversation, []message.Message, error) {
	if err := subject.Audit(PermConversationCreate); err != nil {
		return conversation.Conversation{}, nil, err
	}

	conv, msgs, err := c.parent.prov.Conversations().Unwrap().Create(subject, opts)
	if err != nil {
		return conversation.Conversation{}, nil, err
	}

	if conv.CreatedAt == 0 {
		conv.CreatedAt = xtime.Now()
	}

	conv.CreatedBy = subject.ID()
	if conv.Identity() == "" {
		return conversation.Conversation{}, nil, fmt.Errorf("provider returned empty identity")
	}

	if opt, err := c.parent.repoConversations.FindByID(conv.ID); err != nil || opt.IsSome() {
		if err != nil {
			return conversation.Conversation{}, nil, err
		}

		return conversation.Conversation{}, nil, fmt.Errorf("provider returned an existing library: %s", conv.ID)
	}

	if err := c.parent.repoConversations.Save(conv); err != nil {
		return conversation.Conversation{}, nil, err
	}

	if err := c.parent.idxProvConversations.Put(c.parent.Identity(), conv.ID); err != nil {
		return conversation.Conversation{}, nil, err
	}

	for _, msg := range msgs {
		if msg.CreatedAt == 0 {
			msg.CreatedAt = xtime.Now()
		}

		msg.CreatedBy = subject.ID()
		if err := c.parent.repoMessages.Save(msg); err != nil {
			return conversation.Conversation{}, nil, err
		}

		if err := c.parent.idxMsg.Put(conv.ID, msg.ID); err != nil {
			return conversation.Conversation{}, nil, fmt.Errorf("failed to put conversation/message tuple into index: %w", err)
		}
	}

	return conv, msgs, nil
}

func (c *cacheConversations) Conversation(subject auth.Subject, id conversation.ID) provider.Conversation {
	return &cacheConversation{c.parent, id}
}
