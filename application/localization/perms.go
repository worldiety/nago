// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package localization

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/application/permission"
	"golang.org/x/text/language"
)

var PermReadDir = permission.Declare[ReadDir](
	"nago.localization.readdir",
	i18n.MustString(
		"nago.permissions.localization.readdir_title",
		i18n.Values{
			language.English: "Find localizable sections",
			language.German:  "Übersetzbare Sektionen anzeigen",
		},
	).String(),
	i18n.MustString(
		"nago.permissions.localization.readdir_desc",
		i18n.Values{
			language.English: "Holders of this authorisation can display localizable sections.",
			language.German:  "Träger dieser Berechtigung können lokalisierbare Sektionen anzeigen.",
		},
	).String(),
)

var PermUpdateMessage = permission.Declare[ReadDir](
	"nago.localization.updatemessage",
	i18n.MustString(
		"nago.permissions.localization.updatemessage_title",
		i18n.Values{
			language.English: "Update localized text",
			language.German:  "Übersetzbaren Text aktualisieren",
		},
	).String(),
	i18n.MustString(
		"nago.permissions.localization.updatemessage_desc",
		i18n.Values{
			language.English: "Holders of this authorisation can update localizable text.",
			language.German:  "Träger dieser Berechtigung können übersetzbare Texte aktualisieren.",
		},
	).String(),
)

var PermAddLanguage = permission.Declare[ReadDir](
	"nago.localization.addlanguage",
	i18n.MustString(
		"nago.permissions.localization.addlanguage_title",
		i18n.Values{
			language.English: "Add language",
			language.German:  "Sprache hinzufügen",
		},
	).String(),
	i18n.MustString(
		"nago.permissions.localization.addlanguage_desc",
		i18n.Values{
			language.English: "Holders of this authorization can add additional languages as a translation target.",
			language.German:  "Träger dieser Berechtigung können weitere Sprachen als Übersetzungsziel hinzufügen.",
		},
	).String(),
)
