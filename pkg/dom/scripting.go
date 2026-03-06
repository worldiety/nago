// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <noscript> ────────────────────────────────────────────────────────────────

type Noscript struct{ element }

func NewNoscript() *Noscript { return &Noscript{newElement("noscript")} }

func (e *Noscript) node()            {}
func (e *Noscript) flowContent()     {}
func (e *Noscript) phrasingContent() {}
func (e *Noscript) metadataContent() {}

func (e *Noscript) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Noscript) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <template> ────────────────────────────────────────────────────────────────

type Template struct{ element }

func NewTemplate() *Template { return &Template{newElement("template")} }

func (e *Template) node()            {}
func (e *Template) flowContent()     {}
func (e *Template) phrasingContent() {}
func (e *Template) metadataContent() {}

func (e *Template) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Template) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <slot> ────────────────────────────────────────────────────────────────────

type Slot struct{ element }

func NewSlot() *Slot { return &Slot{newElement("slot")} }

func (e *Slot) node()            {}
func (e *Slot) flowContent()     {}
func (e *Slot) phrasingContent() {}

func (e *Slot) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Slot) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <details> ─────────────────────────────────────────────────────────────────

type Details struct{ element }

func NewDetails() *Details { return &Details{newElement("details")} }

func (e *Details) node()               {}
func (e *Details) flowContent()        {}
func (e *Details) interactiveContent() {}

// AppendChild accepts <summary> and FlowContent.
func (e *Details) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Details) AppendSummary(n *Summary)  { e.appendChildNode(n) }
func (e *Details) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <summary> ─────────────────────────────────────────────────────────────────

type Summary struct{ element }

func NewSummary() *Summary { return &Summary{newElement("summary")} }

func (e *Summary) node()        {}
func (e *Summary) flowContent() {}

func (e *Summary) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Summary) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <dialog> ──────────────────────────────────────────────────────────────────

type Dialog struct{ element }

func NewDialog() *Dialog { return &Dialog{newElement("dialog")} }

func (e *Dialog) node()        {}
func (e *Dialog) flowContent() {}

func (e *Dialog) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Dialog) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <menu> ────────────────────────────────────────────────────────────────────

type Menu struct{ element }

func NewMenu() *Menu { return &Menu{newElement("menu")} }

func (e *Menu) node()        {}
func (e *Menu) flowContent() {}

func (e *Menu) AppendChild(n ListContent) { e.appendChildNode(n) }
func (e *Menu) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <data> ────────────────────────────────────────────────────────────────────

type Data struct{ element }

func NewData() *Data { return &Data{newElement("data")} }

func (e *Data) node()            {}
func (e *Data) flowContent()     {}
func (e *Data) phrasingContent() {}

func (e *Data) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Data) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <ins> (transparent) ───────────────────────────────────────────────────────

type Ins struct{ element }

func NewIns() *Ins { return &Ins{newElement("ins")} }

func (e *Ins) node()            {}
func (e *Ins) flowContent()     {}
func (e *Ins) phrasingContent() {}

func (e *Ins) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Ins) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <del> (transparent) ───────────────────────────────────────────────────────

type Del struct{ element }

func NewDel() *Del { return &Del{newElement("del")} }

func (e *Del) node()            {}
func (e *Del) flowContent()     {}
func (e *Del) phrasingContent() {}

func (e *Del) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Del) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <search> ──────────────────────────────────────────────────────────────────

type Search struct{ element }

func NewSearch() *Search { return &Search{newElement("search")} }

func (e *Search) node()        {}
func (e *Search) flowContent() {}

func (e *Search) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Search) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}
