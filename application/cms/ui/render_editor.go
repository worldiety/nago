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
	"golang.org/x/text/language"
)

func RenderEditor(doc *core.State[*cms.Document], onUpdated func(elem cms.Element)) core.View {
	if doc.Get() == nil {
		return ui.HStack(ui.Text("Keine Seite ausgewählt.")).FullWidth()
	}

	if doc.Get().Body == nil || len(doc.Get().Body.Elements) == 0 {
		return ui.HStack(ui.Text("Es gibt noch keine Content-Elemente auf der Seite.")).FullWidth()
	}

	return renderElementEditor(doc, doc.Get().Body, onUpdated)
}

func renderElementEditor(doc *core.State[*cms.Document], elem cms.Element, onUpdated func(elem cms.Element)) core.View {
	switch e := elem.(type) {
	case *cms.VStack:
		var tmp []core.View
		for _, child := range e.Elements {
			tmp = append(tmp, renderElementEditor(doc, child, onUpdated))
		}
		return ui.VStack(tmp...).FullWidth()
	case *cms.HStack:
		var tmp []core.View
		for _, child := range e.Elements {
			tmp = append(tmp, renderElementEditor(doc, child, onUpdated))
		}
		return ui.HStack(tmp...).FullWidth()
	case *cms.RichText:
		textState := core.DerivedState[string](doc, string(e.Identity())).Init(func() string {
			return e.Text.String()
		}).Observe(func(newValue string) {
			if e.Text == nil {
				e.Text = map[language.Tag]string{}
			}

			e.Text[doc.Window().Locale()] = newValue
			onUpdated(elem)
		})
		return ui.RichTextEditor(textState.Get()).InputValue(textState).FullWidth()
	default:
		return ui.Text(fmt.Sprintf("%T not implemented", e))
	}
}
