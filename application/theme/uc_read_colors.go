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

var DefaultBaseColors = BaseColors{
	Main:        "#1B8C30",
	Interactive: "#F7A823",
	Accent:      "#03613D",
}

func NewReadColors(loadGlobal settings.LoadGlobal) ReadColors {
	return func(subject auth.Subject) (Colors, error) {
		if err := subject.Audit(PermReadColors); err != nil {
			return Colors{}, err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)

		if !cfg.Colors.Dark.Valid() {
			cfg.Colors.Dark = DarkMode(DefaultBaseColors)
		}

		if !cfg.Colors.Light.Valid() {
			cfg.Colors.Light = LightMode(DefaultBaseColors)
		}

		return cfg.Colors, nil
	}
}
