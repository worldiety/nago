// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package uicompletion

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.wdy.de/nago/application/ai/completion"
	"go.wdy.de/nago/application/ai/file"
	"go.wdy.de/nago/pkg/std/async"
	"go.wdy.de/nago/presentation/core"
)

// shotWindow answers Screenshot with a canned result and records the requested options. All other methods of
// core.Window panic through the nil embedding, which proves the tool needs nothing else.
type shotWindow struct {
	core.Window
	shot  core.Screenshot
	err   error
	never bool
	got   core.ScreenshotOptions
}

func (w *shotWindow) Path() core.NavigationPath { return "books" }

func (w *shotWindow) Screenshot(opts core.ScreenshotOptions) *async.Future[core.Screenshot] {
	w.got = opts
	var fut async.Future[core.Screenshot]
	if !w.never {
		fut.Set(w.shot, w.err)
	}
	return &fut
}

func invokeContent(t *testing.T, tool completion.Tool, args string) ([]completion.Content, error) {
	t.Helper()
	if tool.InvokeContent == nil {
		t.Fatalf("screen tool must return content blocks")
	}
	return tool.InvokeContent(nil, json.RawMessage(args))
}

func TestScreenTool_SnapshotOnlyByDefault(t *testing.T) {
	wnd := &shotWindow{shot: core.Screenshot{Snapshot: `- button "Ausleihen"`}}
	tool := ScreenTool(wnd, ScreenToolOptions{})

	if tool.Def.Name != "inspect_screen" {
		t.Fatalf("unexpected name %q", tool.Def.Name)
	}

	content, err := invokeContent(t, tool, `{}`)
	if err != nil {
		t.Fatal(err)
	}

	if wnd.got.Image || !wnd.got.Snapshot {
		t.Fatalf("default must request the snapshot only, got %#v", wnd.got)
	}

	if len(content) != 1 {
		t.Fatalf("expected text only, got %#v", content)
	}

	txt := content[0].(completion.Text).Text
	if !strings.Contains(txt, `- button "Ausleihen"`) || !strings.Contains(txt, "Route: books") {
		t.Fatalf("unexpected text: %q", txt)
	}
}

func TestScreenTool_ImageOnRequest(t *testing.T) {
	wnd := &shotWindow{shot: core.Screenshot{PNG: []byte("png"), Width: 10, Height: 20, Snapshot: "- text: x"}}
	tool := ScreenTool(wnd, ScreenToolOptions{MaxEdge: 800})

	content, err := invokeContent(t, tool, `{"image":true,"selector":" #main "}`)
	if err != nil {
		t.Fatal(err)
	}

	if !wnd.got.Image || wnd.got.Selector != "#main" || wnd.got.MaxEdge != 800 {
		t.Fatalf("unexpected options %#v", wnd.got)
	}

	if len(content) != 2 {
		t.Fatalf("expected text and image, got %#v", content)
	}

	media, ok := content[1].(completion.Media)
	if !ok || media.MimeType != file.PNG || string(media.Source.Data) != "png" {
		t.Fatalf("unexpected media %#v", content[1])
	}

	if !strings.Contains(content[0].(completion.Text).Text, "10x20") {
		t.Fatalf("text should announce the image: %#v", content[0])
	}
}

func TestScreenTool_DisableImageHidesTheOption(t *testing.T) {
	wnd := &shotWindow{shot: core.Screenshot{Snapshot: "- text: x"}}
	tool := ScreenTool(wnd, ScreenToolOptions{DisableImage: true})

	if strings.Contains(string(tool.Def.Schema), `"image"`) {
		t.Fatalf("schema must not offer image: %s", tool.Def.Schema)
	}

	// a model ignoring the schema must still not get an image
	if _, err := invokeContent(t, tool, `{"image":true}`); err != nil {
		t.Fatal(err)
	}

	if wnd.got.Image {
		t.Fatalf("image must never be requested when disabled")
	}
}

func TestScreenTool_ErrorAndTimeout(t *testing.T) {
	wnd := &shotWindow{err: errors.New("no element matches selector")}
	if _, err := invokeContent(t, ScreenTool(wnd, ScreenToolOptions{}), `{}`); err == nil || !strings.Contains(err.Error(), "no element") {
		t.Fatalf("expected the frontend error, got %v", err)
	}

	wnd = &shotWindow{never: true}
	_, err := invokeContent(t, ScreenTool(wnd, ScreenToolOptions{Timeout: 10 * time.Millisecond}), `{}`)
	if !errors.Is(err, ErrScreenshotTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestScreenTool_TruncatesHugeSnapshots(t *testing.T) {
	wnd := &shotWindow{shot: core.Screenshot{Snapshot: strings.Repeat("x", maxScreenSnapshotRunes+10)}}
	content, err := invokeContent(t, ScreenTool(wnd, ScreenToolOptions{}), `{}`)
	if err != nil {
		t.Fatal(err)
	}

	txt := content[0].(completion.Text).Text
	if !strings.Contains(txt, "[truncated") || strings.Count(txt, "x") > maxScreenSnapshotRunes+5 {
		t.Fatalf("snapshot not truncated")
	}
}
