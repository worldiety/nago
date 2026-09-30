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
	"go.wdy.de/nago/application/ai/model"
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

// ReloadProvider instantiates the providers again from the credentials in the vault. This happens
// automatically whenever a secret is created, updated or deleted.
type ReloadProvider func(subject auth.Subject) error

// FindAllModels lists the models of a provider, see [completion.Completions.Models]. It exists only to declare
// [PermFindAllModel].
type FindAllModels func(subject auth.Subject) iter.Seq2[model.Model, error]

type UseCases struct {
	FindProviderByName FindProviderByName
	FindAllProvider    FindAllProvider
	FindProviderByID   FindProviderByID
	ReloadProvider     ReloadProvider
}

func NewUseCases(bus events.Bus, findSecrets secret.FindGroupSecrets, decorator func(provider provider.Provider) (provider.Provider, error)) UseCases {
	var providers concurrent.RWMap[provider.ID, provider.Provider]
	fnReload := NewReloadProvider(&providers, findSecrets, decorator)

	fnInvokeReload := func() {
		if err := fnReload(user.SU()); err != nil {
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
	}
}
