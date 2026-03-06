// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import "io"

// ── <h1> – <h6> ──────────────────────────────────────────────────────────────

type H1 struct{ element }
type H2 struct{ element }
type H3 struct{ element }
type H4 struct{ element }
type H5 struct{ element }
type H6 struct{ element }

func NewH1() *H1 { return &H1{newElement("h1")} }
func NewH2() *H2 { return &H2{newElement("h2")} }
func NewH3() *H3 { return &H3{newElement("h3")} }
func NewH4() *H4 { return &H4{newElement("h4")} }
func NewH5() *H5 { return &H5{newElement("h5")} }
func NewH6() *H6 { return &H6{newElement("h6")} }

// All heading elements are FlowContent and HeadingContent.
// Their AppendChild accepts PhrasingContent.

func (e *H1) node()                         {}
func (e *H1) flowContent()                  {}
func (e *H1) headingContent()               {}
func (e *H1) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H1) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

func (e *H2) node()                         {}
func (e *H2) flowContent()                  {}
func (e *H2) headingContent()               {}
func (e *H2) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H2) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

func (e *H3) node()                         {}
func (e *H3) flowContent()                  {}
func (e *H3) headingContent()               {}
func (e *H3) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H3) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

func (e *H4) node()                         {}
func (e *H4) flowContent()                  {}
func (e *H4) headingContent()               {}
func (e *H4) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H4) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

func (e *H5) node()                         {}
func (e *H5) flowContent()                  {}
func (e *H5) headingContent()               {}
func (e *H5) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H5) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }

func (e *H6) node()                         {}
func (e *H6) flowContent()                  {}
func (e *H6) headingContent()               {}
func (e *H6) AppendChild(n PhrasingContent) { e.appendChildNode(n) }
func (e *H6) Render(w io.Writer) error      { return renderElement(e.tag, e.attrs, e.kids, false, w) }
