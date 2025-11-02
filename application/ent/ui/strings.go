// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uient

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrElements        = i18n.MustString("nago.ent.elements", i18n.Values{language.English: "Elements", language.German: "Elemente"})
	StrDataManagement  = i18n.MustString("nago.ent.datamanagement", i18n.Values{language.English: "Data Management", language.German: "Datenverwaltung"})
	StrManageEntitiesX = i18n.MustVarString("nago.ent.manage_x", i18n.Values{language.English: "Manage items of type {name}. Create, search, update, or delete {name} items.", language.German: "Verwalte Elemente vom Typ {name}. Erstelle, durchsuche, aktualisiere oder lösche {name}-Elemente."})
)
