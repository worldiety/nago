// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	_ "embed"

	"go.wdy.de/nago/application"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/canvas"
	"go.wdy.de/nago/web/vuejs"
)

//go:embed morty_vanilla.png
var morty application.StaticBytes

const (
	handleSize = 10.0 // Größe der Anfasser in Pixeln
	minSize    = 30.0 // Mindestgröße des Bildes in Pixeln
)

// dragMode beschreibt den aktuellen Drag-Zustand.
type dragMode int

const (
	dragNone dragMode = iota
	dragMove          // Bild verschieben
	dragTL            // Skalierung: oben-links
	dragTR            // Skalierung: oben-rechts
	dragBL            // Skalierung: unten-links
	dragBR            // Skalierung: unten-rechts
)

// editorState hält die gesamte veränderliche Zustandsinformation.
type editorState struct {
	x, y           float64 // Position des Bildes (oben-links)
	w, h           float64 // Größe des Bildes
	mode           dragMode
	startX, startY float64 // Zeiger-Position beim PointerDown
	origX, origY   float64 // Bild-Position beim PointerDown
	origW, origH   float64 // Bild-Größe beim PointerDown
}

// inHandle prüft, ob der Zeiger (ex, ey) innerhalb des Anfassers bei (hx, hy) liegt.
func inHandle(ex, ey, hx, hy float64) bool {
	half := handleSize / 2
	return ex >= hx-half && ex <= hx+half && ey >= hy-half && ey <= hy+half
}

// inRect prüft, ob der Zeiger innerhalb des Bild-Rechtecks liegt.
func inRect(s *editorState, ex, ey float64) bool {
	return ex >= s.x && ex <= s.x+s.w && ey >= s.y && ey <= s.y+s.h
}

func main() {
	application.Configure(func(cfg *application.Configurator) {
		cfg.SetApplicationID("de.worldiety.tutorial_94")
		cfg.Serve(vuejs.Dist())

		mortyUrl := cfg.Resource(morty)

		cfg.RootView(".", func(wnd core.Window) core.View {
			const myCanvas = "morty-editor"

			canvasCtx := canvas.Context2D(wnd, myCanvas)

			state := &editorState{
				x: 120, y: 80,
				w: 160, h: 160,
			}

			wnd.AddInputListener(myCanvas, func(e core.InputEvent) {
				// InputEventInvalidate und andere → Initialbild zeichnen
				canvasCtx.LoadImage(1, mortyUrl)

				cx, cy, w, h := state.x, state.y, state.w, state.h

				switch e.Type {
				case core.InputEventPointerDown:
					// Anfasser zuerst prüfen (höhere Priorität als das Bild-Innere)
					switch {
					case inHandle(e.X, e.Y, cx, cy):
						state.mode = dragTL
					case inHandle(e.X, e.Y, cx+w, cy):
						state.mode = dragTR
					case inHandle(e.X, e.Y, cx, cy+h):
						state.mode = dragBL
					case inHandle(e.X, e.Y, cx+w, cy+h):
						state.mode = dragBR
					case inRect(state, e.X, e.Y):
						state.mode = dragMove
					default:
						state.mode = dragNone
					}
					state.startX, state.startY = e.X, e.Y
					state.origX, state.origY = state.x, state.y
					state.origW, state.origH = state.w, state.h

				case core.InputEventPointerMove:
					if state.mode == dragNone {
						return
					}
					dx := e.X - state.startX
					dy := e.Y - state.startY
					switch state.mode {
					case dragMove:
						state.x = state.origX + dx
						state.y = state.origY + dy
					case dragTL:
						nw, nh := state.origW-dx, state.origH-dy
						if nw >= minSize && nh >= minSize {
							state.x, state.y = state.origX+dx, state.origY+dy
							state.w, state.h = nw, nh
						}
					case dragTR:
						nw, nh := state.origW+dx, state.origH-dy
						if nw >= minSize && nh >= minSize {
							state.y = state.origY + dy
							state.w, state.h = nw, nh
						}
					case dragBL:
						nw, nh := state.origW-dx, state.origH+dy
						if nw >= minSize && nh >= minSize {
							state.x = state.origX + dx
							state.w, state.h = nw, nh
						}
					case dragBR:
						nw, nh := state.origW+dx, state.origH+dy
						if nw >= minSize && nh >= minSize {
							state.w, state.h = nw, nh
						}
					}
					redraw(canvasCtx, state)

				case core.InputEventPointerUp, core.InputEventPointerCancel:
					state.mode = dragNone

				default:

					redraw(canvasCtx, state)
				}
			})

			return ui.VStack(
				ui.Text("Morty-Canvas-Editor").Font(ui.Font{Size: "1.2rem", Weight: ui.HeadlineAndTitleFontWeight}),
				ui.Text("Bild verschieben: ins Bild klicken und ziehen  •  Skalieren: Eck-Anfasser (■) ziehen").
					Font(ui.Font{Size: "0.85rem"}),
				canvas.Canvas(myCanvas).Frame(ui.Frame{Width: "600px", Height: "480px"}),
			).Gap(ui.L8).Frame(ui.Frame{}.MatchScreen())
		})
	}).Run()
}

// redraw löscht den Canvas und zeichnet Bild, Auswahlrahmen und Eck-Anfasser.
func redraw(ctx canvas.TContext2D, s *editorState) {
	ctx.NewList(1)

	// Canvas leeren
	ctx.ClearRect(0, 0, 4000, 4000)

	// Morty-Bild skaliert zeichnen (5-Argument-Form via DrawImage2 mit sx=sy=sw=sh=0)
	ctx.DrawImage2(1, s.x, s.y, s.w, s.h, 0, 0, 0, 0)

	// Auswahlrahmen
	ctx.StrokeColor("#1565C0")
	ctx.LineWidth(1.5)
	ctx.StrokeRect(s.x, s.y, s.w, s.h)

	// 4 Eck-Anfasser
	drawHandle(ctx, s.x, s.y)         // oben-links
	drawHandle(ctx, s.x+s.w, s.y)     // oben-rechts
	drawHandle(ctx, s.x, s.y+s.h)     // unten-links
	drawHandle(ctx, s.x+s.w, s.y+s.h) // unten-rechts

	ctx.EndList()
	ctx.CallList(1)
}

// drawHandle zeichnet einen ausgefüllten quadratischen Anfasser zentriert bei (hx, hy).
func drawHandle(ctx canvas.TContext2D, hx, hy float64) {
	half := handleSize / 2
	ctx.FillColor("#1565C0")
	ctx.FillRect(hx-half, hy-half, handleSize, handleSize)
	ctx.StrokeColor("#ffffff")
	ctx.LineWidth(1)
	ctx.StrokeRect(hx-half, hy-half, handleSize, handleSize)
}
