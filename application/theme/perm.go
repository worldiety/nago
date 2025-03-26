// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package theme

import "go.wdy.de/nago/application/permission"

var (
	PermUpdateColors = permission.Declare[UpdateColors]("nago.theme.colors.update", "Themefarben aktualisieren", "Träger dieser Berechtigung können die Themefarben aktualisieren")
	PermReadColors   = permission.Declare[UpdateColors]("nago.theme.colors.read", "Themefarben auslesen", "Träger dieser Berechtigung können die Themefarben auslesen")
)
