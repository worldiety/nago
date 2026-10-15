// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License
//
//gostrict:header
package ui

import (
	"go.wdy.de/nago/pkg/gostrict/std/js"
	nui "go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/footer"
)

// FooterLinkOrNavigationPath is a link or a navigation path, like footer.LinkOrNavigationPath of nago.
type FooterLinkOrNavigationPath = string

// TFooter is the footer of an application or website, like footer.TFooter of nago. See Footer.
type TFooter struct{ ref js.Ref }

func (TFooter) isView() {}

// Footer creates a footer with the default content padding, like footer.Footer of nago.
//
//gostrict:host
func Footer() TFooter {
	return TFooter{ref: wrap(footer.Footer())}
}

// Slogan sets the slogan text of the footer.
//
//gostrict:host
func (t TFooter) Slogan(slogan string) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).Slogan(slogan))}
}

// ProviderName sets the copyright or provider name of the footer.
//
//gostrict:host
func (t TFooter) ProviderName(copyright string) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).ProviderName(copyright))}
}

// Impress sets the link to the imprint.
//
//gostrict:host
func (t TFooter) Impress(impress FooterLinkOrNavigationPath) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).Impress(impress))}
}

// PrivacyPolicy sets the link to the privacy policy.
//
//gostrict:host
func (t TFooter) PrivacyPolicy(gdpr FooterLinkOrNavigationPath) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).PrivacyPolicy(gdpr))}
}

// GeneralTermsAndConditions sets the link to the general terms and conditions.
//
//gostrict:host
func (t TFooter) GeneralTermsAndConditions(gtc FooterLinkOrNavigationPath) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).GeneralTermsAndConditions(gtc))}
}

// TermsOfUse sets the link to the terms of use.
//
//gostrict:host
func (t TFooter) TermsOfUse(termsOfUse FooterLinkOrNavigationPath) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).TermsOfUse(termsOfUse))}
}

// BackgroundColor sets the background color of the footer.
//
//gostrict:host
func (t TFooter) BackgroundColor(backgroundColor Color) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).BackgroundColor(nui.Color(backgroundColor)))}
}

// TextColor sets the text color of the footer.
//
//gostrict:host
func (t TFooter) TextColor(textColor Color) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).TextColor(nui.Color(textColor)))}
}

// ContentPadding sets the padding between the separator line and the view before the footer. Set it to "" to disable
// the padding.
//
//gostrict:host
func (t TFooter) ContentPadding(content Length) TFooter {
	return TFooter{ref: wrap(get[footer.TFooter](t.ref).ContentPadding(nui.Length(content)))}
}
