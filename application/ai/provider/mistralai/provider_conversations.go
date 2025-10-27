// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"
	"time"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/xtime"
)

var _ provider.Conversations = (*mistralConversations)(nil)

type mistralConversations struct {
	parent *mistralProvider
}

func (p *mistralConversations) client() *Client {
	return p.parent.client()
}

func (p *mistralConversations) All(subject auth.Subject) iter.Seq2[conversation.Conversation, error] {
	return func(yield func(conversation.Conversation, error) bool) {
		list, err := p.client().ListConversations()
		if err != nil {
			yield(conversation.Conversation{}, err)
			return
		}

		for _, info := range list {
			if !yield(info.IntoConversation(), nil) {
				return
			}
		}
	}
}

func (p *mistralConversations) FindByID(subject auth.Subject, id conversation.ID) (option.Opt[conversation.Conversation], error) {
	resp, err := p.client().GetConversation(string(id))
	if err != nil {
		return option.Opt[conversation.Conversation]{}, err
	}

	return option.Some(resp.IntoConversation()), nil
}

func (p *mistralConversations) Delete(subject auth.Subject, id conversation.ID) error {
	return p.client().DeleteConversation(string(id))
}

func (p *mistralConversations) Create(subject auth.Subject, opts conversation.CreateOptions) (conversation.Conversation, error) {
	res, err := p.client().CreateConversation(CreateConversationRequest{
		AgentID:      string(opts.Agent),
		Model:        string(opts.Model),
		Description:  opts.Description,
		Name:         opts.Name,
		Instructions: opts.Instructions,
		Store:        opts.CloudStore,
		Stream:       false,
		Inputs:       convInputToMistralInput(opts.Input),
	})

	if err != nil {
		return conversation.Conversation{}, err
	}

	conv := res.IntoConversation()
	conv.CreatedBy = subject.ID()
	conv.CreatedAt = xtime.UnixMilliseconds(time.Now().UnixMilli())
	conv.Agent = opts.Agent
	conv.Name = opts.Name
	conv.Model = opts.Model
	conv.Description = opts.Description
	conv.Instructions = opts.Instructions
	conv.CloudStore = opts.CloudStore

	// TODO response already contains already an arbitrary set of response messages

	return conv, nil
}

func (p *mistralConversations) Messages(subject auth.Subject, id conversation.ID) provider.Messages {
	return &mistralMessages{
		id:     id,
		parent: p.parent,
	}
}

func convInputToMistralInput(contents []message.Content) []Input {
	var inputs []Input
	for _, content := range contents {
		if content.Text.IsSome() {
			inputs = append(inputs, MessageInputEntry{
				Content: TextChunk{Text: content.Text.Unwrap()},
				Role:    RoleUser,
			})
		}
	}

	return inputs
}
