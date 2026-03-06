// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package dom_test

import (
	"strings"
	"testing"

	"go.wdy.de/nago/pkg/dom"
)

// ── Document ──────────────────────────────────────────────────────────────────

func TestNewDocument_RenderDoctype(t *testing.T) {
	doc := dom.NewDocument()
	doc.SetLang("de")
	doc.SetTitle("Hallo Welt")
	doc.AddMeta(map[string]string{"charset": "UTF-8"})
	doc.AddLink(map[string]string{"rel": "stylesheet", "href": "/style.css"})
	doc.AddScript("/app.js", "")

	html, err := doc.RenderToString()
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, "<!DOCTYPE html>")
	mustContain(t, html, `lang="de"`)
	mustContain(t, html, "<title>Hallo Welt</title>")
	mustContain(t, html, `charset="UTF-8"`)
	mustContain(t, html, `href="/style.css"`)
	mustContain(t, html, `src="/app.js"`)
}

// ── Sectioning & Grouping ─────────────────────────────────────────────────────

func TestBodyStructure(t *testing.T) {
	doc := dom.NewDocument()

	main := dom.NewMain()
	main.SetAttr("id", "main-content")

	h1 := dom.NewH1()
	h1.SetTextContent("Willkommen")
	main.AppendChild(h1)

	p := dom.NewP()
	p.AppendChild(dom.NewTextNode("Ein "))
	bold := dom.NewStrong()
	bold.AppendChild(dom.NewTextNode("fetter"))
	p.AppendChild(bold)
	p.AppendChild(dom.NewTextNode(" Text."))
	main.AppendChild(p)

	doc.Body.AppendChild(main)

	html, err := doc.RenderToString()
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, `id="main-content"`)
	mustContain(t, html, "<h1>Willkommen</h1>")
	mustContain(t, html, "<strong>fetter</strong>")
}

// ── Select type-safety ────────────────────────────────────────────────────────

func TestSelect_OnlyOptionOptgroup(t *testing.T) {
	sel := dom.NewSelect()
	sel.SetAttr("name", "farbe")

	opt1 := dom.NewOption("Rot")
	opt1.SetAttr("value", "red")
	sel.AppendChild(opt1)

	grp := dom.NewOptgroup()
	grp.SetAttr("label", "Grüntöne")
	grp.AppendChild(dom.NewOption("Hellgrün"))
	grp.AppendChild(dom.NewOption("Dunkelgrün"))
	sel.AppendChild(grp)

	html, err := dom.RenderToString(sel)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, `<select name="farbe">`)
	mustContain(t, html, `<option value="red">Rot</option>`)
	mustContain(t, html, `<optgroup label="Grüntöne">`)
	mustContain(t, html, "<option>Hellgrün</option>")
}

// ── Table type-safety ─────────────────────────────────────────────────────────

func TestTable_StrictHierarchy(t *testing.T) {
	table := dom.NewTable()

	thead := dom.NewThead()
	tr := dom.NewTr()
	th := dom.NewTh()
	th.SetTextContent("Spalte 1")
	tr.AppendChild(th)
	thead.AppendChild(tr)
	table.AppendChild(thead)

	tbody := dom.NewTbody()
	tr2 := dom.NewTr()
	td := dom.NewTd()
	td.SetTextContent("Wert")
	tr2.AppendChild(td)
	tbody.AppendChild(tr2)
	table.AppendChild(tbody)

	html, err := dom.RenderToString(table)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, "<table>")
	mustContain(t, html, "<thead>")
	mustContain(t, html, "<th>Spalte 1</th>")
	mustContain(t, html, "<tbody>")
	mustContain(t, html, "<td>Wert</td>")
}

// ── SetInnerHTML / InnerHTML round-trip ───────────────────────────────────────

func TestSetInnerHTML_ReplacesChildren(t *testing.T) {
	div := dom.NewDiv()
	div.SetTextContent("alter Inhalt")
	div.SetInnerHTML(`<p>Neu</p><span>inline</span>`)

	html, err := dom.RenderToString(div)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, "<p>Neu</p>")
	mustContain(t, html, "<span>inline</span>")
	mustNotContain(t, html, "alter Inhalt")
}

func TestInnerHTML_RendersChildren(t *testing.T) {
	div := dom.NewDiv()
	p := dom.NewP()
	p.SetTextContent("Hallo")
	div.AppendChild(p)

	inner := div.InnerHTML()
	if inner != "<p>Hallo</p>" {
		t.Errorf("InnerHTML = %q, want %q", inner, "<p>Hallo</p>")
	}
}

// ── TextContent escaping ──────────────────────────────────────────────────────

func TestTextContent_HTMLEscaping(t *testing.T) {
	span := dom.NewSpan()
	span.SetTextContent("<script>alert('xss')</script>")

	html, err := dom.RenderToString(span)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustNotContain(t, html, "<script>")
	mustContain(t, html, "&lt;script&gt;")
}

// ── Attribute escaping ────────────────────────────────────────────────────────

func TestAttr_Escaping(t *testing.T) {
	a := dom.NewA()
	a.SetAttr("href", `/path?a=1&b=2`)
	a.AppendChild(dom.NewTextNode("Link"))

	html, err := dom.RenderToString(a)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, `href="/path?a=1&amp;b=2"`)
}

// ── GetElementByID / GetElementsByTag ─────────────────────────────────────────

func TestGetElementByID(t *testing.T) {
	doc := dom.NewDocument()

	div := dom.NewDiv()
	div.SetAttr("id", "target")
	div.SetTextContent("gefunden")
	doc.Body.AppendChild(div)

	found := doc.GetElementByID("target")
	if found == nil {
		t.Fatal("GetElementByID returned nil")
	}
	if found.TextContent() != "gefunden" {
		t.Errorf("TextContent = %q, want %q", found.TextContent(), "gefunden")
	}
}

func TestGetElementsByTag(t *testing.T) {
	doc := dom.NewDocument()

	for i := 0; i < 3; i++ {
		p := dom.NewP()
		p.SetTextContent("Absatz")
		doc.Body.AppendChild(p)
	}

	nodes := doc.GetElementsByTag("p")
	if len(nodes) != 3 {
		t.Errorf("GetElementsByTag returned %d nodes, want 3", len(nodes))
	}
}

// ── Void elements ─────────────────────────────────────────────────────────────

func TestVoidElements_NoClosingTag(t *testing.T) {
	br, _ := dom.RenderToString(dom.NewBr())
	mustContain(t, br, "<br>")
	mustNotContain(t, br, "</br>")

	hr, _ := dom.RenderToString(dom.NewHr())
	mustContain(t, hr, "<hr>")
	mustNotContain(t, hr, "</hr>")

	img := dom.NewImg()
	img.SetAttr("src", "/logo.png")
	img.SetAttr("alt", "Logo")
	imgHtml, _ := dom.RenderToString(img)
	mustContain(t, imgHtml, `src="/logo.png"`)
	mustNotContain(t, imgHtml, "</img>")
}

// ── Nested structure ──────────────────────────────────────────────────────────

func TestNestedDivs(t *testing.T) {
	outer := dom.NewDiv()
	outer.SetAttr("class", "outer")
	inner := dom.NewDiv()
	inner.SetAttr("class", "inner")
	p := dom.NewP()
	p.SetTextContent("Inhalt")
	inner.AppendChild(p)
	outer.AppendChild(inner)

	html, err := dom.RenderToString(outer)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, `<div class="outer"><div class="inner"><p>Inhalt</p></div></div>`)
}

// ── Figure / Figcaption ───────────────────────────────────────────────────────

func TestFigure(t *testing.T) {
	fig := dom.NewFigure()
	img := dom.NewImg()
	img.SetAttr("src", "/bild.jpg")
	img.SetAttr("alt", "Ein Bild")
	fig.AppendChild(img)
	cap := dom.NewFigcaption()
	cap.SetTextContent("Bildunterschrift")
	fig.AppendChild(cap)

	html, err := dom.RenderToString(fig)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, "<figure>")
	mustContain(t, html, `src="/bild.jpg"`)
	mustContain(t, html, "<figcaption>Bildunterschrift</figcaption>")
}

// ── Ruby ──────────────────────────────────────────────────────────────────────

func TestRuby(t *testing.T) {
	ruby := dom.NewRuby()
	ruby.AppendText(dom.NewTextNode("漢"))
	rt := dom.NewRt()
	rt.SetTextContent("かん")
	ruby.AppendChild(rt)

	html, err := dom.RenderToString(ruby)
	if err != nil {
		t.Fatalf("RenderToString: %v", err)
	}
	mustContain(t, html, "<ruby>")
	mustContain(t, html, "<rt>かん</rt>")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected output to contain %q\ngot: %s", sub, s)
	}
}

func mustNotContain(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Errorf("expected output NOT to contain %q\ngot: %s", sub, s)
	}
}
