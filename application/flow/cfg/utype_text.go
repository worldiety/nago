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
	StrTextLabel       = i18n.MustString("nago.flow.utype.text.label", i18n.Values{language.English: "Text", language.German: "Text"})
	StrTextDescription = i18n.MustString("nago.flow.utype.text.desc", i18n.Values{language.English: "Text represents the underlying type for single line text fields.", language.German: "Text stellt den unterliegenden Typ für einzeilige Textfelder dar."})
)
