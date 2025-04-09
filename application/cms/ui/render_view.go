// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicms

import (
	"fmt"
	"go.wdy.de/nago/application/cms"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"os"
	"strings"
)

func RenderPage(wnd core.Window, prefix core.NavigationPath, bySlug cms.FindBySlug) core.View {
	slug := strings.TrimPrefix(string(wnd.Path()), string(prefix))[1:]
	optDoc, err := bySlug(wnd.Subject(), cms.Slug(slug))
	if err != nil {
		return alert.BannerError(err)
	}

	if optDoc.IsNone() {
		return alert.BannerError(os.ErrNotExist)
	}

	doc := optDoc.Unwrap()

	return ui.VStack(
		ui.WindowTitle(doc.Title.Match(wnd.Locale())),
		Render(wnd, doc),
	).Alignment(ui.Leading).FullWidth()
}

func Render(wnd core.Window, doc *cms.Document) core.View {
	if doc == nil {
		return nil
	}

	return renderElement(wnd, doc.Body)
}

func renderElement(wnd core.Window, elem cms.Element) core.View {
	switch e := elem.(type) {
	case *cms.VStack:
		var tmp []core.View
		for _, child := range e.Elements {
			tmp = append(tmp, renderElement(wnd, child))
		}
		return ui.VStack(tmp...).FullWidth()
	case *cms.HStack:
		var tmp []core.View
		for _, child := range e.Elements {
			tmp = append(tmp, renderElement(wnd, child))
		}
		return ui.HStack(tmp...).FullWidth()
	case *cms.RichText:
		return ui.RichText(e.Text.Match(wnd.Locale())).FullWidth()
	default:
		return ui.Text(fmt.Sprintf("%T not implemented", e))
	}
}
