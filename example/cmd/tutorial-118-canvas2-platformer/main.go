// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Command tutorial-118-canvas2-platformer is a 2D platformer, which a program of the gostrict subset runs in the
// browser on a canvas, see package game. The game reports a finished level to its backend, which keeps the best
// time of all players and pushes it to every open game, also to those of other windows and users.
package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/example/cmd/tutorial-118-canvas2-platformer/game/messages"
	"go.wdy.de/nago/pkg/gostrict"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas2"
	"go.wdy.de/nago/presentation/ui/program"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed game
var sources embed.FS

// highscores keeps the best time and the channels of all open games, to which it pushes a new best time.
type highscores struct {
	mu       sync.Mutex
	best     messages.Highscore
	channels map[*program.Channel]bool
}

func (h *highscores) push(ch *program.Channel) {
	h.mu.Lock()
	best := h.best
	h.mu.Unlock()
	if best.Holder != "" {
		_ = ch.Send(best)
	}
}

// backend receives the messages of one game.
type backend struct {
	scores *highscores
	ch     *program.Channel
}

// Hello is sent by a game, when it starts, also after a reconnect, so it gets the current highscore.
func (b backend) Hello(ctx context.Context, m messages.Hello) error {
	b.scores.push(b.ch)
	return nil
}

func (b backend) Finished(ctx context.Context, m messages.Finished) error {
	h := b.scores
	h.mu.Lock()
	improved := h.best.Holder == "" || m.Seconds < h.best.Seconds
	if improved {
		h.best = messages.Highscore{Holder: m.Player, Seconds: m.Seconds}
	}
	var all []*program.Channel
	for ch := range h.channels {
		all = append(all, ch)
	}
	h.mu.Unlock()

	if improved {
		for _, ch := range all {
			h.push(ch)
		}
	}
	return nil
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_118")
		cfg.Serve(vuejs.Dist())

		bundle := program.MustCompileFS(sources, "game")
		sprites := cfg.Resource(application.StaticBytes(spriteSheet()))
		scores := &highscores{channels: map[*program.Channel]bool{}}

		cfg.RootView(".", func(wnd core.Window) core.View {
			const id = "platformer"
			ch := program.ChannelOf(wnd, id)
			core.StateOf[bool](wnd, "registered").Init(func() bool {
				scores.mu.Lock()
				scores.channels[ch] = true
				scores.mu.Unlock()
				wnd.AddDestroyObserver(func() {
					scores.mu.Lock()
					delete(scores.channels, ch)
					scores.mu.Unlock()
				}, core.DestroyOnClose)
				return true
			})

			player := "guest"
			if wnd.Subject().Valid() {
				player = wnd.Subject().Name()
			}
			be := backend{scores: scores, ch: ch}
			router := new(gostrict.Router)
			gostrict.Handle(router, be.Hello)
			gostrict.Handle(router, be.Finished)

			return ui.VStack(
				ui.H1("Platformer"),
				ui.Text("Run with the arrow keys or A and D, jump with space. Click into the game first. On a touch screen, tap the left or the right third to run and the middle to jump."),
				canvas2.Canvas(id, bundle).
					Init("Init", player, string(sprites)).
					Animate("Frame").
					Resize("Resize").
					OnMessage(router.Dispatch).
					Frame(ui.Frame{Width: ui.L880, Height: ui.L400, MaxWidth: ui.Full}),
			).Gap(ui.L16).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

// spriteSheet draws the four sprites of 16 by 16 pixels: ground, coin, player and flag.
func spriteSheet() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 64, 16))
	set := func(frame, x, y int, c color.NRGBA) {
		img.SetNRGBA(frame*16+x, y, c)
	}
	for y := range 16 {
		for x := range 16 {
			// ground: grass on top of earth with a few stones
			switch {
			case y < 3:
				set(0, x, y, color.NRGBA{R: 76, G: 175, B: 80, A: 255})
			case (x*7+y*3)%11 == 0:
				set(0, x, y, color.NRGBA{R: 120, G: 110, B: 100, A: 255})
			default:
				set(0, x, y, color.NRGBA{R: 139, G: 90, B: 43, A: 255})
			}

			// coin: a golden disc with a rim
			dx, dy := float64(x)-7.5, float64(y)-7.5
			if r := dx*dx/25 + dy*dy/49; r <= 1 {
				c := color.NRGBA{R: 255, G: 217, B: 61, A: 255}
				if r > 0.6 {
					c = color.NRGBA{R: 214, G: 160, B: 20, A: 255}
				}
				set(1, x, y, c)
			}

			// player: a blue body with eyes, which looks to the right
			if x >= 3 && x <= 12 && y >= 2 {
				c := color.NRGBA{R: 77, G: 150, B: 255, A: 255}
				if y >= 5 && y <= 7 && (x == 8 || x == 11) {
					c = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				}
				if y >= 13 && x >= 5 && x <= 10 && x != 7 && x != 8 {
					c = color.NRGBA{R: 40, G: 60, B: 120, A: 255}
				}
				set(2, x, y, c)
			}

			// flag: a pole with a red flag
			if x == 3 {
				set(3, x, y, color.NRGBA{R: 90, G: 90, B: 90, A: 255})
			}
			if x > 3 && x < 14 && y >= 1 && y <= 7 {
				set(3, x, y, color.NRGBA{R: 255, G: 77, B: 77, A: 255})
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(fmt.Errorf("cannot encode the sprites: %w", err))
	}
	return buf.Bytes()
}
