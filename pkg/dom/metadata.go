// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import (
	"bytes"
	"io"
)

// ── <head> ────────────────────────────────────────────────────────────────────

// Head represents the <head> element.
// Only MetadataContent is permitted as children.
type Head struct{ element }

func NewHead() *Head { return &Head{newElement("head")} }

func (h *Head) node()            {}
func (h *Head) metadataContent() {}

// AppendChild appends a MetadataContent node to <head>.
func (h *Head) AppendChild(n MetadataContent) {
	h.appendChildNode(n)
}

func (h *Head) Render(w io.Writer) error {
	return renderElement(h.tag, h.attrs, h.kids, false, w)
}

// ── <title> ───────────────────────────────────────────────────────────────────

// Title represents the <title> element.
type Title struct{ element }

func NewTitle(text string) *Title {
	t := &Title{newElement("title")}
	t.kids = []Node{NewTextNode(text)}
	return t
}

func (t *Title) node()            {}
func (t *Title) metadataContent() {}

func (t *Title) Render(w io.Writer) error {
	return renderElement(t.tag, t.attrs, t.kids, false, w)
}

// ── <style> ───────────────────────────────────────────────────────────────────

// Style represents the <style> element.
// Its content is raw CSS text (not HTML-escaped).
type Style struct {
	element
	css string
}

func NewStyle(css string) *Style {
	s := &Style{element: newElement("style"), css: css}
	return s
}

func (s *Style) node()            {}
func (s *Style) metadataContent() {}

func (s *Style) Render(w io.Writer) error {
	if _, err := io.WriteString(w, "<style>"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, s.css); err != nil {
		return err
	}
	_, err := io.WriteString(w, "</style>")
	return err
}

// SetCSS replaces the raw CSS content.
func (s *Style) SetCSS(css string) { s.css = css }

// CSS returns the raw CSS content.
func (s *Style) CSS() string { return s.css }

// ── <script> ──────────────────────────────────────────────────────────────────

// Script represents the <script> element.
// Inline JS content is written raw (not HTML-escaped).
type Script struct {
	element
	inline string
}

func NewScript() *Script { return &Script{element: newElement("script")} }

// NewInlineScript creates a <script> with inline JS content.
func NewInlineScript(js string) *Script {
	return &Script{element: newElement("script"), inline: js}
}

func (s *Script) node()            {}
func (s *Script) metadataContent() {}
func (s *Script) flowContent()     {}

func (s *Script) Render(w io.Writer) error {
	if _, err := io.WriteString(w, "<script"); err != nil {
		return err
	}
	var b bytes.Buffer
	_ = renderElement("script", s.attrs, nil, false, &b)
	// reuse attribute rendering by stripping the wrapping tag
	attrStr := b.String()
	// attrStr is "<script ...></script>", extract attributes part
	if len(attrStr) > len("<script></script>") {
		attrPart := attrStr[len("<script") : len(attrStr)-len("></script>")]
		if _, err := io.WriteString(w, attrPart); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	if s.inline != "" {
		if _, err := io.WriteString(w, s.inline); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</script>")
	return err
}

// SetInlineJS replaces inline JavaScript content.
func (s *Script) SetInlineJS(js string) { s.inline = js }

// ── <meta> (void) ─────────────────────────────────────────────────────────────

// Meta represents the <meta> void element.
type Meta struct{ voidElement }

func NewMeta() *Meta { return &Meta{newVoidElement("meta")} }

func (m *Meta) node()            {}
func (m *Meta) metadataContent() {}

// ── <link> (void) ─────────────────────────────────────────────────────────────

// Link represents the <link> void element.
type Link struct{ voidElement }

func NewLink() *Link { return &Link{newVoidElement("link")} }

func (l *Link) node()            {}
func (l *Link) metadataContent() {}
func (l *Link) flowContent()     {}
func (l *Link) phrasingContent() {}

// ── <base> (void) ─────────────────────────────────────────────────────────────

// Base represents the <base> void element.
type Base struct{ voidElement }

func NewBase() *Base { return &Base{newVoidElement("base")} }

func (b *Base) node()            {}
func (b *Base) metadataContent() {}
