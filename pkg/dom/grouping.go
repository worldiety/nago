// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <div> ─────────────────────────────────────────────────────────────────────

type Div struct{ element }

func NewDiv() *Div { return &Div{newElement("div")} }

func (e *Div) node()        {}
func (e *Div) flowContent() {}

func (e *Div) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Div) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <p> ───────────────────────────────────────────────────────────────────────

type P struct{ element }

func NewP() *P { return &P{newElement("p")} }

func (e *P) node()        {}
func (e *P) flowContent() {}

func (e *P) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *P) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <pre> ─────────────────────────────────────────────────────────────────────

type Pre struct{ element }

func NewPre() *Pre { return &Pre{newElement("pre")} }

func (e *Pre) node()        {}
func (e *Pre) flowContent() {}

func (e *Pre) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Pre) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <blockquote> ──────────────────────────────────────────────────────────────

type Blockquote struct{ element }

func NewBlockquote() *Blockquote { return &Blockquote{newElement("blockquote")} }

func (e *Blockquote) node()        {}
func (e *Blockquote) flowContent() {}

func (e *Blockquote) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Blockquote) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <ul> ──────────────────────────────────────────────────────────────────────

type Ul struct{ element }

func NewUl() *Ul { return &Ul{newElement("ul")} }

func (e *Ul) node()        {}
func (e *Ul) flowContent() {}

// AppendChild accepts only <li>.
func (e *Ul) AppendChild(n ListContent) { e.appendChildNode(n) }
func (e *Ul) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <ol> ──────────────────────────────────────────────────────────────────────

type Ol struct{ element }

func NewOl() *Ol { return &Ol{newElement("ol")} }

func (e *Ol) node()        {}
func (e *Ol) flowContent() {}

// AppendChild accepts only <li>.
func (e *Ol) AppendChild(n ListContent) { e.appendChildNode(n) }
func (e *Ol) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <li> ──────────────────────────────────────────────────────────────────────

type Li struct{ element }

func NewLi() *Li { return &Li{newElement("li")} }

func (e *Li) node()        {}
func (e *Li) flowContent() {}
func (e *Li) listContent() {}

func (e *Li) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Li) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <dl> ──────────────────────────────────────────────────────────────────────

type Dl struct{ element }

func NewDl() *Dl { return &Dl{newElement("dl")} }

func (e *Dl) node()        {}
func (e *Dl) flowContent() {}

// AppendChild accepts only <dt> and <dd>.
func (e *Dl) AppendChild(n DlContent)  { e.appendChildNode(n) }
func (e *Dl) Render(w io.Writer) error { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <dt> ──────────────────────────────────────────────────────────────────────

type Dt struct{ element }

func NewDt() *Dt { return &Dt{newElement("dt")} }

func (e *Dt) node()        {}
func (e *Dt) flowContent() {}
func (e *Dt) dlContent()   {}

func (e *Dt) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Dt) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <dd> ──────────────────────────────────────────────────────────────────────

type Dd struct{ element }

func NewDd() *Dd { return &Dd{newElement("dd")} }

func (e *Dd) node()        {}
func (e *Dd) flowContent() {}
func (e *Dd) dlContent()   {}

func (e *Dd) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Dd) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <figure> ──────────────────────────────────────────────────────────────────

type Figure struct{ element }

func NewFigure() *Figure { return &Figure{newElement("figure")} }

func (e *Figure) node()          {}
func (e *Figure) flowContent()   {}
func (e *Figure) figureContent() {}

// AppendChild accepts FlowContent and Figcaption (FigureContent).
func (e *Figure) AppendChild(n FigureContent) { e.appendChildNode(n) }
func (e *Figure) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <figcaption> ──────────────────────────────────────────────────────────────

type Figcaption struct{ element }

func NewFigcaption() *Figcaption { return &Figcaption{newElement("figcaption")} }

func (e *Figcaption) node()          {}
func (e *Figcaption) flowContent()   {}
func (e *Figcaption) figureContent() {}

func (e *Figcaption) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Figcaption) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <hr> (void) ───────────────────────────────────────────────────────────────

type Hr struct{ voidElement }

func NewHr() *Hr { return &Hr{newVoidElement("hr")} }

func (e *Hr) node()        {}
func (e *Hr) flowContent() {}
