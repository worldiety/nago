// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ui_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/nagotest"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/engine/gojaengine"
	"go.wdy.de/nago/pkg/gostrict/engine/vm"
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
)

// clipboardProgram is clipboardView as a program: the first verb is what it does while it renders, the second what
// the click does.
const clipboardProgram = `package v

import (
	"go.wdy.de/nago/pkg/gostrict/lib/ui"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

func Render(wnd ui.Window) ui.View {
	n := ui.IntStateOf(wnd, "n")
	%s
	return ui.VStack(ui.Text("clicks="+text.FromInt(n.Get())), ui.PrimaryButton(func() {
		n.Set(n.Get() + 1)
		%s
	}).Title("copy"))
}
`

// clipboardView writes s while it renders or by the click of its button, and counts the clicks.
func clipboardView(wnd ui.Window, rendering bool, s string) ui.View {
	n := ui.IntStateOf(wnd, "n")
	if rendering {
		ui.ClipboardSetText(wnd, s)
	}
	return ui.VStack(ui.Text("clicks="+strconv.Itoa(int(n.Get()))), ui.PrimaryButton(func() {
		n.Set(n.Get() + 1)
		if !rendering {
			ui.ClipboardSetText(wnd, s)
		}
	}).Title("copy"))
}

// clipboardWritten are the texts, which the window has sent to the browser, in their order.
func clipboardWritten(w *nagotest.Window) []string {
	var res []string
	for _, e := range w.Events() {
		if c, ok := e.(*proto.ClipboardWriteTextRequested); ok {
			res = append(res, string(c.Text))
		}
	}
	return res
}

// TestClipboardSetText writes through the function natively, in goja and in the vm: a click and the rendering write
// the unchanged text as one event each to the browser, an empty text sends no event, the counter rows show the same
// view without the click and with a text, which is not empty. The number of renderings is not part of the function,
// so equal neighbours of the rendering rows count once.
func TestClipboardSetText(t *testing.T) {
	const special = "line 1\n\"ü€\"\t😀 <b>&amp;</b>  "
	long := strings.Repeat("0123456789", 20000)
	rows := []struct {
		name      string
		rendering bool
		text      string
		clicks    int
		want      []string
	}{
		{name: "a click writes", text: "copied", clicks: 1, want: []string{"copied"}},
		{name: "without the click nothing is written", text: "copied"},
		{name: "each click writes again", text: "copied", clicks: 2, want: []string{"copied", "copied"}},
		{name: "a click with an empty text writes nothing", text: "", clicks: 1},
		{name: "one blank is not empty", text: " ", clicks: 1, want: []string{" "}},
		{name: "the text stays as it is", text: special, clicks: 1, want: []string{special}},
		{name: "there is no limit of the length", text: long, clicks: 1, want: []string{long}},
		{name: "rendering writes without a click", rendering: true, text: "rendered", want: []string{"rendered"}},
		{name: "rendering an empty text writes nothing", rendering: true, text: ""},
	}

	compile := func(t *testing.T, rendering bool, s string) *gostrict.Bundle {
		t.Helper()
		call := "ui.ClipboardSetText(wnd, " + strconv.Quote(s) + ")"
		if rendering {
			return compileView(t, fmt.Sprintf(clipboardProgram, call, ""))
		}
		return compileView(t, fmt.Sprintf(clipboardProgram, "", call))
	}
	engines := map[string]gostrict.Engine{"goja": &gojaengine.Engine{}, "vm": &vm.Engine{}, "native": nil}

	for _, r := range rows {
		for name, e := range engines {
			t.Run(name+"/"+r.name, func(t *testing.T) {
				render := func(wnd core.Window) core.View { return ui.Core(clipboardView(ui.WindowOf(wnd), r.rendering, r.text)) }
				if e != nil {
					in, err := e.Load(context.Background(), compile(t, r.rendering, r.text), ui.Host())
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = in.Close() })
					render = func(wnd core.Window) core.View { return ui.Render(context.Background(), in, "Render", wnd) }
				}
				app := nagotest.New(t, func(cfg *application.Configurator) {
					cfg.SetApplicationID("de.worldiety.gostrict.clipboard")
					cfg.RootView("page", render)
				})
				w := app.Open(t, nil, "page")

				for i := range r.clicks {
					w.Click(w.Find(nagotest.Text("copy")))
					// the click has reached its function, also if it wrote nothing
					w.Find(nagotest.Text("clicks=" + strconv.Itoa(i+1)))
				}

				got := clipboardWritten(w)
				if r.rendering {
					got = slices.Compact(got)
				}
				if !slices.Equal(got, r.want) {
					t.Errorf("written %.200q, want %.200q", got, r.want)
				}
				last := ""
				if len(r.want) > 0 {
					last = r.want[len(r.want)-1]
				}
				if c := w.Clipboard(); c != last {
					t.Errorf("Clipboard() = %.200q, want %.200q", c, last)
				}
			})
		}
	}
}

// TestClipboardSetTextEmptyKeepsText shows in one window, that an empty text leaves the text written before and that
// a text written afterwards replaces it.
func TestClipboardSetTextEmptyKeepsText(t *testing.T) {
	app := nagotest.New(t, func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.gostrict.clipboard.keep")
		cfg.RootView("page", func(wnd core.Window) core.View {
			w := ui.WindowOf(wnd)
			return ui.Core(ui.VStack(
				ui.PrimaryButton(func() { ui.ClipboardSetText(w, "old") }).Title("old"),
				ui.PrimaryButton(func() { ui.ClipboardSetText(w, "") }).Title("empty"),
				ui.PrimaryButton(func() { ui.ClipboardSetText(w, "new") }).Title("new"),
			))
		})
	})
	w := app.Open(t, nil, "page")

	for _, step := range []struct {
		click string
		want  []string
	}{
		{"old", []string{"old"}},
		{"empty", []string{"old"}},
		{"new", []string{"old", "new"}},
		{"empty", []string{"old", "new"}},
	} {
		w.Click(w.Find(nagotest.Text(step.click)))
		if got := clipboardWritten(w); !slices.Equal(got, step.want) {
			t.Errorf("after %s: written %q, want %q", step.click, got, step.want)
		}
		if c := w.Clipboard(); c != step.want[len(step.want)-1] {
			t.Errorf("after %s: Clipboard() = %q, want %q", step.click, c, step.want[len(step.want)-1])
		}
	}
}
