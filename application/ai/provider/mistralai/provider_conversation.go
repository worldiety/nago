// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Conversation = (*mistralMessages)(nil)

type mistralMessages struct {
	id     conversation.ID
	parent *mistralProvider
}

func (p *mistralMessages) Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error) {
	var tmp []Input
	if opts.MessageInput.IsSome() {
		tmp = append(tmp, MessageInputEntry{
			Content: TextChunk{Text: opts.MessageInput.Unwrap()},
			Role:    "user", // TODO is that always correct?
		})
	}

	resp, err := p.client().AppendConversation(string(p.id), AppendConversationRequest{
		Inputs: tmp,
		Store:  opts.CloudStore,
		Stream: false,
	})

	if err != nil {
		return nil, err
	}

	var msgs []message.Message

	if opts.Role == "" {
		opts.Role = "user" // this API design makes me crazy, nothing is explained
	}
	// we don't know, because the Mistral API is broken, again. How do I get this, without requesting the entire thread?
	msgs = append(msgs, message.Message{
		ID:            "", // TODO make this transient?
		CreatedAt:     0,
		CreatedBy:     "",
		Role:          opts.Role,
		MessageInput:  opts.MessageInput,
		MessageOutput: option.Ptr[string]{},
	})
	for _, entry := range resp.Outputs {
		msgs = append(msgs, entry.Value.IntoMessage())
	}

	return msgs, nil
}

func (p *mistralMessages) Identity() conversation.ID {
	return p.id
}

func (p *mistralMessages) All(subject auth.Subject) iter.Seq2[message.Message, error] {
	return func(yield func(message.Message, error) bool) {
		resp, err := p.client().ListEntries(string(p.id))
		if err != nil {
			yield(message.Message{}, err)
			return
		}

		for _, info := range resp {
			if !yield(info.Value.IntoMessage(), nil) {
				return
			}
		}
	}
}

func (p *mistralMessages) client() *Client {
	return p.parent.client()
}
