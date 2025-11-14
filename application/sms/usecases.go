// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package sms

import (
	"context"
	"iter"
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/sms/message"
	"go.wdy.de/nago/application/sms/provider"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
	"go.wdy.de/nago/pkg/xsync"
)

type ReloadProviderOptions struct {
}
type ReloadProvider func(subject auth.Subject, opts ReloadProviderOptions) error

type SendOptions struct {
	NoQueue bool
}
type Send func(subject auth.Subject, sms message.SendRequested, opts SendOptions) (message.ID, error)

type FindAllMessageIDs func(subject auth.Subject) iter.Seq2[message.ID, error]

type FindMessageByID func(subject auth.Subject, id message.ID) (option.Opt[message.SMS], error)

type DeleteMessageByID func(subject auth.Subject, id message.ID) error

type UseCases struct {
	Send              Send
	ReloadProvider    ReloadProvider
	FindAllMessageIDs FindAllMessageIDs
	FindMessageByID   FindMessageByID
	DeleteMessageByID DeleteMessageByID
}

func NewUseCases(ctx context.Context, bus events.Bus, findSecrets secret.FindGroupSecrets, repo message.Repository) UseCases {
	var providers concurrent.RWMap[provider.ID, provider.Provider]
	fnReload := NewReloadProvider(&providers, findSecrets)

	fnInvokeReload := func() {
		if err := fnReload(user.SU(), ReloadProviderOptions{}); err != nil {
			slog.Error("failed to reload sms providers", "err", err.Error())
		}
	}

	fnInvokeReload()

	sendFn := NewSend(repo, &providers)

	xsync.GoFn(func() {
		loop(ctx, repo, sendFn)
	})

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
		if _, err := sendFn(user.SU(), evt, SendOptions{}); err != nil {
			slog.Error("failed to send sms", "err", err.Error())
			return
		}

	})

	return UseCases{
		ReloadProvider:    fnReload,
		Send:              sendFn,
		FindAllMessageIDs: NewFindAllMessageIDs(repo),
		FindMessageByID:   NewFindMessageByID(repo),
		DeleteMessageByID: NewDeleteMessageByID(repo),
	}
}
