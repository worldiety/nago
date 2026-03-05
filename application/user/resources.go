// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package user

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrResUsers = i18n.MustString("nago.user.resources.name", i18n.Values{language.German: "Nutzer", language.English: "Users"})
	StrResDesc  = i18n.MustString("nago.user.resources.desc", i18n.Values{language.German: "Authentifizierte und autorisierte Benutzer bzw. Konten im System.", language.English: "Authenticated and authorized users or accounts in the system."})
)
