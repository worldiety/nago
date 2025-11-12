// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mistralai

import (
	"iter"
	"log/slog"

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

	resp, err := p.client().AppendConversation(string(p.id), AppendConversationRequest{
		Inputs: tmp,
		Store:  opts.CloudStore,
		Stream: false,
	})

	if err != nil {
		return nil, err
	}

	// TODO will request all messages again from the API but that may increase costs and certainly increase latency.
	// TODO @Mistral Team please just go fix your API
	_ = resp // ignore the incomplete and partial result and instead do...
	if len(resp.Outputs) > 1 {
		slog.Warn("@Torben: check if mistral fixed their API")
	}
	list, err := p.client().ListEntries(string(p.id))
	if err != nil {
		return nil, err
	}

	var msgs []message.Message

	if opts.Role == "" {
		opts.Role = "user" // this API design makes me crazy, nothing is explained
	}

	for _, entry := range list {
		for _, value := range entry.Values {
			msgs = append(msgs, value.IntoMessages()...)
		}

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
