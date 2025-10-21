// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package conversation

import (
	"fmt"
	"os"
	"sync"

	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/eventstore"
	"go.wdy.de/nago/pkg/xslices"
	"go.wdy.de/nago/pkg/xtime"
)

func NewAppend(mutex *sync.Mutex, bus events.Bus, repo Repository, repoMsg message.Repository, idx *data.CompositeIndex[ID, message.ID]) Append {
	return func(subject auth.Subject, opts AppendOptions) (message.ID, error) {
		if !subject.HasPermission(PermAppend) && !subject.HasResourcePermission(repo.Name(), string(opts.Conversation), PermAppend) {
			return "", subject.Audit(PermAppend)
		}

		mutex.Lock()
		defer mutex.Unlock()

		if len(opts.Input) == 0 {
			return "", fmt.Errorf("input must not be empty")
		}

		optConv, err := repo.FindByID(opts.Conversation)
		if err != nil {
			return "", err
		}

		if optConv.IsNone() {
			return "", fmt.Errorf("conversation is gone %q: %w", opts.Conversation, os.ErrNotExist)
		}

		msg := message.Message{
			ID:        message.ID(eventstore.NewID()),
			CreatedAt: xtime.Now(),
			CreatedBy: subject.ID(),
			Inputs:    xslices.New(opts.Input...),
		}

		if optMsg, err := repoMsg.FindByID(msg.ID); err != nil || optMsg.IsSome() {
			if err != nil {
				return "", err
			}

			return "", fmt.Errorf("message already exists: %q", msg.ID)
		}

		if err := repoMsg.Save(msg); err != nil {
			return "", err
		}

		if err := idx.Put(opts.Conversation, msg.ID); err != nil {
			return "", fmt.Errorf("failed to index message %q: %w", msg.ID, err)
		}

		bus.Publish(MessageAppended{
			Conversation: opts.Conversation,
			Message:      msg.ID,
		})

		return msg.ID, nil
	}
}
