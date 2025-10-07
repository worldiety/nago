// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package rstring

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	ActionOpen = i18n.MustString(
		"nago.common.action.open",
		i18n.Values{
			language.English: "Open",
			language.German:  "Öffnen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionSaveAndNext = i18n.MustString(
		"nago.common.action.save_and_next",
		i18n.Values{
			language.English: "Save and next",
			language.German:  "Speichern und weiter",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionPrevious = i18n.MustString(
		"nago.common.action.previous",
		i18n.Values{
			language.English: "Previous",
			language.German:  "Vorheriger",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionNext = i18n.MustString(
		"nago.common.action.next",
		i18n.Values{
			language.English: "Next",
			language.German:  "Nächster",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionSelect = i18n.MustString(
		"nago.common.action.select",
		i18n.Values{
			language.English: "Select",
			language.German:  "Auswählen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionAdd = i18n.MustString(
		"nago.common.action.add",
		i18n.Values{
			language.English: "Add",
			language.German:  "Hinzufügen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionBack = i18n.MustString(
		"nago.common.action.back",
		i18n.Values{
			language.English: "Back",
			language.German:  "Zurück",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionCancel = i18n.MustString(
		"nago.common.action.cancel",
		i18n.Values{
			language.English: "Cancel",
			language.German:  "Abbrechen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionSave = i18n.MustString(
		"nago.common.action.save",
		i18n.Values{
			language.English: "Save",
			language.German:  "Speichern",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionCreate = i18n.MustString(
		"nago.common.action.create",
		i18n.Values{
			language.English: "Create",
			language.German:  "Erstellen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionEdit = i18n.MustString(
		"nago.common.action.edit",
		i18n.Values{
			language.English: "Edit",
			language.German:  "Bearbeiten",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionLogin = i18n.MustString(
		"nago.common.action.login",
		i18n.Values{
			language.English: "Login",
			language.German:  "Anmelden",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionApply = i18n.MustString(
		"nago.common.action.apply",
		i18n.Values{
			language.English: "Apply",
			language.German:  "Übernehmen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionClose = i18n.MustString(
		"nago.common.action.close",
		i18n.Values{
			language.English: "Close",
			language.German:  "Schließen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionConfirm = i18n.MustString(
		"nago.common.action.confirm",
		i18n.Values{
			language.English: "Confirm",
			language.German:  "Bestätigen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionDelete = i18n.MustString(
		"nago.common.action.delete",
		i18n.Values{
			language.English: "Delete",
			language.German:  "Löschen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionDownload = i18n.MustString(
		"nago.common.action.download",
		i18n.Values{
			language.English: "Download",
			language.German:  "Herunterladen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	ActionRename = i18n.MustString(
		"nago.common.action.rename",
		i18n.Values{
			language.English: "Rename",
			language.German:  "Umbenennen",
		},
		i18n.LocalizationHint("This text is usually used on buttons and is displayed where space must be minified. So keep it as short and generic as possible."),
	)
)

var (
	LabelLanguage = i18n.MustString(
		"nago.common.label.language",
		i18n.Values{
			language.English: "Language",
			language.German:  "Sprache",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelNothingSelected = i18n.MustString(
		"nago.common.label.nothing_selected",
		i18n.Values{
			language.English: "Nothing selected",
			language.German:  "Nichts gewählt",
		},
		i18n.LocalizationHint("This text is usually used on pickers where space must be minified. So keep it as short and generic as possible."),
	)

	LabelPleaseWait = i18n.MustString(
		"nago.common.label.please_wait",
		i18n.Values{
			language.English: "Please wait...",
			language.German:  "Einen Moment bitte...",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelError = i18n.MustString(
		"nago.common.label.error",
		i18n.Values{
			language.English: "Error",
			language.German:  "Fehler",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelOptions = i18n.MustString(
		"nago.common.label.options",
		i18n.Values{
			language.English: "Options",
			language.German:  "Optionen",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelName = i18n.MustString(
		"nago.common.label.name",
		i18n.Values{
			language.English: "Name",
			language.German:  "Name",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelChanged = i18n.MustString(
		"nago.common.label.changed",
		i18n.Values{
			language.English: "Modified",
			language.German:  "Geändert",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelChangedBy = i18n.MustString(
		"nago.common.label.changed_by",
		i18n.Values{
			language.English: "Modified by",
			language.German:  "Geändert von",
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)

	LabelXItems = i18n.MustQuantityString(
		"nago.common.label.x_items",

		i18n.QValues{
			language.English: i18n.Quantities{
				One:   "{x} item",
				Other: "{x} items",
			},
			language.German: i18n.Quantities{
				One:   "{x} Element",
				Other: "{x} Elemente",
			},
		},
		i18n.LocalizationHint("This text is usually used on buttons or pickers is displayed where space must be minified. So keep it as short and generic as possible."),
	)
)
