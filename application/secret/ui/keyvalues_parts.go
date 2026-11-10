// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/list"
)

// The look of the key values follows password managers: quiet rows in grouped cards, a small dimmed label above each
// value and the state of a field within its row, instead of sentences below it.

// concealedText stands for a value, which is not in the browser.
const concealedText = "••••••••"

var (
	concealedFont = ui.Font{Name: ui.MonoFontName, Size: ui.L14, Weight: ui.MonoFontWeight}
	keyFont       = ui.Font{Name: ui.MonoFontName, Size: ui.L12, Weight: ui.MonoFontWeight}
)

// muted dims a view. The muted colors of the theme have too little contrast for small text.
func muted(view core.View) core.View {
	return ui.VStack(view).Alignment(ui.Leading).Opacity(0.7)
}

func smallMuted(text string) core.View {
	return muted(ui.Text(text).Font(ui.BodySmall))
}

// tile is the square for a logo. It is white in any theme, so that a dark logo stays visible.
func tile(content core.View) ui.DecoredView {
	return ui.VStack(content).
		BackgroundColor(ui.ColorWhite).
		Border(ui.Border{}.Radius(ui.L12).Width(ui.L1).Color(ui.ColorLine)).
		Frame(ui.Frame{}.Size(ui.L60, ui.L60))
}

// emptyTile asks for a logo. In contrast to a logo, its icon follows the theme.
func emptyTile() ui.DecoredView {
	return ui.VStack(muted(ui.ImageIcon(heroOutline.Photo).Frame(ui.Frame{}.Size(ui.L24, ui.L24)))).
		BackgroundColor(ui.ColorCardBody).
		Border(ui.Border{}.Radius(ui.L12).Width(ui.L1).Color(ui.ColorLine)).
		Frame(ui.Frame{}.Size(ui.L60, ui.L60))
}

// sectionBlock places the heading above the card, as the card holds nothing but the fields.
func sectionBlock(heading core.View, rows []core.View) core.View {
	return ui.VStack(
		heading,
		list.List(rows...).ColorHover(ui.ColorCardBody).FullWidth(),
	).Alignment(ui.Leading).Gap(ui.L8).FullWidth()
}

func sectionHeading(section secret.KeyValueSection, trailing core.View) core.View {
	return ui.HStack(
		ui.VStack(
			ui.If(section.Label != "", ui.Text(section.Label).Font(ui.TitleMedium)),
			ui.If(section.SupportingText != "", smallMuted(section.SupportingText)),
		).Alignment(ui.Leading),
		ui.Spacer(),
		trailing,
	).FullWidth()
}

// fieldUse tells in a line who consumes the value, so that no field needs a sentence about it.
func fieldUse(field secret.KeyValueField) core.View {
	icon, text, key := heroOutline.CommandLine, "Programme lesen ·", field.Key
	switch name, isHeader := secret.HeaderForKey(field.Key); {
	case field.Key == "":
		icon, text, key = heroOutline.ExclamationTriangle, "Schlüssel fehlt", ""
	case field.Key == secret.KeyBaseURL:
		icon, text, key = heroOutline.Link, "Verbindung · Adresse", ""
	case field.Key == secret.KeyHeaderBearerAuth:
		icon, text, key = heroOutline.Link, "Verbindung · Bearer-Token", ""
	case field.Key == secret.KeyHeaderBasicAuthUser, field.Key == secret.KeyHeaderBasicAuthPassword:
		icon, text, key = heroOutline.Link, "Verbindung · Basic-Auth", ""
	case isHeader:
		icon, text, key = heroOutline.Link, "Verbindung · Header", name
	case secret.IsFixedKey(field.Key):
		if name, isQuery := secret.QueryForKey(field.Key); isQuery {
			icon, text, key = heroOutline.Link, "Verbindung · Query", name
		} else {
			icon, text, key = heroOutline.ExclamationTriangle, "Keine gültige Verbindungsangabe ·", field.Key
		}
	case field.Protection == secret.WriteOnly:
		icon, text = heroOutline.LockClosed, "Für Programme gesperrt ·"
	}

	return muted(ui.HStack(
		ui.ImageIcon(icon).Frame(ui.Frame{}.Size(ui.L14, ui.L14)),
		ui.Text(text).Font(ui.BodySmall),
		ui.If(key != "", ui.Text(key).Font(keyFont)),
	).Gap(ui.L4).Wrap(true))
}

func fieldTitle(field secret.KeyValueField) string {
	switch {
	case field.Label != "":
		return field.Label
	case field.Key != "":
		return field.Key
	default:
		return "Neues Feld"
	}
}

// logoImage renders an SVG, which the frontend sanitizes before it becomes a part of the page, see [ui.TImage.Embed].
func logoImage(svg string, size ui.Length) core.View {
	return ui.Image().Embed([]byte(svg)).ObjectFit(ui.FitContain).Frame(ui.Frame{}.Size(size, size))
}
