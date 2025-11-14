// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package spryng

import (
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/application/sms/provider"
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

func (p *Provider) Send(subject auth.Subject, sms message.SendRequested) (message.ID, error) {
	res, err := p.cl.Send(sms)
	if err != nil {
		return "", err
	}

	return message.ID(res.Id), nil
}

func (p *Provider) Name() string {
	return p.settings.Name
}

func (p *Provider) Identity() provider.ID {
	return p.id
}
