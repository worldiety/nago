// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package cfgflow

import (
	"github.com/worldiety/i18n"
	"golang.org/x/text/language"
)

var (
	StrCheckboxLabel       = i18n.MustString("nago.flow.utype.checkbox.label", i18n.Values{language.English: "Checkbox", language.German: "Ankreuzfeld"})
	StrCheckboxDescription = i18n.MustString("nago.flow.utype.checkbox.desc", i18n.Values{language.English: "Checkbox represents the underlying type for single line text fields.", language.German: "Ankreuzfelder stellen den unterliegenden Typ für Ankreuzoptionen dar."})
)
