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

// Defaults returns the colors, which an application proposes, as long as the user has not chosen any. Colors which
// are not valid fall back to [DefaultBaseColors].
type Defaults func() Colors

func NewReadColors(loadGlobal settings.LoadGlobal, defaults Defaults) ReadColors {
	return func(subject auth.Subject) (Colors, error) {
		if err := subject.Audit(PermReadColors); err != nil {
			return Colors{}, err
		}

		cfg := settings.ReadGlobal[Settings](loadGlobal)

		return withDefaults(cfg.Colors, defaults), nil
	}
}

// withDefaults replaces the colors of a mode, which the user has not chosen, by the proposal of the application or
// by the defaults of nago.
func withDefaults(colors Colors, defaults Defaults) Colors {
	var proposed Colors
	if defaults != nil {
		proposed = defaults()
	}

	if !colors.Dark.Valid() {
		colors.Dark = proposed.Dark
		if !colors.Dark.Valid() {
			colors.Dark = DarkMode(DefaultBaseColors)
		}
	}

	if !colors.Light.Valid() {
		colors.Light = proposed.Light
		if !colors.Light.Valid() {
			colors.Light = LightMode(DefaultBaseColors)
		}
	}

	return colors
}
