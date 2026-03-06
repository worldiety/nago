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

// Document represents a complete HTML5 document with a fixed <html>, <head>
// and <body> structure.
type Document struct {
	Head *Head
	Body *Body
	// lang sets the lang attribute on the <html> element
	lang  string
	attrs map[string]string
}

// NewDocument creates a new Document with empty <head> and <body>.
func NewDocument() *Document {
	return &Document{
		Head:  NewHead(),
		Body:  NewBody(),
		attrs: make(map[string]string),
	}
}

// ── Node interface compliance ─────────────────────────────────────────────────

func (d *Document) node()                   {}
func (d *Document) tagName() string         { return "#document" }
func (d *Document) isVoid() bool            { return false }
func (d *Document) appendChildNode(_ Node)  {}
func (d *Document) children() []Node        { return []Node{d.Head, d.Body} }
func (d *Document) Parent() Node            { return nil }
func (d *Document) SetAttr(k, v string)     { d.attrs[k] = v }
func (d *Document) GetAttr(k string) string { return d.attrs[k] }
func (d *Document) RemoveAttr(k string)     { delete(d.attrs, k) }
func (d *Document) SetTextContent(_ string) {}
func (d *Document) TextContent() string     { return "" }
func (d *Document) SetInnerHTML(_ string)   {}
func (d *Document) InnerHTML() string       { return "" }

// Render writes the full HTML5 document including <!DOCTYPE html> to w.
func (d *Document) Render(w io.Writer) error {
	if _, err := io.WriteString(w, "<!DOCTYPE html>\n"); err != nil {
		return err
	}
	// <html lang="...">
	htmlAttrs := make(map[string]string)
	for k, v := range d.attrs {
		htmlAttrs[k] = v
	}
	if d.lang != "" {
		htmlAttrs["lang"] = d.lang
	}
	if _, err := io.WriteString(w, "<html"); err != nil {
		return err
	}
	for k, v := range htmlAttrs {
		if _, err := io.WriteString(w, ` `+k+`="`+v+`"`); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	if err := d.Head.Render(w); err != nil {
		return err
	}
	if err := d.Body.Render(w); err != nil {
		return err
	}
	_, err := io.WriteString(w, "</html>")
	return err
}

// RenderToString renders the document into a string.
func (d *Document) RenderToString() (string, error) {
	var b bytes.Buffer
	if err := d.Render(&b); err != nil {
		return "", err
	}
	return b.String(), nil
}

// ── Language ──────────────────────────────────────────────────────────────────

// SetLang sets the lang attribute on the <html> element.
func (d *Document) SetLang(lang string) { d.lang = lang }

// Lang returns the lang attribute of the <html> element.
func (d *Document) Lang() string { return d.lang }

// ── Head helpers ──────────────────────────────────────────────────────────────

// SetTitle sets or replaces the <title> element in <head>.
func (d *Document) SetTitle(text string) {
	// remove existing title nodes
	kids := d.Head.kids[:0]
	for _, k := range d.Head.kids {
		if k.tagName() != "title" {
			kids = append(kids, k)
		}
	}
	d.Head.kids = append(kids, NewTitle(text))
}

// AddMeta appends a <meta> element with the given attributes to <head>.
func (d *Document) AddMeta(attrs map[string]string) {
	m := NewMeta()
	for k, v := range attrs {
		m.SetAttr(k, v)
	}
	d.Head.AppendChild(m)
}

// AddLink appends a <link> element with the given attributes to <head>.
func (d *Document) AddLink(attrs map[string]string) {
	l := NewLink()
	for k, v := range attrs {
		l.SetAttr(k, v)
	}
	d.Head.AppendChild(l)
}

// AddScript appends a <script> element to <head>.
// Pass src="" to create an inline script; set integrity="" to omit it.
func (d *Document) AddScript(src, integrity string) {
	s := NewScript()
	if src != "" {
		s.SetAttr("src", src)
	}
	if integrity != "" {
		s.SetAttr("integrity", integrity)
		s.SetAttr("crossorigin", "anonymous")
	}
	d.Head.AppendChild(s)
}

// AddStyle appends a <style> element with CSS content to <head>.
func (d *Document) AddStyle(css string) {
	d.Head.AppendChild(NewStyle(css))
}

// ── DOM search ────────────────────────────────────────────────────────────────

// GetElementByID performs a depth-first search and returns the first node
// whose "id" attribute matches the given id, or nil if not found.
func (d *Document) GetElementByID(id string) Node {
	if found := getElementById(d.Head, id); found != nil {
		return found
	}
	return getElementById(d.Body, id)
}

// GetElementsByTag performs a depth-first search and returns all nodes
// whose tag name matches the given tag (case-insensitive).
func (d *Document) GetElementsByTag(tag string) []Node {
	var result []Node
	getElementsByTag(d.Head, tag, &result)
	getElementsByTag(d.Body, tag, &result)
	return result
}

// ── helpers ───────────────────────────────────────────────────────────────────

func getElementById(n Node, id string) Node {
	if n.GetAttr("id") == id {
		return n
	}
	for _, child := range n.children() {
		if found := getElementById(child, id); found != nil {
			return found
		}
	}
	return nil
}

func getElementsByTag(n Node, tag string, result *[]Node) {
	if n.tagName() == tag {
		*result = append(*result, n)
	}
	for _, child := range n.children() {
		getElementsByTag(child, tag, result)
	}
}

// ── Standalone render helpers ─────────────────────────────────────────────────

// Render writes any Node to the given io.Writer.
func Render(n Node, w io.Writer) error { return n.Render(w) }

// RenderToString renders any Node to a string.
func RenderToString(n Node) (string, error) {
	var b bytes.Buffer
	if err := n.Render(&b); err != nil {
		return "", err
	}
	return b.String(), nil
}
