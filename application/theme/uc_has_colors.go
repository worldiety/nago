// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import (
	"go.wdy.de/nago/application/settings"
	"go.wdy.de/nago/auth"
)

func NewHasColors(loadGlobal settings.LoadGlobal) HasColors {
	return func(subject auth.Subject) (bool, error) {
		if err := subject.Audit(PermReadColors); err != nil {
			return false, err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)

		return cfg.Colors.Dark.Valid() && cfg.Colors.Light.Valid(), nil
	}
}
