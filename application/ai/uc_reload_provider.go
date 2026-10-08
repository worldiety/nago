// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

// built is a provider made from a secret, kept with the credentials it was made of.
type built struct {
	creds secret.Credentials
	prov  provider.Provider
}

// NewReloadProvider builds the providers from the secrets of the vault, puts the services next to them and decorates
// all of them.
//
// A reload is safe for its readers and against itself:
//   - reloads run one after another, so that a slow, older reload never publishes over a newer one,
//   - the providers are built aside and published at once, so that a reader never finds none or half of them,
//   - a provider is only built anew if its credentials changed. Building is not free: a provider may hold caches,
//     rate limiters or loaded models, which a reload would otherwise discard, e.g. for every change of an unrelated
//     secret. Decorating, in contrast, happens on every reload.
//
// A provider whose decoration fails is left out, and the errors are returned, after the others have been published.
func NewReloadProvider(m *concurrent.RWMap[provider.ID, provider.Provider], findSecrets secret.FindGroupSecrets, decorator func(provider provider.Provider) (provider.Provider, error), services ...provider.Provider) ReloadProvider {
	var mutex sync.Mutex
	cache := map[provider.ID]built{}

	decorate := func(prov provider.Provider) (provider.Provider, error) {
		if decorator == nil {
			return prov, nil
		}

		return checked(prov, decorator)
	}

	return func(subject auth.Subject) error {
		if err := subject.Audit(PermReloadProvider); err != nil {
			return err
		}

		mutex.Lock()
		defer mutex.Unlock()

		next := map[provider.ID]provider.Provider{}
		var errs []error
		publish := func(prov provider.Provider) {
			decorated, err := decorate(prov)
			if err != nil {
				slog.Error("failed to decorate provider", "provider", prov.Identity(), "err", err.Error())
				errs = append(errs, fmt.Errorf("provider %s: %w", prov.Identity(), err))
				return
			}

			next[decorated.Identity()] = decorated
		}

		// services provision themselves and are there regardless of the vault
		for _, prov := range services {
			publish(prov)
		}

		seen := map[provider.ID]bool{}
		for sec, err := range findSecrets(user.SU(), group.System) {
			if err != nil {
				slog.Error("failed to load credential", "err", err.Error())
				continue
			}

			id := provider.ID(sec.ID)
			entry, ok := cache[id]
			if !ok || !reflect.DeepEqual(entry.creds, sec.Credentials) {
				// The concrete provider is resolved through the global registry. A provider is only registered
				// (and thus compiled in) when the host application side-imports its package, e.g.
				// _ "go.wdy.de/nago/application/ai/provider/anthropic". Unknown credentials types are skipped.
				prov, ok := provider.NewProviderFor(id, sec.Credentials)
				if !ok {
					continue
				}

				entry = built{creds: sec.Credentials, prov: prov}
				cache[id] = entry
			}

			seen[id] = true
			publish(entry.prov)
		}

		for id := range cache {
			if !seen[id] {
				delete(cache, id)
			}
		}

		m.Replace(next)
		return errors.Join(errs...)
	}
}
