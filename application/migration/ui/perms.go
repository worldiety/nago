// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uimigration

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"golang.org/x/text/language"
)

// Any permission here must be checked in the view layer, because this is part of the bootstrapping phase
// and the user and auth package depends on it. It does not make sense to depend on it.
// Important: this is not idiomatic and only needed for bootstrapping reasons here. Do not copy this pattern.
var (
	PermViewMigration = permission.Declare[func()](
		"nago.migration.view",
		i18n.MustString(
			"nago.migration.view.title",
			i18n.Values{
				language.English: "View data migrations",
				language.German:  "Datenmigrationen ansehen",
			},
		).String(),
		i18n.MustString(
			"nago.migration.view.desc",
			i18n.Values{
				language.English: "User with this permission can inspect outstanding or applied system data migrations.",
				language.German:  "Benutzer mit dieser Berechtigung können unerledigte oder angewendete Systemdatenmigrationen überprüfen.",
			},
		).String(),
	)

	PermViewReApply = permission.Declare[func()](
		"nago.migration.reapply",
		i18n.MustString(
			"nago.migration.reapply.title",
			i18n.Values{
				language.English: "Re-apply data migrations",
				language.German:  "Datenmigrationen erneut anwenden",
			},
		).String(),
		i18n.MustString(
			"nago.migration.reapply.desc",
			i18n.Values{
				language.English: "User with this permission can re-apply system data migrations.",
				language.German:  "Benutzer mit dieser Berechtigung können unerledigte oder angewendete Systemdatenmigrationen erneut anwenden.",
			},
		).String(),
	)
)
