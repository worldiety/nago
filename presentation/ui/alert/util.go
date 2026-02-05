// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package alert

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"golang.org/x/text/language"
)

var (
	StrNotFound    = i18n.MustString("nago.alert.not_found", i18n.Values{language.German: "Nicht gefunden", language.English: "Not found"})
	StrNotFoundMsg = i18n.MustString("nago.alert.not_found_msg", i18n.Values{language.German: "Die Seite, Funktion oder Resource wurde nicht gefunden.", language.English: "The page, function or resource was not found."})
)

func NotFound() core.View {
	return core.RenderView(func(wnd core.Window) core.View {
		return ui.VStack(
			ui.WindowTitle(StrNotFound.Get(wnd)),
			Banner(StrNotFound.Get(wnd), StrNotFoundMsg.Get(wnd)),
		)
	})
}
