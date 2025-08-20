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

// TFooter is a overlay component (Footer).
// This component models the footer of an application or website,
// typically containing branding, legal links, and styling options.
type TFooter struct {
	logo            ui.DecoredView       // logo displayed in the footer
	slogan          string               // slogan or tagline text
	impress         LinkOrNavigationPath // link to the imprint page
	gdpr            LinkOrNavigationPath // link to the GDPR/privacy policy
	gtc             LinkOrNavigationPath // link to the general terms and conditions
	termOfUse       LinkOrNavigationPath // link to the terms of use
	copyright       string               // copyright notice text
	backgroundColor ui.Color             // background color of the footer
	textColor       ui.Color             // text color used inside the footer
	contentPadding  ui.Length            // padding applied to the footer content
}

// Footer creates a new TFooter with default content padding.
func Footer() TFooter {
	return TFooter{contentPadding: ui.L48}
}

// Logo sets the logo view of the footer.
func (t TFooter) Logo(logo ui.DecoredView) TFooter {
	t.logo = logo
	return t
}

// Slogan sets the slogan text of the footer.
func (t TFooter) Slogan(slogan string) TFooter {
	t.slogan = slogan
	return t
}

// ProviderName sets the copyright/provider name of the footer.
func (t TFooter) ProviderName(copyright string) TFooter {
	t.copyright = copyright
	return t
}

// Impress sets the imprint link of the footer.
func (t TFooter) Impress(impress LinkOrNavigationPath) TFooter {
	t.impress = impress
	return t
}

// PrivacyPolicy sets the privacy policy (GDPR) link of the footer.
func (t TFooter) PrivacyPolicy(gdpr LinkOrNavigationPath) TFooter {
	t.gdpr = gdpr
	return t
}

// GeneralTermsAndConditions sets the GTC link of the footer.
func (t TFooter) GeneralTermsAndConditions(gtc LinkOrNavigationPath) TFooter {
	t.gtc = gtc
	return t
}

// TermsOfUse sets the terms of use link of the footer.
func (t TFooter) TermsOfUse(termsOfUse LinkOrNavigationPath) TFooter {
	t.termOfUse = termsOfUse
	return t
}

// BackgroundColor sets the background color of the footer.
func (t TFooter) BackgroundColor(backgroundColor ui.Color) TFooter {
	t.backgroundColor = backgroundColor
	return t
}

// TextColor sets the text color of the footer.
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

// Render builds and returns the RenderNode for the TFooter.
// It arranges the footer into the following parts:
// - Optional spacing at the top (content padding)
// - Logo (if provided)
// - Slogan (if provided)
// - Legal section with links (Impressum, GDPR, GTC, Terms of Use) and copyright
// The layout adapts between vertical stacking on mobile and horizontal
// stacking on larger screens.
func (t TFooter) Render(ctx core.RenderContext) core.RenderNode {
	anyLegal := t.impress != "" || t.gtc != "" || t.gdpr != "" || t.copyright != ""
	height := ui.L160

	if t.slogan != "" && anyLegal {
		height = ui.L200
	}

	wnd := ctx.Window()
	isMobile := ctx.Window().Info().SizeClass <= core.SizeClassSmall

	return ui.VStack(
		ui.If(t.contentPadding != "", ui.Space(t.contentPadding)),
		ui.VStack(
			ui.IfFunc(t.logo != nil, func() core.View {
				return t.logo.WithFrame(func(frame ui.Frame) ui.Frame {
					if frame.IsZero() {
						return ui.Frame{Height: ui.L64}
					}

					return frame
				})
			}),
			ui.If(t.slogan != "", ui.Text(t.slogan)),
			ui.IfFunc(anyLegal, func() core.View {
				var tmp []core.View
				tmp = append(tmp,
					ui.If(t.impress != "", ui.Link(wnd, "Impressum", t.impress, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.gdpr != "", ui.Link(wnd, "Datenschutz", t.gdpr, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.gtc != "", ui.Link(wnd, "AGB", t.gtc, ui.LinkTargetNewWindowOrTab)),
					ui.If(t.termOfUse != "", ui.Link(wnd, "Nutzungsbedingungen", t.termOfUse, ui.LinkTargetNewWindowOrTab)),
					ui.If(!isMobile, ui.Space(ui.L16)),
					ui.If(t.copyright != "", ui.Text(t.copyright)),
				)

				if isMobile {
					return ui.VStack(tmp...).Gap(ui.L16).Alignment(ui.Leading).FullWidth()
				}

				return ui.HStack(tmp...).Gap(ui.L16)
			}),
		).TextColor(t.textColor).
			Gap(ui.L24).
			BackgroundColor(t.backgroundColor).
			Border(ui.Border{TopColor: ui.M6, TopWidth: ui.L1}).
			Padding(ui.Padding{}.All(ui.L16)).
			Frame(ui.Frame{Width: ui.Full, MinHeight: height}),
	).FullWidth().Render(ctx)
}
