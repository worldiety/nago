// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"iter"
	"log/slog"

	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/events"
	"go.wdy.de/nago/pkg/std/concurrent"
)

type FindProviderByName func(subject auth.Subject, name string) (option.Opt[provider.Provider], error)

// FindAllProvider returns the known providers sorted asc by name.
type FindAllProvider func(subject auth.Subject) iter.Seq2[provider.Provider, error]

type FindProviderByID func(subject auth.Subject, id provider.ID) (option.Opt[provider.Provider], error)

type ReloadProviderOptions struct {
	// LoadAll indicates if the entire provider should be scraped and stored into the cache (if configured).
	LoadAll bool
}
type ReloadProvider func(subject auth.Subject, opts ReloadProviderOptions) error

type ClearCache func(subject auth.Subject) error

type UseCases struct {
	FindProviderByName FindProviderByName
	FindAllProvider    FindAllProvider
	FindProviderByID   FindProviderByID
	ReloadProvider     ReloadProvider
	ClearCache         ClearCache
}

func NewUseCases(bus events.Bus, findSecrets secret.FindGroupSecrets, decorator func(provider provider.Provider) (provider.Provider, error)) UseCases {
	var providers concurrent.RWMap[provider.ID, provider.Provider]
	fnReload := NewReloadProvider(&providers, findSecrets, decorator)

	fnInvokeReload := func() {
		if err := fnReload(user.SU(), ReloadProviderOptions{}); err != nil {
			slog.Error("failed to reload providers", "err", err.Error())
		}
	}

	fnInvokeReload()

	events.SubscribeFor(bus, func(evt secret.Created) {
		fnInvokeReload()
	})

	events.SubscribeFor(bus, func(evt secret.Updated) {
		fnInvokeReload()
	})

	events.SubscribeFor(bus, func(evt secret.Deleted) {
		fnInvokeReload()
	})

	return UseCases{
		ReloadProvider:     fnReload,
		FindProviderByName: NewFindProviderByName(&providers),
		FindAllProvider:    NewFindAllProvider(&providers),
		FindProviderByID:   NewFindProviderByID(&providers),
		ClearCache:         NewClearCache(decorator),
	}
}
