// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom

import (
	"bytes"
	"html"
	"io"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// element is the shared base struct embedded by every non-void HTML element.
type element struct {
	tag    string
	attrs  map[string]string
	parent Node
	kids   []Node
}

func newElement(tag string) element {
	return element{tag: tag, attrs: make(map[string]string)}
}

// ── Node sentinel & internal accessors ────────────────────────────────────────

func (e *element) node() {}

func (e *element) tagName() string  { return e.tag }
func (e *element) isVoid() bool     { return false }
func (e *element) children() []Node { return e.kids }

func (e *element) appendChildNode(n Node) {
	e.kids = append(e.kids, n)
}

// ── Tree ──────────────────────────────────────────────────────────────────────

func (e *element) Parent() Node { return e.parent }

// ── Attributes ────────────────────────────────────────────────────────────────

func (e *element) SetAttr(key, value string) { e.attrs[key] = value }
func (e *element) GetAttr(key string) string { return e.attrs[key] }
func (e *element) RemoveAttr(key string)     { delete(e.attrs, key) }

// ── Text content ──────────────────────────────────────────────────────────────

// SetTextContent replaces all children with a single TextNode.
func (e *element) SetTextContent(s string) {
	e.kids = []Node{NewTextNode(s)}
}

// TextContent returns the concatenated text of all descendant text nodes.
func (e *element) TextContent() string {
	var b strings.Builder
	collectText(e.kids, &b)
	return b.String()
}

func collectText(kids []Node, b *strings.Builder) {
	for _, k := range kids {
		if tn, ok := k.(*TextNode); ok {
			b.WriteString(tn.content)
		} else {
			collectText(k.children(), b)
		}
	}
}

// ── Inner HTML ────────────────────────────────────────────────────────────────

// SetInnerHTML parses raw HTML (best-effort via x/net/html ParseFragment)
// and replaces all children with the resulting node tree.
func (e *element) SetInnerHTML(raw string) {
	a := atom.Lookup([]byte(e.tag))
	ctx := &xhtml.Node{
		Type:     xhtml.ElementNode,
		Data:     e.tag,
		DataAtom: a,
	}
	nodes, err := xhtml.ParseFragment(strings.NewReader(raw), ctx)
	if err != nil || len(nodes) == 0 {
		// best-effort: fall back to plain text node
		e.kids = []Node{NewTextNode(raw)}
		return
	}
	e.kids = make([]Node, 0, len(nodes))
	for _, n := range nodes {
		e.kids = append(e.kids, convertXNode(n))
	}
}

// InnerHTML renders all children into a string.
func (e *element) InnerHTML() string {
	var b bytes.Buffer
	for _, k := range e.kids {
		_ = k.Render(&b)
	}
	return b.String()
}

// ── Rendering ─────────────────────────────────────────────────────────────────

func (e *element) Render(w io.Writer) error {
	return renderElement(e.tag, e.attrs, e.kids, false, w)
}

// ── voidElement ───────────────────────────────────────────────────────────────

// voidElement is the shared base for elements that must not have children.
type voidElement struct {
	tag   string
	attrs map[string]string
	par   Node
}

func newVoidElement(tag string) voidElement {
	return voidElement{tag: tag, attrs: make(map[string]string)}
}

func (v *voidElement) node()                    {}
func (v *voidElement) tagName() string          { return v.tag }
func (v *voidElement) isVoid() bool             { return true }
func (v *voidElement) children() []Node         { return nil }
func (v *voidElement) appendChildNode(_ Node)   {} // no-op
func (v *voidElement) Parent() Node             { return v.par }
func (v *voidElement) SetAttr(k, val string)    { v.attrs[k] = val }
func (v *voidElement) GetAttr(k string) string  { return v.attrs[k] }
func (v *voidElement) RemoveAttr(k string)      { delete(v.attrs, k) }
func (v *voidElement) SetTextContent(_ string)  {} // no-op for void
func (v *voidElement) TextContent() string      { return "" }
func (v *voidElement) SetInnerHTML(_ string)    {} // no-op for void
func (v *voidElement) InnerHTML() string        { return "" }
func (v *voidElement) Render(w io.Writer) error { return renderElement(v.tag, v.attrs, nil, true, w) }

// ── TextNode ──────────────────────────────────────────────────────────────────

// TextNode represents a plain-text node. Its content is HTML-escaped on render.
type TextNode struct {
	content string
	par     Node
}

func NewTextNode(s string) *TextNode { return &TextNode{content: s} }

func (t *TextNode) node()                   {}
func (t *TextNode) tagName() string         { return "" }
func (t *TextNode) isVoid() bool            { return false }
func (t *TextNode) children() []Node        { return nil }
func (t *TextNode) appendChildNode(_ Node)  {}
func (t *TextNode) Parent() Node            { return t.par }
func (t *TextNode) SetAttr(_, _ string)     {}
func (t *TextNode) GetAttr(_ string) string { return "" }
func (t *TextNode) RemoveAttr(_ string)     {}
func (t *TextNode) SetTextContent(s string) { t.content = s }
func (t *TextNode) TextContent() string     { return t.content }
func (t *TextNode) SetInnerHTML(s string)   { t.content = s }
func (t *TextNode) InnerHTML() string       { return html.EscapeString(t.content) }
func (t *TextNode) Render(w io.Writer) error {
	_, err := io.WriteString(w, html.EscapeString(t.content))
	return err
}

// ── convertXNode: x/net/html → dom.Node ──────────────────────────────────────

func convertXNode(n *xhtml.Node) Node {
	switch n.Type {
	case xhtml.TextNode:
		return NewTextNode(n.Data)
	case xhtml.ElementNode:
		el := &GenericElement{element: newElement(n.Data)}
		for _, a := range n.Attr {
			el.SetAttr(a.Key, a.Val)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			child := convertXNode(c)
			el.appendChildNode(child)
		}
		return el
	default:
		return NewTextNode("")
	}
}

// GenericElement is used exclusively by SetInnerHTML to wrap parsed nodes
// that do not correspond to a specific typed element. It is intentionally
// not exported as a rich API – use the typed constructors for normal work.
type GenericElement struct {
	element
}

func (g *GenericElement) flowContent()     {}
func (g *GenericElement) phrasingContent() {}
func (g *GenericElement) Render(w io.Writer) error {
	return renderElement(g.tag, g.attrs, g.kids, false, w)
}
