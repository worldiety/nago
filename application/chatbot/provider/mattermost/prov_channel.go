// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package mattermost

import (
	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/application/chatbot/message"
	"go.wdy.de/nago/application/chatbot/provider"
	"go.wdy.de/nago/auth"
)

var _ provider.Channel = (*mattermostChannel)(nil)

type mattermostChannel struct {
	parent *Provider
	id     channel.ID
}

func (m *mattermostChannel) Post(subject auth.Subject, opts message.CreateOptions) (message.Message, error) {
	resp, err := m.parent.cl.Post(CreatePostRequest{
		ChannelId: string(m.id),
		Message:   opts.Message,
	})

	if err != nil {
		return message.Message{}, err
	}

	return message.Message{
		ID:      message.ID(resp.Id),
		Channel: m.id,
		Message: opts.Message,
	}, nil
}
