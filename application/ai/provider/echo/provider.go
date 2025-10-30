// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package echo

import (
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/conversation"
	"go.wdy.de/nago/application/ai/message"
	"go.wdy.de/nago/application/ai/model"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/xtime"
)

type Provider struct {
	id   provider.ID
	name string
}

func New(id provider.ID, name string) *Provider {
	return &Provider{}
}

func (p *Provider) Identity() provider.ID {
	return p.id
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) Description() string {
	return ""
}

func (p *Provider) Models() provider.Models {
	return models{}
}

func (p *Provider) Libraries() option.Opt[provider.Libraries] {
	return option.None[provider.Libraries]()
}

func (p *Provider) Agents() option.Opt[provider.Agents] {
	return option.None[provider.Agents]()
}

func (p *Provider) Conversations() option.Opt[provider.Conversations] {
	return option.Some[provider.Conversations](conversations{})
}

type conversations struct {
}

func (c conversations) All(subject auth.Subject) iter.Seq2[conversation.Conversation, error] {
	return func(yield func(conversation.Conversation, error) bool) {

	}
}

func (c conversations) FindByID(subject auth.Subject, id conversation.ID) (option.Opt[conversation.Conversation], error) {
	return option.None[conversation.Conversation](), nil
}

func (c conversations) Delete(subject auth.Subject, id conversation.ID) error {
	return nil
}

func (c conversations) Create(subject auth.Subject, opts conversation.CreateOptions) (conversation.Conversation, []message.Message, error) {
	var res []message.Message
	for _, m := range opts.Input {

		res = append(res, message.Message{
			ID:           data.RandIdent[message.ID](),
			CreatedAt:    xtime.Now(),
			CreatedBy:    subject.ID(),
			Role:         message.User,
			MessageInput: m.Text,
		})
	}

	for _, m := range opts.Input {
		tmp := "hello echo: "
		if m.Text.IsSome() {
			tmp += m.Text.Unwrap()
		}
		res = append(res, message.Message{
			ID:            data.RandIdent[message.ID](),
			CreatedAt:     xtime.Now(),
			CreatedBy:     subject.ID(),
			Role:          message.AssistantRole,
			MessageOutput: option.Pointer(&tmp),
		})
	}
	return conversation.Conversation{
			ID:           data.RandIdent[conversation.ID](),
			Agent:        opts.Agent,
			Model:        opts.Model,
			Name:         opts.Name,
			Description:  opts.Description,
			Instructions: opts.Instructions,
			CreatedAt:    xtime.Now(),
			CreatedBy:    subject.ID(),
		},
		res, nil
}

func (c conversations) Conversation(subject auth.Subject, id conversation.ID) provider.Conversation {
	return echoConv{id: id}
}

type echoConv struct {
	id conversation.ID
}

func (e echoConv) Identity() conversation.ID {
	return e.id
}

func (e echoConv) All(subject auth.Subject) iter.Seq2[message.Message, error] {
	return func(yield func(message.Message, error) bool) {

	}
}

func (e echoConv) Append(subject auth.Subject, opts message.AppendOptions) ([]message.Message, error) {
	var res []message.Message

	res = append(res, message.Message{
		ID:           data.RandIdent[message.ID](),
		CreatedAt:    xtime.Now(),
		CreatedBy:    subject.ID(),
		Role:         message.User,
		MessageInput: opts.MessageInput,
	})

	tmp := "hello echo: "
	if opts.MessageInput.IsSome() {
		tmp += opts.MessageInput.Unwrap()
	}

	res = append(res, message.Message{
		ID:            data.RandIdent[message.ID](),
		CreatedAt:     xtime.Now(),
		CreatedBy:     subject.ID(),
		Role:          message.AssistantRole,
		MessageOutput: option.Pointer(&tmp),
	})

	return res, nil
}

type models struct {
}

func (m models) All(subject auth.Subject) iter.Seq2[model.Model, error] {
	return func(yield func(model.Model, error) bool) {
		yield(model.Model{
			ID:                 "echo",
			Name:               "Echo",
			Description:        "Parrots what you say",
			DefaultTemperature: 0,
		}, nil)
	}
}
