// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package draw uses most of the 2D context of package lib/canvas, so that a test compares the commands of the
// engines and of the bindings.
package draw

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
	"go.wdy.de/nago/pkg/gostrict/std/list"
)

var ctx = canvas.Get()

// Scene is the state of the drawing.
type Scene struct {
	Star   canvas.Path
	Sky    canvas.Gradient
	Sprite canvas.Image
	Tiles  canvas.Pattern
	Stars  list.List[float64]
}

// New creates the paths, styles and images of the scene.
func New() *Scene {
	star := canvas.NewPath()
	for i := range int32(10) {
		r := 20.0
		if i%2 == 1 {
			r = 8
		}
		a := float64(i) * math.Pi / 5
		if i == 0 {
			canvas.MoveTo(star, r*math.Sin(a), -r*math.Cos(a))
			continue
		}
		canvas.LineTo(star, r*math.Sin(a), -r*math.Cos(a))
	}
	canvas.ClosePath(star)

	sky := canvas.NewLinearGradient(ctx, 0, 0, 0, canvas.Height(ctx))
	canvas.AddColorStop(sky, 0, 0x0b1026ff)
	canvas.AddColorStop(sky, 1, 0x2b4c7eff)

	sprite := canvas.LoadImage("/sprites.png")
	tiles := canvas.NewPattern(ctx, sprite, "repeat")
	return &Scene{Star: star, Sky: sky, Sprite: sprite, Tiles: tiles, Stars: list.Of(0.1, 0.35, 0.6, 0.85)}
}

// Frame draws the scene at the time t in seconds.
func Frame(s *Scene, t float64) {
	w, h := canvas.Width(ctx), canvas.Height(ctx)
	canvas.Clear(ctx)
	canvas.FillPaint(ctx, s.Sky)
	canvas.FillRect(ctx, 0, 0, w, h)

	canvas.Save(ctx)
	canvas.GlobalAlpha(ctx, 0.8)
	canvas.FillColor(ctx, 0xffd93dff)
	canvas.Shadow(ctx, 0xffd93d80, 12, 0, 0)
	for _, x := range s.Stars.All() {
		canvas.Save(ctx)
		canvas.Translate(ctx, x*w, h*0.3)
		canvas.Rotate(ctx, t+x)
		canvas.Scale(ctx, 1, 1)
		canvas.FillPath(ctx, s.Star)
		canvas.Restore(ctx)
	}
	canvas.Restore(ctx)

	canvas.BeginPath(ctx)
	canvas.Arc(ctx, w/2, h, 40, math.Pi, 2*math.Pi, false)
	canvas.Ellipse(ctx, 20, 20, 10, 5, 0.5, 0, 2*math.Pi, true)
	canvas.RoundRect(ctx, 5, 5, 30, 20, 4)
	canvas.BezierCurveTo(ctx, 1, 2, 3, 4, 5, 6)
	canvas.QuadraticCurveTo(ctx, 1, 2, 3, 4)
	canvas.ArcTo(ctx, 1, 2, 3, 4, 5)
	canvas.Rect(ctx, 0, h-10, w, 10)
	canvas.FillEvenOdd(ctx)
	canvas.StrokeColor(ctx, 0xffffffff)
	canvas.LineWidth(ctx, 2)
	canvas.LineCap(ctx, "round")
	canvas.LineJoin(ctx, "bevel")
	canvas.MiterLimit(ctx, 4)
	canvas.SetLineDash(ctx, "4 2")
	canvas.LineDashOffset(ctx, t)
	canvas.Stroke(ctx)
	canvas.SetLineDash(ctx, "")

	canvas.FillPaint(ctx, s.Tiles)
	canvas.StrokeRect(ctx, 1, 1, 10, 10)
	canvas.ClearRect(ctx, 2, 2, 3, 3)
	canvas.DrawImage(ctx, s.Sprite, 0, 0)
	canvas.DrawImageScaled(ctx, s.Sprite, 0, 0, 16, 16)
	canvas.DrawImageRegion(ctx, s.Sprite, 0, 0, 8, 8, 10, 10, 32, 32)

	canvas.Font(ctx, "bold 16px sans-serif")
	canvas.TextAlign(ctx, "left")
	canvas.TextBaseline(ctx, "top")
	canvas.Direction(ctx, "ltr")
	width, _, _ := canvas.MeasureText(ctx, "gostrict")
	canvas.FillText(ctx, "gostrict", w-width-8, 8)
	canvas.FillTextMax(ctx, "max", 0, 0, 20)
	canvas.StrokeText(ctx, "outline", 0, 30)
	canvas.StrokeTextMax(ctx, "outline", 0, 40, 30)

	canvas.SetTransform(ctx, 1, 0, 0, 1, 10, 10)
	canvas.Transform(ctx, 1, 0, 0, 1, 1, 1)
	canvas.ResetTransform(ctx)
	canvas.CompositeOperation(ctx, "source-over")
	canvas.ImageSmoothing(ctx, false)
	canvas.StrokeStyle(ctx, "red")
	canvas.FillStyle(ctx, "blue")
	canvas.StrokePath(ctx, s.Star)
	canvas.ClipPath(ctx, s.Star)
	canvas.Clip(ctx)
	canvas.Fill(ctx)
	if canvas.IsPointInPath(ctx, s.Star, 1, 1) || canvas.IsPointInStroke(ctx, s.Star, 1, 1) {
		canvas.FillRect(ctx, 0, 0, 1, 1)
	}
}
