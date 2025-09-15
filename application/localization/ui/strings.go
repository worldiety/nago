// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uilocalization

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrTranslations       = i18n.MustString("nago.localization.translations", i18n.Values{language.English: "Translations", language.German: "Übersetzungen"})
	StrTranslationSecText = i18n.MustVarString("nago.localization.translations_text", i18n.Values{language.English: "{totalAmount} total strings\n{missingAmount} not translated", language.German: "{totalAmount} Texte insgesamt\n{missingAmount} nicht übersetzt"}, i18n.LocalizationHint("A translation directory has some basic statistic in the admin center. This show how many missing and total strings are available."), i18n.LocalizationVarHint("missingAmount", "An integer variable the count of not translated strings in the bundle"), i18n.LocalizationVarHint("totalAmount", "The absolute count of all recursively contained strings."))
	StrStringKeysTitle    = i18n.MustString("nago.localization.admin.string_keys.title", i18n.Values{language.English: "String Keys", language.German: "Klartext Schlüssel"})
	StrStringKeysDesc     = i18n.MustString("nago.localization.admin.string_keys.desc", i18n.Values{language.English: "String Keys are also text elements which can be translated, but they have no key and instead are identified by their default translation value.", language.German: "String Keys sind ebenfalls Textelemente, die übersetzt werden können, jedoch haben sie keinen Schlüssel und werden stattdessen anhand ihres Standardübersetzungswerts identifiziert."})
	StrLanguagesTitle     = i18n.MustString("nago.localization.admin.languages.title", i18n.Values{language.English: "Languages", language.German: "Sprachen"})
	StrLanguagesDesc      = i18n.MustString("nago.localization.admin.languages.desc", i18n.Values{language.English: "Configure available languages and locales.", language.German: "Ländereinstellungen und Sprachen einstellen."})
)

var (
	StrNotTranslated             = i18n.MustString("nago.localization.admin.not_translated", i18n.Values{language.English: "This text is not yet translated.", language.German: "Dieser Text ist noch nicht übersetzt."})
	StrPriorities                = i18n.MustString("nago.localization.admin.priorities", i18n.Values{language.English: "Priorities", language.German: "Prioritäten"})
	StrFallback                  = i18n.MustString("nago.localization.admin.fallback", i18n.Values{language.English: "Fallback choice", language.German: "Rückfallauswahl"})
	StrAddLanguage               = i18n.MustString("nago.localization.admin.add_language", i18n.Values{language.English: "Add language", language.German: "Sprache hinzufügen"})
	StrAddLanguageSupportingText = i18n.MustString("nago.localization.admin.add_language_supporting_text", i18n.Values{language.English: "Use a valid BCP 47 or IETF language tag to create a placeholder resource bundle for future translations. Examples: en or en-GB.", language.German: "Verwenden Sie ein gültiges BCP 47- oder IETF-Sprach-Tag, um ein Platzhalter-Ressourcenpaket für zukünftige Übersetzungen zu erstellen. Beispiele: en oder en-GB."})
)
