// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command tutorial-119-canvas2-photobook edits a double page of a photo book on a canvas. The editor is a program
// of the gostrict subset, which runs in the browser, see package book, so dragging, scaling, rotating and snapping
// react within the frame. The backend receives the uploads, serves their images and stores the layout, which the
// editor saves after each change, and pushes it to the editors of all other windows.
package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/example/cmd/tutorial-119-canvas2-photobook/book/messages"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas2"
	"go.wdy.de/nago/presentation/ui/program"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed book
var sources embed.FS

const photoPath = "/api/tutorial-119/photos/"

// store keeps the uploaded photos and the layout of the double page in memory and the channels of the open
// editors.
type store struct {
	mu       sync.Mutex
	photos   map[string][]byte
	layout   []messages.Photo
	channels map[*program.Channel]bool
}

func (s *store) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	buf, ok := s.photos[strings.TrimPrefix(r.URL.Path, photoPath)]
	s.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(buf))
}

func (s *store) layoutMessage() messages.Layout {
	s.mu.Lock()
	defer s.mu.Unlock()
	photos := list.New[messages.Photo]()
	photos.Append(s.layout...)
	return messages.Layout{Photos: photos}
}

// backend receives the messages of one editor.
type backend struct {
	store *store
	ch    *program.Channel
}

// Hello is sent by the editor, when it starts, also after a reconnect, so that it shows the current layout.
func (b backend) Hello(ctx context.Context, m messages.Hello) error {
	return b.ch.Send(b.store.layoutMessage())
}

func (b backend) Save(ctx context.Context, m messages.Save) error {
	s := b.store
	s.mu.Lock()
	s.layout = s.layout[:0]
	for p := range m.Photos.Values() {
		s.layout = append(s.layout, p)
	}
	var others []*program.Channel
	for ch := range s.channels {
		if ch != b.ch {
			others = append(others, ch)
		}
	}
	s.mu.Unlock()

	msg := s.layoutMessage()
	for _, ch := range others {
		_ = ch.Send(msg)
	}
	return nil
}

// upload keeps the images and puts them on the double page of the editor.
func (b backend) upload(files []core.File) {
	for _, f := range files {
		var buf bytes.Buffer
		if _, err := f.Transfer(&buf); err != nil {
			continue
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(buf.Bytes()))
		if err != nil {
			continue
		}

		s := b.store
		s.mu.Lock()
		id := fmt.Sprintf("%d", len(s.photos)+1)
		s.photos[id] = buf.Bytes()
		s.mu.Unlock()

		_ = b.ch.Send(messages.Add{URL: photoPath + id, Width: float64(cfg.Width), Height: float64(cfg.Height)})
	}
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_119")
		cfg.Serve(vuejs.Dist())

		bundle := program.MustCompileFS(sources, "book")
		s := &store{photos: map[string][]byte{}, channels: map[*program.Channel]bool{}}
		// the endpoint is mounted, so it serves all paths below it
		cfg.HandleFunc(strings.TrimSuffix(photoPath, "/"), s.serve)

		cfg.RootView(".", func(wnd core.Window) core.View {
			const id = "photobook"
			ch := program.ChannelOf(wnd, id)
			core.StateOf[bool](wnd, "registered").Init(func() bool {
				s.mu.Lock()
				s.channels[ch] = true
				s.mu.Unlock()
				wnd.AddDestroyObserver(func() {
					s.mu.Lock()
					delete(s.channels, ch)
					s.mu.Unlock()
				}, core.DestroyOnClose)
				return true
			})
			be := backend{store: s, ch: ch}
			router := new(gostrict.Router)
			gostrict.Handle(router, be.Hello)
			gostrict.Handle(router, be.Save)

			return ui.VStack(
				ui.H1("Photo book"),
				ui.HStack(
					ui.PrimaryButton(func() {
						wnd.ImportFiles(core.ImportFilesOptions{
							Multiple:         true,
							AllowedMimeTypes: []string{"image/jpeg", "image/png"},
							OnCompletion:     be.upload,
						})
					}).Title("Add photos"),
					ui.Text("Drag a photo to move it, its corner to scale it and the handle above it to rotate it. Delete removes it, ] and [ change its layer."),
				).Gap(ui.L16),
				canvas2.Canvas(id, bundle).
					Init("Init").
					Animate("Frame").
					Resize("Resize").
					OnMessage(router.Dispatch).
					Frame(ui.Frame{Width: ui.Full, Height: ui.L560}),
			).Gap(ui.L16).Frame(ui.Frame{Width: ui.Full}.MatchScreen())
		})
	}).Run()
}
