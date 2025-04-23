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
	"go.wdy.de/nago/presentation/core"
)

func NewReadFonts(loadGlobal settings.LoadGlobal) ReadFonts {
	return func(subject auth.Subject) (core.Fonts, error) {
		if err := subject.Audit(PermReadColors); err != nil {
			return core.Fonts{}, err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)
		return cfg.Fonts, nil
	}
}
