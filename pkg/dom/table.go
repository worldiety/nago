// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <table> ───────────────────────────────────────────────────────────────────

type Table struct{ element }

func NewTable() *Table { return &Table{newElement("table")} }

func (e *Table) node()        {}
func (e *Table) flowContent() {}

// AppendChild accepts <caption>, <colgroup>, <thead>, <tbody>, <tfoot>.
func (e *Table) AppendChild(n TableContent) { e.appendChildNode(n) }
func (e *Table) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <caption> ─────────────────────────────────────────────────────────────────

type Caption struct{ element }

func NewCaption() *Caption { return &Caption{newElement("caption")} }

func (e *Caption) node()         {}
func (e *Caption) tableContent() {}

func (e *Caption) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Caption) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <colgroup> ────────────────────────────────────────────────────────────────

type Colgroup struct{ element }

func NewColgroup() *Colgroup { return &Colgroup{newElement("colgroup")} }

func (e *Colgroup) node()         {}
func (e *Colgroup) tableContent() {}

// AppendChild accepts only <col>.
func (e *Colgroup) AppendChild(n ColGroupContent) { e.appendChildNode(n) }
func (e *Colgroup) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <col> (void) ──────────────────────────────────────────────────────────────

type Col struct{ voidElement }

func NewCol() *Col { return &Col{newVoidElement("col")} }

func (e *Col) node()            {}
func (e *Col) colGroupContent() {}

// ── <thead> ───────────────────────────────────────────────────────────────────

type Thead struct{ element }

func NewThead() *Thead { return &Thead{newElement("thead")} }

func (e *Thead) node()         {}
func (e *Thead) tableContent() {}

// AppendChild accepts only <tr>.
func (e *Thead) AppendChild(n TableSectionContent) { e.appendChildNode(n) }
func (e *Thead) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <tbody> ───────────────────────────────────────────────────────────────────

type Tbody struct{ element }

func NewTbody() *Tbody { return &Tbody{newElement("tbody")} }

func (e *Tbody) node()         {}
func (e *Tbody) tableContent() {}

func (e *Tbody) AppendChild(n TableSectionContent) { e.appendChildNode(n) }
func (e *Tbody) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <tfoot> ───────────────────────────────────────────────────────────────────

type Tfoot struct{ element }

func NewTfoot() *Tfoot { return &Tfoot{newElement("tfoot")} }

func (e *Tfoot) node()         {}
func (e *Tfoot) tableContent() {}

func (e *Tfoot) AppendChild(n TableSectionContent) { e.appendChildNode(n) }
func (e *Tfoot) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <tr> ──────────────────────────────────────────────────────────────────────

type Tr struct{ element }

func NewTr() *Tr { return &Tr{newElement("tr")} }

func (e *Tr) node()                {}
func (e *Tr) tableSectionContent() {}

// AppendChild accepts only <th> and <td>.
func (e *Tr) AppendChild(n TableRowContent) { e.appendChildNode(n) }
func (e *Tr) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <th> ──────────────────────────────────────────────────────────────────────

type Th struct{ element }

func NewTh() *Th { return &Th{newElement("th")} }

func (e *Th) node()            {}
func (e *Th) tableRowContent() {}

func (e *Th) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Th) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <td> ──────────────────────────────────────────────────────────────────────

type Td struct{ element }

func NewTd() *Td { return &Td{newElement("td")} }

func (e *Td) node()            {}
func (e *Td) tableRowContent() {}

func (e *Td) AppendChild(n FlowContent) { e.appendChildNode(n) }
func (e *Td) Render(w io.Writer) error  { return renderElement(e.tag, e.attrs, e.kids, false, w) }
