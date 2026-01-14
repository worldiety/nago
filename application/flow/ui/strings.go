// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uiflow

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrGroupTitle       = i18n.MustString("nago.flow.admin.title", i18n.Values{language.German: "Flow Daten & Formulare", language.English: "Flow Data & Forms"})
	StrGroupDescription = i18n.MustString("nago.flow.admin.description", i18n.Values{language.German: "Arbeitsbereiche für dynamische Datenmodellierung und Formulare verwalten.", language.English: "Manage workspaces for dynamic data modeling and forms."})
	StrWorkspaces       = i18n.MustString("nago.flow.admin.workspaces", i18n.Values{language.German: "Arbeitsbereiche", language.English: "Workspaces"})
)
