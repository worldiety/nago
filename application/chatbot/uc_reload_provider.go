// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package chatbot

import (
	"log/slog"

	"go.wdy.de/nago/application/chatbot/provider"
	"go.wdy.de/nago/application/chatbot/provider/mattermost"
	"go.wdy.de/nago/application/group"
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/application/user"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewReloadProvider(m *concurrent.RWMap[provider.ID, provider.Provider], findSecrets secret.FindGroupSecrets) ReloadProvider {
	return func(subject auth.Subject, opts ReloadProviderOptions) error {
		if err := subject.Audit(PermReloadProvider); err != nil {
			return err
		}

		m.Clear()

		for sec, err := range findSecrets(user.SU(), group.System) {
			if err != nil {
				slog.Error("failed to load credential", "err", err.Error())
				continue
			}

			var prov provider.Provider
			switch cfg := sec.Credentials.(type) {
			case mattermost.Settings:
				prov = mattermost.NewProvider(provider.ID(sec.ID), cfg)
			}

			if prov == nil {
				continue
			}

			m.Put(prov.Identity(), prov)
		}

		return nil
	}
}
