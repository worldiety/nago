// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"github.com/worldiety/option"
	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/std/concurrent"
)

func NewFindProviderByName(m *concurrent.RWMap[provider.ID, provider.Provider]) FindProviderByName {
	return func(subject auth.Subject, name string) (option.Opt[provider.Provider], error) {
		if err := subject.Audit(PermFindProviderByName); err != nil {
			return option.Opt[provider.Provider]{}, err
		}

		for _, p := range m.All() {
			if p.Name() == name {
				return option.Some(p), nil
			}
		}

		return option.None[provider.Provider](), nil
	}
}
