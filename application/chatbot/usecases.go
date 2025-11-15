// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chatbot

import (
	"context"
	"log/slog"

	"go.wdy.de/nago/application/chatbot/message"
	"go.wdy.de/nago/application/chatbot/provider"
	"go.wdy.de/nago/application/secret"
	user2 "go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
)

type ReloadProviderOptions struct {
}
type ReloadProvider func(subject auth.Subject, opts ReloadProviderOptions) error

type SendOptions struct {
}
type Send func(subject auth.Subject, post message.SendRequested, opts SendOptions) (message.ID, error)

type UseCases struct {
	ReloadProvider ReloadProvider
	Send           Send
}

func NewUseCases(ctx context.Context, bus events.Bus, findSecrets secret.FindGroupSecrets) UseCases {
	var providers concurrent.RWMap[provider.ID, provider.Provider]
	fnReload := NewReloadProvider(&providers, findSecrets)

	fnInvokeReload := func() {
		if err := fnReload(user2.SU(), ReloadProviderOptions{}); err != nil {
			slog.Error("failed to reload sms providers", "err", err.Error())
		}
	}

	fnInvokeReload()

	sendFn := NewSend(&providers)

	events.SubscribeFor(bus, func(evt secret.Created) {
		fnInvokeReload()
	})

	events.SubscribeFor(bus, func(evt secret.Updated) {
		fnInvokeReload()
	})

	events.SubscribeFor(bus, func(evt secret.Deleted) {
		fnInvokeReload()
	})

	events.SubscribeFor(bus, func(evt message.SendRequested) {
		if _, err := sendFn(user2.SU(), evt, SendOptions{}); err != nil {
			slog.Error("chatbot failed to send message", "err", err.Error())
			return
		}

	})

	return UseCases{
		ReloadProvider: fnReload,
		Send:           sendFn,
	}
}
