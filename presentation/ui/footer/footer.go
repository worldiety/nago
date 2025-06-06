// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package footer

import (
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type LinkOrNavigationPath = string
type TFooter struct {
	logo            ui.DecoredView
	slogan          string
	impress         LinkOrNavigationPath
	gdpr            LinkOrNavigationPath
	gtc             LinkOrNavigationPath
	termOfUse       LinkOrNavigationPath
	copyright       string
	backgroundColor ui.Color
	textColor       ui.Color
	contentPadding  ui.Length
}

func Footer() TFooter {
	return TFooter{contentPadding: ui.L48}
}

func (t TFooter) Logo(logo ui.DecoredView) TFooter {
	t.logo = logo
	return t
}

func (t TFooter) Slogan(slogan string) TFooter {
	t.slogan = slogan
	return t
}

func (t TFooter) ProviderName(copyright string) TFooter {
	t.copyright = copyright
	return t
}

func (t TFooter) Impress(impress LinkOrNavigationPath) TFooter {
	t.impress = impress
	return t
}

func (t TFooter) PrivacyPolicy(gdpr LinkOrNavigationPath) TFooter {
	t.gdpr = gdpr
	return t
}

func (t TFooter) GeneralTermsAndConditions(gtc LinkOrNavigationPath) TFooter {
	t.gtc = gtc
	return t
}

func (t TFooter) TermsOfUse(termsOfUse LinkOrNavigationPath) TFooter {
	t.termOfUse = termsOfUse
	return t
}

func (t TFooter) BackgroundColor(backgroundColor ui.Color) TFooter {
	t.backgroundColor = backgroundColor
	return t
}

func (t TFooter) TextColor(textColor ui.Color) TFooter {
	t.textColor = textColor
	return t
}

// ContentPadding is a padding between the footer line separator and the
// view before the footer (typically the actual page content). Set to "" to
// disable any padding.
func (t TFooter) ContentPadding(content ui.Length) TFooter {
	t.contentPadding = content
	return t
}

func (t TFooter) Render(ctx core.RenderContext) core.RenderNode {
	anyLegal := t.impress != "" || t.gtc != "" || t.gdpr != ""
	height := ui.L160

	if t.slogan != "" && anyLegal {
		height = ui.L200
	}
	wnd := ctx.Window()
	return ui.VStack(
		ui.If(t.contentPadding != "", ui.Space(t.contentPadding)),
		ui.VStack(
			ui.IfFunc(t.logo != nil, func() core.View {
				return t.logo.Frame(ui.Frame{Height: ui.L64})
			}),
			ui.If(t.slogan != "", ui.Text(t.slogan)),
			ui.IfFunc(anyLegal, func() core.View {
				return ui.HStack(
					ui.If(t.impress != "", ui.Link(wnd, "Impressum", t.impress, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.gdpr != "", ui.Link(wnd, "Datenschutz", t.gdpr, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.gtc != "", ui.Link(wnd, "AGB", t.gtc, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.termOfUse != "", ui.Link(wnd, "Nutzungsbedingungen", t.termOfUse, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.copyright != "", ui.Text(t.copyright).Padding(ui.Padding{Left: ui.L16})),
				).Gap(ui.L16)
			}),
		).TextColor(t.textColor).
			Gap(ui.L24).
			BackgroundColor(t.backgroundColor).
			Border(ui.Border{TopColor: ui.M6, TopWidth: ui.L1}).
			Frame(ui.Frame{Width: ui.Full, Height: height}),
	).FullWidth().Render(ctx)
}
