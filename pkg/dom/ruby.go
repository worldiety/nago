// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <ruby> ────────────────────────────────────────────────────────────────────

type Ruby struct{ element }

func NewRuby() *Ruby { return &Ruby{newElement("ruby")} }

func (e *Ruby) node()            {}
func (e *Ruby) flowContent()     {}
func (e *Ruby) phrasingContent() {}

// AppendChild accepts <rt> and <rp> (RubyContent).
func (e *Ruby) AppendChild(n RubyContent)    { e.appendChildNode(n) }
func (e *Ruby) AppendText(n PhrasingContent) { e.appendChildNode(n) }
func (e *Ruby) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── <rt> ──────────────────────────────────────────────────────────────────────

type Rt struct{ element }

func NewRt() *Rt { return &Rt{newElement("rt")} }

func (e *Rt) node()        {}
func (e *Rt) rubyContent() {}

func (e *Rt) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Rt) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

// ── <rp> ──────────────────────────────────────────────────────────────────────

type Rp struct{ element }

func NewRp() *Rp { return &Rp{newElement("rp")} }

func (e *Rp) node()        {}
func (e *Rp) rubyContent() {}

func (e *Rp) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *Rp) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }
