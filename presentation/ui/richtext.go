// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui

import (
	"github.com/worldiety/i18n"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"golang.org/x/text/language"
)

var (
	StrConfirmExternalLink = i18n.MustString("richText.confirm_external_link", i18n.Values{language.German: "Diese externe Adresse in einem neuen Tab öffnen?", language.English: "Open this external address in a new tab?"})
)

// TRichText is a basic component (Rich Text).
// It renders a block of rich text content, which may include formatting
// such as bold, italic, links, or other markup depending on the renderer.
// The component supports layout control via frame settings.
//
// A relative link of the own origin navigates within the application, like [TStack.HRef], instead of loading
// the page anew. An external link opens in a new tab, which gets no access to this one.
type TRichText struct {
	value               string // raw rich text value
	frame               Frame  // layout frame
	confirmExternalLink bool
}

// RichText creates a new rich text component with the given value.
func RichText(value string) TRichText {
	return TRichText{
		value: value,
	}
}

// Value sets the rich text content to be rendered.
func (c TRichText) Value(value string) TRichText {
	c.value = value
	return c
}

// Frame sets the layout frame for the rich text component.
func (c TRichText) Frame(frame Frame) TRichText {
	c.frame = frame
	return c
}

// ConfirmExternalLinks asks the user before an external link is opened and shows its full address. Use it for
// text which neither the developer nor the user wrote, e.g. the answer of an AI model, whose links may carry
// data to a foreign server.
func (c TRichText) ConfirmExternalLinks(confirm bool) TRichText {
	c.confirmExternalLink = confirm
	return c
}

// FullWidth expands the rich text component to take the full available width.
func (c TRichText) FullWidth() TRichText {
	c.frame = c.frame.FullWidth()
	return c
}

// Render builds and returns the protocol representation of the rich text component,
// including its value and layout frame.
func (c TRichText) Render(ctx core.RenderContext) core.RenderNode {
	var confirmation string
	if c.confirmExternalLink {
		confirmation = StrConfirmExternalLink.Get(ctx.Window())
	}

	return &proto.RichText{
		Value:                    proto.Str(c.value),
		Frame:                    c.frame.ora(),
		ExternalLinkConfirmation: proto.Str(confirmation),
	}
}
