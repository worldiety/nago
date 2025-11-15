// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mattermost

import (
	"errors"
	"iter"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/application/chatbot/provider"
	"go.wdy.de/nago/application/chatbot/user"
	"go.wdy.de/nago/auth"
)

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	id       provider.ID
	settings Settings
	cl       *Client
}

func NewProvider(id provider.ID, settings Settings) *Provider {
	return &Provider{cl: NewClient(settings), settings: settings, id: id}
}

func (p *Provider) Create(subject auth.Subject, users ...user.ID) (channel.Channel, error) {
	var tmp []string
	for _, id := range users {
		tmp = append(tmp, string(id))
	}

	c, err := p.cl.CreateChannelDirect(tmp...)
	if err != nil {
		return channel.Channel{}, err
	}

	return c.IntoChannel(), nil
}

func (p *Provider) Channel(id channel.ID) provider.Channel {
	return &mattermostChannel{
		parent: p,
		id:     id,
	}
}

func (p *Provider) Me(subject auth.Subject) (user.User, error) {
	res, err := p.cl.UsersMe()
	if err != nil {
		return user.User{}, err
	}

	return res.IntoUser(), nil
}

func (p *Provider) All(subject auth.Subject) iter.Seq2[user.User, error] {
	return func(yield func(user.User, error) bool) {
		users, err := p.cl.Users()
		if err != nil {
			yield(user.User{}, err)
			return
		}

		for _, u := range users {
			if !yield(u.IntoUser(), nil) {
				return
			}
		}
	}
}

func (p *Provider) FindByEmail(subject auth.Subject, mail user.Email) (option.Opt[user.User], error) {
	optUsr, err := p.cl.UserByEmail(string(mail))
	if err != nil {
		return option.None[user.User](), err
	}

	if optUsr.IsNone() {
		return option.None[user.User](), errors.New("user not found")
	}

	return option.Some(optUsr.Unwrap().IntoUser()), nil
}

func (p *Provider) Users() provider.Users {
	return p
}

func (p *Provider) Channels() provider.Channels {
	return p
}

func (p *Provider) Name() string {
	return p.settings.Name
}

func (p *Provider) Identity() provider.ID {
	return p.id
}
