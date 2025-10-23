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

func NewFindProviderByID(m *concurrent.RWMap[provider.ID, provider.Provider]) FindProviderByID {
	return func(subject auth.Subject, id provider.ID) (option.Opt[provider.Provider], error) {
		if err := subject.Audit(PermFindProviderByID); err != nil {
			return option.Opt[provider.Provider]{}, err
		}

		if v, ok := m.Get(id); ok {
			return option.Some(v), nil
		}

		return option.None[provider.Provider](), nil
	}
}
