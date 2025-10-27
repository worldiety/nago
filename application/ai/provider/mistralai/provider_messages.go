// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"

	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Messages = (*mistralMessages)(nil)

type mistralMessages struct {
	id     conversation.ID
	parent *mistralProvider
}

func (p *mistralMessages) Conversation() conversation.ID {
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
