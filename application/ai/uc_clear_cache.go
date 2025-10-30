// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ai

import (
	"fmt"
	"log/slog"

	"go.wdy.de/nago/application/ai/provider"
	"go.wdy.de/nago/application/ai/provider/cache"
	"go.wdy.de/nago/application/ai/provider/echo"
	"go.wdy.de/nago/auth"
)

func NewClearCache(decorator func(provider provider.Provider) (provider.Provider, error)) ClearCache {
	return func(subject auth.Subject) error {
		if err := subject.Audit(PermClearCache); err != nil {
			return err
		}

		tmp, err := decorator(echo.New("echo", "echo"))
		if err != nil {
			return fmt.Errorf("cannot decorate dummy flush provider for clearance: %w", err)
		}

		if c, ok := tmp.(*cache.Provider); ok {
			if err := c.Clear(); err != nil {
				return fmt.Errorf("cannot clear provider cache: %w", err)
			}

			slog.Warn("cleared cache triggered by echo provider", "user", subject.ID())
		}

		return nil
	}
}
