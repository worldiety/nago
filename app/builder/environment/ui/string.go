// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uienv

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrEnvironment         = i18n.MustString("nbuilder.environment.title", i18n.Values{language.English: "Environment", language.German: "Umgebung"})
	StrCreateEnvironment   = i18n.MustString("nbuilder.environment.create", i18n.Values{language.English: "create Environment", language.German: "Neue Umgebung erstellen"})
	StrAppsAndEnvironments = i18n.MustString("nbuilder.environment.apps_and_env", i18n.Values{language.English: "Apps & Environments", language.German: "Umgebungen und Apps"})

	StrCreateApp = i18n.MustString("nbuilder.environment.create_app", i18n.Values{language.English: "create app", language.German: "Neue App erstellen"})

	StrNamespaces = i18n.MustString("nbuilder.namespace.title", i18n.Values{language.English: "Namespaces", language.German: "Namespaces"})
)
