// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// All sectioning elements accept FlowContent children.
// They implement SectioningContent (which embeds FlowContent).

// ── <body> ────────────────────────────────────────────────────────────────────

type Body struct{ element }

func NewBody() *Body { return &Body{newElement("body")} }

func (e *Body) node()        {}
func (e *Body) flowContent() {}

func (e *Body) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Body) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <main> ────────────────────────────────────────────────────────────────────

type Main struct{ element }

func NewMain() *Main { return &Main{newElement("main")} }

func (e *Main) node()              {}
func (e *Main) flowContent()       {}
func (e *Main) sectioningContent() {}

func (e *Main) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Main) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <article> ─────────────────────────────────────────────────────────────────

type Article struct{ element }

func NewArticle() *Article { return &Article{newElement("article")} }

func (e *Article) node()              {}
func (e *Article) flowContent()       {}
func (e *Article) sectioningContent() {}

func (e *Article) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Article) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <section> ─────────────────────────────────────────────────────────────────

type Section struct{ element }

func NewSection() *Section { return &Section{newElement("section")} }

func (e *Section) node()              {}
func (e *Section) flowContent()       {}
func (e *Section) sectioningContent() {}

func (e *Section) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Section) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <aside> ───────────────────────────────────────────────────────────────────

type Aside struct{ element }

func NewAside() *Aside { return &Aside{newElement("aside")} }

func (e *Aside) node()              {}
func (e *Aside) flowContent()       {}
func (e *Aside) sectioningContent() {}

func (e *Aside) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Aside) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <nav> ─────────────────────────────────────────────────────────────────────

type Nav struct{ element }

func NewNav() *Nav { return &Nav{newElement("nav")} }

func (e *Nav) node()              {}
func (e *Nav) flowContent()       {}
func (e *Nav) sectioningContent() {}

func (e *Nav) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Nav) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <header> ──────────────────────────────────────────────────────────────────

type Header struct{ element }

func NewHeader() *Header { return &Header{newElement("header")} }

func (e *Header) node()        {}
func (e *Header) flowContent() {}

func (e *Header) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Header) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <footer> ──────────────────────────────────────────────────────────────────

type Footer struct{ element }

func NewFooter() *Footer { return &Footer{newElement("footer")} }

func (e *Footer) node()        {}
func (e *Footer) flowContent() {}

func (e *Footer) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Footer) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <address> ─────────────────────────────────────────────────────────────────

type Address struct{ element }

func NewAddress() *Address { return &Address{newElement("address")} }

func (e *Address) node()        {}
func (e *Address) flowContent() {}

func (e *Address) AppendChild(n FlowContent) { e.appendChildNode(n) }

func (e *Address) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}
