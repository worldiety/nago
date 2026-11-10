// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uisecret

import (
	"net/url"
	"strconv"

	"go.wdy.de/nago/application/secret"
	"go.wdy.de/nago/presentation/core"
	heroOutline "go.wdy.de/nago/presentation/icons/hero/outline"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"go.wdy.de/nago/presentation/ui/avatar"
	"go.wdy.de/nago/presentation/ui/tags"
)

const notSetText = "Nicht gesetzt"

// KeyValuesDetails shows a [secret.KeyValues] read-only. A concealed value is only sent to the browser by an action of
// the user, who reveals or copies it, and a write-only value never is. Pass the stored secret and not the result of
// [secret.WithoutWriteOnly], because only the stored secret tells whether a write-only value is set.
func KeyValuesDetails(kv secret.KeyValues) core.View {
	return tKeyValuesDetails{kv: kv}
}

type tKeyValuesDetails struct {
	kv secret.KeyValues
}

func (c tKeyValuesDetails) Render(ctx core.RenderContext) core.RenderNode {
	views := []core.View{c.header()}

	for s, section := range c.kv.Sections {
		rows := make([]core.View, 0, len(section.Fields))
		for f, field := range section.Fields {
			rows = append(rows, detailRow(ctx.Window(), field, strconv.Itoa(s)+"."+strconv.Itoa(f)))
		}

		if len(rows) == 0 {
			rows = append(rows, smallMuted("Keine Felder"))
		}

		views = append(views, sectionBlock(sectionHeading(section, nil), rows))
	}

	if len(c.kv.Sections) == 0 {
		views = append(views, smallMuted("Noch keine Felder"))
	}

	return ui.VStack(views...).Alignment(ui.Leading).Gap(ui.L24).FullWidth().Render(ctx)
}

func (c tKeyValuesDetails) header() core.View {
	var logo core.View = avatar.Text(c.kv.Name).Size(ui.L60)
	if c.kv.LogoSVG != "" {
		logo = tile(logoImage(c.kv.LogoSVG, ui.L40))
	}

	return ui.HStack(
		logo,
		ui.VStack(
			ui.Text(c.kv.Name).Font(ui.HeadlineSmall),
			ui.If(c.kv.Description != "", muted(ui.Text(c.kv.Description).Font(ui.BodyMedium))),
		).Alignment(ui.Leading).FullWidth(),
	).Gap(ui.L16).FullWidth()
}

func detailRow(wnd core.Window, field secret.KeyValueField, position string) core.View {
	// a field which has not been saved yet has no identifier
	id := position + "." + field.ID
	revealed := core.StateOf[bool](wnd, "nago.secret.keyvalues.reveal."+id)

	return ui.HStack(
		ui.VStack(
			smallMuted(fieldTitle(field)),
			detailValue(field, revealed.Get()),
			ui.If(field.SupportingText != "", smallMuted(field.SupportingText)),
		).Alignment(ui.Leading).Gap(ui.L2).FullWidth(),
		detailActions(wnd, field, id, revealed),
	).Gap(ui.L8).FullWidth()
}

func detailValue(field secret.KeyValueField, revealed bool) core.View {
	switch {
	case field.Protection == secret.WriteOnly:
		if field.Value != "" {
			return tags.StatusBadge(ui.SG0, "Gespeichert")
		}

		return tags.StatusBadge(ui.SW0, "Fehlt")
	case field.Value == "":
		return smallMuted(notSetText)
	case field.Protection == secret.Concealed && !revealed:
		return ui.Text(concealedText).Font(concealedFont)
	case field.Protection == secret.Concealed:
		return valueText(field.Value).Font(concealedFont)
	case field.Kind == secret.FieldBool:
		if field.Value == "true" {
			return ui.Text("Ja").Font(ui.BodyMedium)
		}

		return ui.Text("Nein").Font(ui.BodyMedium)
	case field.Kind == secret.FieldURL && isWebAddress(field.Value):
		return valueText(field.Value).Font(ui.BodyMedium).Link(field.Value, "_blank")
	default:
		return valueText(field.Value).Font(ui.BodyMedium)
	}
}

// detailActions are always visible, because a touch screen knows no hover.
func detailActions(wnd core.Window, field secret.KeyValueField, id string, revealed *core.State[bool]) core.View {
	if field.Protection == secret.WriteOnly || field.Value == "" {
		return nil
	}

	var reveal, copyValue core.View
	if field.Protection == secret.Concealed {
		icon, label := heroOutline.Eye, "Wert anzeigen"
		if revealed.Get() {
			icon, label = heroOutline.EyeSlash, "Wert verbergen"
		}

		reveal = ui.TertiaryButton(func() {
			revealed.Set(!revealed.Get())
		}).PreIcon(icon).AccessibilityLabel(label + ": " + fieldTitle(field))
	}

	if field.Kind != secret.FieldBool {
		copied := core.StateOf[string](wnd, "nago.secret.keyvalues.copied")
		icon := heroOutline.ClipboardDocument
		if copied.Get() == id {
			icon = heroOutline.ClipboardDocumentCheck
		}

		copyValue = ui.TertiaryButton(func() {
			if err := wnd.Clipboard().SetText(field.Value); err != nil {
				alert.ShowBannerError(wnd, err)
				return
			}

			copied.Set(id)
		}).PreIcon(icon).AccessibilityLabel("Wert kopieren: " + fieldTitle(field))
	}

	return ui.HStack(reveal, copyValue)
}

func valueText(value string) ui.TText {
	return ui.Text(value).WordBreak(ui.WordBreakBreakWord).WhiteSpace(ui.WhiteSpacePreWrap)
}

// isWebAddress keeps a value like javascript:alert(1) from becoming a link.
func isWebAddress(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https")
}
