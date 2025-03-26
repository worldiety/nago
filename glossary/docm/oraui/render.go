// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oraui

import (
	"fmt"
	"go.wdy.de/nago/glossary/docm"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

func Render(doc *docm.Document) core.View {
	return renderElem(doc.Body)
}

func renderElem(elem docm.Element) core.View {
	switch elem := elem.(type) {
	case docm.Sequence:
		if len(elem) == 0 {
			return nil
		}

		var views []core.View
		for _, element := range elem {
			views = append(views, renderElem(element))
		}

		return ui.VStack(views...).Alignment(ui.Leading).Gap(ui.L8)
	case *docm.Text:
		return ui.Text(elem.Value)
	case *docm.Heading:
		return renderHeading(elem)
	case *docm.List:
		return renderList(elem)
	case *docm.Pre:
		return renderPre(elem)
	case *docm.Par:
		return renderPar(elem)
	case *docm.Image:
		return renderImg(elem)
	default:
		return ui.Text(fmt.Sprintf("unsupported docm type: %T", elem))
	}
}

func renderImg(elem *docm.Image) core.View {
	return ui.Image().URI(core.URI(elem.URL)).Frame(ui.Frame{}.FullWidth())
}

func renderPar(elem *docm.Par) core.View {
	var views []core.View
	for _, element := range elem.Children {
		views = append(views, renderElem(element))
	}

	return ui.TextLayout(views...)
}

func renderPre(elem *docm.Pre) core.View {
	var views []core.View
	for idx, line := range elem.Lines {
		views = append(views, ui.Text(fmt.Sprintf("%3d %s", idx+1, line)).Font(ui.Font{Name: "monospace"}))
	}
	return ui.VStack(views...).Alignment(ui.Leading)
}

func renderList(elem *docm.List) core.View {
	var views []core.View
	for _, child := range elem.Children {
		views = append(views, ui.HStack(
			ui.Text("•").Padding(ui.Padding{Right: ui.L8}),
			renderElem(child),
		))
	}
	return ui.VStack(views...).Alignment(ui.Leading)
}

func renderHeading(elem *docm.Heading) core.View {
	font := ui.Title
	switch elem.Level {
	case 2:
		font = ui.SubTitle
	case 3:
		font = ui.Font{
			Size:   "1rem",
			Weight: ui.BoldFontWeight,
			Style:  ui.ItalicFontStyle,
		}
	case 4:
		font = ui.Font{
			Size:  "1rem",
			Style: ui.ItalicFontStyle,
		}

	case 5:
		font = ui.Font{
			Size:  "0.5rem",
			Style: ui.ItalicFontStyle,
		}
	}
	return ui.VStack(ui.TextLayout(renderElem(elem.Body)).Font(font).Padding(ui.Padding{Top: ui.L12, Bottom: ui.L8}))
}
