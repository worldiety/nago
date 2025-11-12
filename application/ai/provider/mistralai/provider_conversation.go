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

var _ provider.Conversation = (*mistralMessages)(nil)

type mistralMessages struct {
	id     conversation.ID
	parent *mistralProvider
}

// Append is broken for this implementation, because the Mistral API does not include any input message, even
// though they were processed and got already identifiers etc. We currently just assign a fake identifier
// and return the input messages, however that is actually not what someone would expect.
func (p *mistralMessages) Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error) {
	tmp := convInputToMistralInput(p.client(), opts.Input).Values

	_, err := p.client().AppendConversation(string(p.id), AppendConversationRequest{
		Inputs: tmp,
		Store:  opts.CloudStore,
		Stream: false,
	})

	if err != nil {
		return nil, err
	}

	// the mistral API is not symmetric and input/output messages are not in-sync regarding the actual
	// conversation history, therefore let us not pretend otherwise and instead tell the callee
	// that we are broken by returning nil

	return nil, nil
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
			for _, value := range info.Values {
				for _, m := range value.IntoMessages() {
					if !yield(m, nil) {
						return
					}
				}

			}

		}
	}
}

func (p *mistralMessages) client() *Client {
	return p.parent.client()
}
