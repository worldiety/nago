// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// All inline/phrasing elements implement PhrasingContent (and thus FlowContent).

// ── TextNode as PhrasingContent ───────────────────────────────────────────────

func (t *TextNode) flowContent()     {}
func (t *TextNode) phrasingContent() {}

// ── <span> ────────────────────────────────────────────────────────────────────

type Span struct{ element }

func NewSpan() *Span { return &Span{newElement("span")} }

func (e *Span) node()            {}
func (e *Span) flowContent()     {}
func (e *Span) phrasingContent() {}

func (e *Span) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Span) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <a> (transparent: implements both FlowContent and PhrasingContent) ────────

type A struct{ element }

func NewA() *A { return &A{newElement("a")} }

func (e *A) node()               {}
func (e *A) flowContent()        {}
func (e *A) phrasingContent()    {}
func (e *A) interactiveContent() {}

// AppendChild accepts PhrasingContent (most common use).
// For block-level transparent use, wrap children as FlowContent.
func (e *A) AppendChild(n PhrasingContent) { e.appendChildNode(n) }

// AppendFlow appends a FlowContent node (e.g. *Div) as a transparent child.
func (e *A) AppendFlow(n FlowContent) { e.appendChildNode(n) }
func (e *A) Render(w io.Writer) error { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <strong> ──────────────────────────────────────────────────────────────────

type Strong struct{ element }

func NewStrong() *Strong { return &Strong{newElement("strong")} }

func (e *Strong) node()            {}
func (e *Strong) flowContent()     {}
func (e *Strong) phrasingContent() {}

func (e *Strong) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Strong) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <em> ──────────────────────────────────────────────────────────────────────

type Em struct{ element }

func NewEm() *Em { return &Em{newElement("em")} }

func (e *Em) node()            {}
func (e *Em) flowContent()     {}
func (e *Em) phrasingContent() {}

func (e *Em) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Em) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <code> ────────────────────────────────────────────────────────────────────

type Code struct{ element }

func NewCode() *Code { return &Code{newElement("code")} }

func (e *Code) node()            {}
func (e *Code) flowContent()     {}
func (e *Code) phrasingContent() {}

func (e *Code) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Code) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <small> ───────────────────────────────────────────────────────────────────

type Small struct{ element }

func NewSmall() *Small { return &Small{newElement("small")} }

func (e *Small) node()            {}
func (e *Small) flowContent()     {}
func (e *Small) phrasingContent() {}

func (e *Small) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Small) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <b> ───────────────────────────────────────────────────────────────────────

type B struct{ element }

func NewB() *B { return &B{newElement("b")} }

func (e *B) node()            {}
func (e *B) flowContent()     {}
func (e *B) phrasingContent() {}

func (e *B) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *B) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <i> ───────────────────────────────────────────────────────────────────────

type I struct{ element }

func NewI() *I { return &I{newElement("i")} }

func (e *I) node()            {}
func (e *I) flowContent()     {}
func (e *I) phrasingContent() {}

func (e *I) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *I) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <u> ───────────────────────────────────────────────────────────────────────

type U struct{ element }

func NewU() *U { return &U{newElement("u")} }

func (e *U) node()            {}
func (e *U) flowContent()     {}
func (e *U) phrasingContent() {}

func (e *U) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *U) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <s> ───────────────────────────────────────────────────────────────────────

type S struct{ element }

func NewS() *S { return &S{newElement("s")} }

func (e *S) node()            {}
func (e *S) flowContent()     {}
func (e *S) phrasingContent() {}

func (e *S) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *S) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <cite> ────────────────────────────────────────────────────────────────────

type Cite struct{ element }

func NewCite() *Cite { return &Cite{newElement("cite")} }

func (e *Cite) node()            {}
func (e *Cite) flowContent()     {}
func (e *Cite) phrasingContent() {}

func (e *Cite) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Cite) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <q> ───────────────────────────────────────────────────────────────────────

type Q struct{ element }

func NewQ() *Q { return &Q{newElement("q")} }

func (e *Q) node()            {}
func (e *Q) flowContent()     {}
func (e *Q) phrasingContent() {}

func (e *Q) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Q) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <abbr> ────────────────────────────────────────────────────────────────────

type Abbr struct{ element }

func NewAbbr() *Abbr { return &Abbr{newElement("abbr")} }

func (e *Abbr) node()            {}
func (e *Abbr) flowContent()     {}
func (e *Abbr) phrasingContent() {}

func (e *Abbr) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Abbr) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <time> ────────────────────────────────────────────────────────────────────

type Time struct{ element }

func NewTime() *Time { return &Time{newElement("time")} }

func (e *Time) node()            {}
func (e *Time) flowContent()     {}
func (e *Time) phrasingContent() {}

func (e *Time) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Time) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <mark> ────────────────────────────────────────────────────────────────────

type Mark struct{ element }

func NewMark() *Mark { return &Mark{newElement("mark")} }

func (e *Mark) node()            {}
func (e *Mark) flowContent()     {}
func (e *Mark) phrasingContent() {}

func (e *Mark) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Mark) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <sub> ─────────────────────────────────────────────────────────────────────

type Sub struct{ element }

func NewSub() *Sub { return &Sub{newElement("sub")} }

func (e *Sub) node()            {}
func (e *Sub) flowContent()     {}
func (e *Sub) phrasingContent() {}

func (e *Sub) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Sub) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <sup> ─────────────────────────────────────────────────────────────────────

type Sup struct{ element }

func NewSup() *Sup { return &Sup{newElement("sup")} }

func (e *Sup) node()            {}
func (e *Sup) flowContent()     {}
func (e *Sup) phrasingContent() {}

func (e *Sup) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Sup) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <kbd> ─────────────────────────────────────────────────────────────────────

type Kbd struct{ element }

func NewKbd() *Kbd { return &Kbd{newElement("kbd")} }

func (e *Kbd) node()            {}
func (e *Kbd) flowContent()     {}
func (e *Kbd) phrasingContent() {}

func (e *Kbd) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Kbd) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <samp> ────────────────────────────────────────────────────────────────────

type Samp struct{ element }

func NewSamp() *Samp { return &Samp{newElement("samp")} }

func (e *Samp) node()            {}
func (e *Samp) flowContent()     {}
func (e *Samp) phrasingContent() {}

func (e *Samp) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Samp) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <var> ─────────────────────────────────────────────────────────────────────

type Var struct{ element }

func NewVar() *Var { return &Var{newElement("var")} }

func (e *Var) node()            {}
func (e *Var) flowContent()     {}
func (e *Var) phrasingContent() {}

func (e *Var) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Var) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <bdo> ─────────────────────────────────────────────────────────────────────

type Bdo struct{ element }

func NewBdo() *Bdo { return &Bdo{newElement("bdo")} }

func (e *Bdo) node()            {}
func (e *Bdo) flowContent()     {}
func (e *Bdo) phrasingContent() {}

func (e *Bdo) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Bdo) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <bdi> ─────────────────────────────────────────────────────────────────────

type Bdi struct{ element }

func NewBdi() *Bdi { return &Bdi{newElement("bdi")} }

func (e *Bdi) node()            {}
func (e *Bdi) flowContent()     {}
func (e *Bdi) phrasingContent() {}

func (e *Bdi) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Bdi) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <dfn> ─────────────────────────────────────────────────────────────────────

type Dfn struct{ element }

func NewDfn() *Dfn { return &Dfn{newElement("dfn")} }

func (e *Dfn) node()            {}
func (e *Dfn) flowContent()     {}
func (e *Dfn) phrasingContent() {}

func (e *Dfn) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Dfn) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <br> (void) ───────────────────────────────────────────────────────────────

type Br struct{ voidElement }

func NewBr() *Br { return &Br{newVoidElement("br")} }

func (e *Br) node()            {}
func (e *Br) flowContent()     {}
func (e *Br) phrasingContent() {}

// ── <wbr> (void) ──────────────────────────────────────────────────────────────

type Wbr struct{ voidElement }

func NewWbr() *Wbr { return &Wbr{newVoidElement("wbr")} }

func (e *Wbr) node()            {}
func (e *Wbr) flowContent()     {}
func (e *Wbr) phrasingContent() {}
