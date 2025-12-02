// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
)

type Options struct {
	Window        core.Window // the Window is required for links to work properly
	RichText      bool
	TrimParagraph bool
}

// RichText is a little helper factory to get some markdown text into a RichText.
func RichText(value string) ui.TRichText {
	return markdown(Options{RichText: true, TrimParagraph: true}, []byte(value))
}

func markdown(opts Options, source []byte) ui.TRichText {
	md := newMd()
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return ui.RichText(err.Error())
	}

	b := buf.Bytes()
	if opts.TrimParagraph {
		b = bytes.TrimSpace(b)
		if bytes.HasPrefix(b, []byte("<p>")) {
			b = b[3:]
		}

		if bytes.HasSuffix(b, []byte("</p>")) {
			b = b[:len(b)-4]
		}
	}

	return ui.RichText(string(b)).FullWidth()
}

func newMd() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

// Render parses the given source as a markdown dialect and interprets it as views.
func Render(opts Options, source []byte) core.View {
	if opts.RichText {
		return markdown(opts, source)
	}

	md := newMd()

	r := renderer{}
	node := md.Parser().Parse(text.NewReader(source))

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := n.(type) {
		case *ast.Document:
			if entering {
				r.Push(&mutDocument{})
			} // intentionally never pop that, so that at the end, we just have that on stack
		case *ast.Heading:
			if entering {
				r.Push(&mutHeading{level: n.Level})
			} else {
				r.Pop()
			}
		case *ast.Text:
			if entering {
				r.Push(&mutText{value: string(n.Value(source)), linebreak: n.HardLineBreak()})
			} else {
				r.Pop()
			}
		case *ast.Paragraph:
			if entering {
				r.Push(&mutTextLayout{})
			} else {
				r.Pop()
			}

		case *ast.Link:
			if entering {
				r.Push(&mutLink{href: string(n.Destination)})
			} else {
				r.Pop()
			}
		default:
			//fmt.Printf("not implemented %T\n", n)
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		panic(err) // unreachable
	}

	return r.Top().Render(opts.Window)
}

// renderer manages a stack of mutView while traversing the AST.
type renderer struct {
	stack []mutView
}

func (r *renderer) Pop() mutView {
	t := r.stack[len(r.stack)-1]
	r.stack[len(r.stack)-1] = nil
	r.stack = r.stack[:len(r.stack)-1]
	return t
}

func (r *renderer) Push(v mutView) {
	top := r.Top()
	if top != nil {
		r.Top().Add(v)
	}
	r.stack = append(r.stack, v)
}

func (r *renderer) Top() mutView {
	if len(r.stack) == 0 {
		return nil
	}

	return r.stack[len(r.stack)-1]
}

// mutView is a lightweight interface for building intermediate nodes.
type mutView interface {
	Add(view mutView)
	Render(wnd core.Window) core.View
}

// mutDocument corresponds to a markdown document root.
type mutDocument struct {
	views []mutView
}

func (c *mutDocument) Add(view mutView) {
	c.views = append(c.views, view)
}

func (c *mutDocument) Render(wnd core.Window) core.View {
	var tmp []core.View
	for _, view := range c.views {
		tmp = append(tmp, view.Render(wnd))
	}

	return ui.VStack(tmp...).Alignment(ui.Leading).Gap(ui.L16).FullWidth()
}

// mutHeading represents a markdown heading (only plain text supported).
type mutHeading struct {
	level int
	views []mutView
}

func (c *mutHeading) Add(view mutView) {
	c.views = append(c.views, view)
}

func (c *mutHeading) Render(wnd core.Window) core.View {
	// this is a limitation of nago: the heading type just accepts plain text
	buf := strings.Builder{}
	for _, view := range c.views {
		if t, ok := view.(*mutText); ok {
			buf.WriteString(t.value)
		}
	}

	return ui.Heading(c.level, buf.String())
}

// mutText is a text leaf node with optional hard line break.
type mutText struct {
	value     string
	linebreak bool
}

func (c *mutText) Add(view mutView) {
	panic("leaf type cannot contain others")
}

func (c *mutText) Render(wnd core.Window) core.View {
	return ui.Text(c.value).LineBreak(c.linebreak)
}

// mutTextLayout groups inline text fragments into a paragraph.
type mutTextLayout struct {
	views []mutView
}

func (c *mutTextLayout) Add(view mutView) {
	c.views = append(c.views, view)
}

func (c *mutTextLayout) Render(wnd core.Window) core.View {
	var tmp []core.View
	for _, view := range c.views {
		tmp = append(tmp, view.Render(wnd))
	}
	return ui.TextLayout(tmp...)
}

// mutLink represents a markdown link, limited to plain text as label.
type mutLink struct {
	views []mutView
	href  string
}

func (c *mutLink) Add(view mutView) {
	c.views = append(c.views, view)
}

func (c *mutLink) Render(wnd core.Window) core.View {
	// this is a limitation of nago: the link type just accepts plain text
	buf := strings.Builder{}
	for _, view := range c.views {
		if t, ok := view.(*mutText); ok {
			buf.WriteString(t.value)
		}
	}

	return ui.Link(wnd, buf.String(), c.href, "_blank")
}
