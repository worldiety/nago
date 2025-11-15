// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chatbot

import (
	"fmt"
	"os"
	"sync/atomic"

	"go.wdy.de/nago/application/chatbot/channel"
	"go.wdy.de/nago/application/chatbot/message"
	"go.wdy.de/nago/application/chatbot/provider"
	"go.wdy.de/nago/application/chatbot/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewSend(providers *concurrent.RWMap[provider.ID, provider.Provider]) Send {
	var userLookup concurrent.RWMap[user.Email, user.ID]
	var chanLookup concurrent.RWMap[user.ID, channel.ID]
	var me atomic.Pointer[user.User]

	return func(subject auth.Subject, post message.SendRequested, opts SendOptions) (message.ID, error) {
		if err := subject.Audit(PermSend); err != nil {
			return "", err
		}

		var prov provider.Provider
		for id, p := range providers.All() {
			if prov == nil {
				prov = p
				continue
			}

			if provider.ID(post.ProviderHint) == id || prov.Name() == post.ProviderHint {
				prov = p
				break
			}
		}

		if prov == nil {
			return "", fmt.Errorf("no provider configured")
		}

		if post.RecipientByMail == "" && post.RecipientByID == "" {
			return "", fmt.Errorf("no recipients configured")
		}

		if post.RecipientByID == "" {
			if v, ok := userLookup.Get(post.RecipientByMail); ok {
				post.RecipientByID = v
			} else {
				optUsr, err := prov.Users().FindByEmail(subject, post.RecipientByMail)
				if err != nil {
					return "", fmt.Errorf("failed to find chatbot user by mail: %s: %w", post.RecipientByMail, err)
				}
				if optUsr.IsNone() {
					return "", fmt.Errorf("chatbot provider does not know the users email: %s: %w", post.RecipientByMail, os.ErrNotExist)
				}

				usr := optUsr.Unwrap()
				post.RecipientByID = usr.ID
				userLookup.Put(usr.Email, usr.ID)
			}
		}

		if me.Load() == nil {
			usr, err := prov.Users().Me(subject)
			if err != nil {
				return "", fmt.Errorf("failed to find user myself: %w", err)
			}

			me.Store(&usr)
		}

		var chanId channel.ID
		if v, ok := chanLookup.Get(post.RecipientByID); ok {
			chanId = v
		} else {
			ch, err := prov.Channels().Create(subject, post.RecipientByID, me.Load().ID)
			if err != nil {
				return "", fmt.Errorf("failed to create chatbot channel to %s: %w", post.RecipientByID, err)
			}

			chanId = ch.ID
			chanLookup.Put(post.RecipientByID, ch.ID)
		}

		msg, err := prov.Channels().Channel(chanId).Post(subject, message.CreateOptions{Message: post.Text})
		if err != nil {
			return "", fmt.Errorf("chatbot failed to post into channel: %s: %w", chanId, err)
		}

		return msg.ID, nil
	}
}
