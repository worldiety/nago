// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package markdown_test

import (
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/markdown"
)

// The heading of a rendered document, e.g. of a file preview, must not rename the browser tab, unlike the heading
// of the page itself.
func TestHeadingDoesNotSetTheWindowTitle(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.markdowntest")
		cfg.RootView(".", func(wnd core.Window) core.View {
			return ui.VStack(
				ui.H1("Page"),
				markdown.Render(markdown.Options{Window: wnd}, []byte("# Document\n\ntext")),
			)
		})
	})

	w := app.Open(t, nil, ".")
	w.Find(nagotest.Text("Document")).Exactly(1)
	titles := w.FindAll(nagotest.Type[*proto.WindowTitle]())
	titles.Exactly(1)
	if got := string(titles.Node().Component.(*proto.WindowTitle).Value); got != "Page" {
		t.Fatalf("expected the title of the page, got %q", got)
	}
}

// The plain texts become labels of the given component, e.g. of a checkbox, but links stay links.
func TestLabelFor(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.markdowntest")
		cfg.RootView(".", func(wnd core.Window) core.View {
			return markdown.Render(markdown.Options{Window: wnd, LabelFor: "box"}, []byte("I accept the [terms](/terms) now."))
		})
	})

	w := app.Open(t, nil, ".")
	for _, text := range []string{"I accept the ", " now."} {
		n := w.Find(nagotest.Text(text)).Node().Component.(*proto.TextView)
		if n.LabelFor != "box" {
			t.Fatalf("%q: expected a label of the box, got %q", text, n.LabelFor)
		}
	}

	if n := w.Find(nagotest.Text("terms")).Node().Component.(*proto.TextView); n.LabelFor != "" {
		t.Fatalf("a link must not be a label, got %q", n.LabelFor)
	}
}
