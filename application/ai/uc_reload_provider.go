// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"log/slog"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewReloadProvider(m *concurrent.RWMap[provider.ID, provider.Provider], findSecrets secret.FindGroupSecrets, decorator func(provider provider.Provider) (provider.Provider, error), services ...provider.Provider) ReloadProvider {
	return func(subject auth.Subject) error {
		if err := subject.Audit(PermReloadProvider); err != nil {
			return err
		}

		m.Clear()

		// services provision themselves and are there regardless of the vault
		for _, prov := range services {
			if decorator != nil {
				decorated, err := decorator(prov)
				if err != nil {
					slog.Error("failed to decorate provider", "provider", prov.Identity(), "err", err.Error())
					continue
				}

				prov = decorated
			}

			m.Put(prov.Identity(), prov)
		}

		for sec, err := range findSecrets(user.SU(), group.System) {
			if err != nil {
				slog.Error("failed to load credential", "err", err.Error())
				continue
			}

			// The concrete provider is resolved through the global registry. A provider is only registered
			// (and thus compiled in) when the host application side-imports its package, e.g.
			// _ "go.wdy.de/nago/application/ai/provider/anthropic". Unknown credentials types are skipped.
			prov, ok := provider.NewProviderFor(provider.ID(sec.ID), sec.Credentials)
			if !ok {
				continue
			}

			if decorator != nil {
				prov, err = decorator(prov)
				if err != nil {
					slog.Error("failed to decorate provider", "provider", prov.Identity(), "err", err.Error())
					continue
				}
			}

			m.Put(prov.Identity(), prov)
		}

		return nil
	}
}
