// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package book is the editor of a double page of a photo book, a program of the gostrict subset, which runs in the
// browser and draws on a 2D canvas. The user drags a photo to move it, its lower right corner to scale it and the
// handle above it to rotate it. The edges and centers of a moving photo snap to the pages and to the other photos.
// Delete removes the selected photo, ] and [ move it a layer up or down, the arrow keys nudge it.
package book

import (
	"math"

	"go.wdy.de/nago/example/cmd/tutorial-119-canvas2-photobook/book/messages"
	"go.wdy.de/nago/pkg/gostrict/lib/asset"
	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
)

// The double page of two A4 pages in millimeters.
const (
	pageW   = 210.0
	pageH   = 297.0
	spreadW = 2 * pageW
	margin  = 24.0
	// snap is the distance in CSS pixels, within which an edge snaps.
	snap = 6.0
	// handle is the radius of a handle in CSS pixels.
	handle = 8.0
)

// Mode is what a drag does.
type Mode int32

const (
	None Mode = iota
	Move
	Scale
	Rotate
)

var ctx = canvas.Get()

// Book is the state of the editor.
type Book struct {
	Photos   list.List[messages.Photo]
	Images   dict.Map[string, canvas.Image]
	Selected int32
	Mode     Mode
	// the pointer and the photo at the start of the drag
	StartX, StartY float64
	Start          messages.Photo
	Changed        bool
	Width, Height  float64
	// Guides are the lines in millimeters, to which the moving photo has snapped: x for vertical lines, y for
	// horizontal ones.
	GuideX list.List[float64]
	GuideY list.List[float64]
}

// Init starts an empty editor and asks the backend for the layout.
func Init() (*Book, msg.Out[messages.Hello]) {
	b := &Book{Photos: list.New[messages.Photo](), Images: dict.New[string, canvas.Image](), Selected: -1, GuideX: list.New[float64](), GuideY: list.New[float64]()}
	out := msg.New[messages.Hello]()
	out.Send(messages.Hello{})
	return b, out
}

// Resize takes the size of the canvas.
func Resize(b *Book, w, h float64) {
	b.Width, b.Height = w, h
}

// view returns the scale from millimeters to CSS pixels and the position of the double page.
func view(b *Book) (float64, float64, float64) {
	s := max(min((b.Width-2*margin)/spreadW, (b.Height-2*margin)/pageH), 0.01)
	return s, (b.Width - spreadW*s) / 2, (b.Height - pageH*s) / 2
}

func image(b *Book, url string) canvas.Image {
	if img, ok := b.Images.Get(url); ok {
		return img
	}
	img := canvas.LoadImage(url)
	b.Images.Set(url, img)
	return img
}

// OnLayout replaces the layout.
func OnLayout(b *Book, l messages.Layout) {
	b.Photos = l.Photos.Clone()
	if b.Selected >= b.Photos.Len() {
		b.Selected = -1
	}
	b.Mode = None
}

// OnAdd puts a new photo into the middle of the left page, 120 mm wide or high, and selects it.
func OnAdd(b *Book, a messages.Add) msg.Out[messages.Save] {
	w, h := 120.0, 120.0
	if a.Width > 0 && a.Height > 0 {
		if a.Width >= a.Height {
			h = w * a.Height / a.Width
		} else {
			w = h * a.Width / a.Height
		}
	}
	b.Photos.Append(messages.Photo{URL: a.URL, X: pageW / 2, Y: pageH / 2, W: w, H: h})
	b.Selected = b.Photos.Len() - 1
	return save(b)
}

// OnLoaded draws a photo, once its image is loaded.
func OnLoaded(b *Book, e asset.Loaded) {
}

func save(b *Book) msg.Out[messages.Save] {
	out := msg.New[messages.Save]()
	out.Send(messages.Save{Photos: b.Photos.Clone()})
	return out
}

// local returns the point (x, y) in millimeters relative to the center of the photo and without its rotation.
func local(p messages.Photo, x, y float64) (float64, float64) {
	dx, dy := x-p.X, y-p.Y
	c, s := math.Cos(-p.Rotation), math.Sin(-p.Rotation)
	return dx*c - dy*s, dx*s + dy*c
}

// world returns the point of the photo relative to its center in millimeters of the double page.
func world(p messages.Photo, lx, ly float64) (float64, float64) {
	c, s := math.Cos(p.Rotation), math.Sin(p.Rotation)
	return p.X + lx*c - ly*s, p.Y + lx*s + ly*c
}

// hit returns the topmost photo at (x, y) in millimeters and what a drag of it does.
func hit(b *Book, x, y, scale float64) (int32, Mode) {
	r := handle / scale
	if p, ok := b.Photos.Get(b.Selected); ok {
		hx, hy := world(p, p.W/2, p.H/2)
		if math.Hypot(x-hx, y-hy) <= r {
			return b.Selected, Scale
		}
		rx, ry := world(p, 0, -p.H/2-4*r)
		if math.Hypot(x-rx, y-ry) <= r {
			return b.Selected, Rotate
		}
	}
	for i := b.Photos.Len() - 1; i >= 0; i-- {
		p, ok := b.Photos.Get(i)
		if !ok {
			continue
		}
		lx, ly := local(p, x, y)
		if math.Abs(lx) <= p.W/2 && math.Abs(ly) <= p.H/2 {
			return i, Move
		}
	}
	return -1, None
}

// OnInput edits the layout with the pointer, the wheel and the keys.
func OnInput(b *Book, e input.Event) msg.Out[messages.Save] {
	scale, ox, oy := view(b)
	x, y := (e.X-ox)/scale, (e.Y-oy)/scale

	switch e.Type {
	case input.PointerDown:
		i, mode := hit(b, x, y, scale)
		b.Selected = i
		b.Mode = mode
		b.StartX, b.StartY = x, y
		if p, ok := b.Photos.Get(i); ok {
			b.Start = p
		}
		b.Changed = false
	case input.PointerMove:
		drag(b, x, y, scale)
	case input.PointerUp, input.PointerCancel:
		b.Mode = None
		b.GuideX = list.New[float64]()
		b.GuideY = list.New[float64]()
		if b.Changed {
			b.Changed = false
			return save(b)
		}
	case input.MouseWheel:
		if p, ok := b.Photos.Get(b.Selected); ok {
			f := math.Exp(-e.DeltaY / 500)
			p.W, p.H = p.W*f, p.H*f
			b.Photos.Set(b.Selected, p)
			return save(b)
		}
	case input.KeyDown:
		return key(b, e)
	default:
	}
	return msg.New[messages.Save]()
}

func drag(b *Book, x, y, scale float64) {
	p := b.Start
	switch b.Mode {
	case Move:
		p.X += x - b.StartX
		p.Y += y - b.StartY
		p = snapTo(b, p, snap/scale)
	case Scale:
		// the corner follows the pointer, the center stays and the ratio of the sides is kept
		lx, ly := local(b.Start, x, y)
		f := max(math.Max(2*lx/b.Start.W, 2*ly/b.Start.H), 0.05)
		p.W, p.H = b.Start.W*f, b.Start.H*f
	case Rotate:
		a := math.Atan2(y-p.Y, x-p.X) + math.Pi/2
		// snap to multiples of 15 degrees
		step := math.Pi / 12
		if r := math.Round(a/step) * step; math.Abs(r-a) < 0.05 {
			a = r
		}
		p.Rotation = a
	case None:
		return
	}
	b.Photos.Set(b.Selected, p)
	b.Changed = true
}

// bounds returns the half width and height of the photo, which the rotation enlarges.
func bounds(p messages.Photo) (float64, float64) {
	c, s := math.Abs(math.Cos(p.Rotation)), math.Abs(math.Sin(p.Rotation))
	return p.W/2*c + p.H/2*s, p.W/2*s + p.H/2*c
}

// snapTo moves the photo, so that its nearest edge or center lies on a line of the pages or of another photo.
func snapTo(b *Book, p messages.Photo, within float64) messages.Photo {
	hw, hh := bounds(p)
	xs := list.Of(0.0, pageW/2, pageW, pageW*1.5, spreadW)
	ys := list.Of(0.0, pageH/2, pageH)
	for i, o := range b.Photos.All() {
		if i == b.Selected {
			continue
		}
		ow, oh := bounds(o)
		xs.Append(o.X-ow, o.X, o.X+ow)
		ys.Append(o.Y-oh, o.Y, o.Y+oh)
	}

	b.GuideX = list.New[float64]()
	b.GuideY = list.New[float64]()
	if d, line, ok := nearest(xs, list.Of(p.X-hw, p.X, p.X+hw), within); ok {
		p.X += d
		b.GuideX.Append(line)
	}
	if d, line, ok := nearest(ys, list.Of(p.Y-hh, p.Y, p.Y+hh), within); ok {
		p.Y += d
		b.GuideY.Append(line)
	}
	return p
}

// nearest returns the shift of the nearest pair of a line and an edge, which are closer than within.
func nearest(lines, edges list.List[float64], within float64) (float64, float64, bool) {
	best, at, found := within, 0.0, false
	shift := 0.0
	for _, l := range lines.All() {
		for _, e := range edges.All() {
			if d := math.Abs(l - e); d < best {
				best, at, found, shift = d, l, true, l-e
			}
		}
	}
	return shift, at, found
}

func key(b *Book, e input.Event) msg.Out[messages.Save] {
	p, ok := b.Photos.Get(b.Selected)
	if !ok {
		return msg.New[messages.Save]()
	}
	step := 1.0
	if e.Shift {
		step = 10
	}
	switch e.Code {
	case "Delete", "Backspace":
		b.Photos.Delete(b.Selected)
		b.Selected = -1
		return save(b)
	case "BracketRight":
		if q, ok := b.Photos.Get(b.Selected + 1); ok {
			b.Photos.Set(b.Selected+1, p)
			b.Photos.Set(b.Selected, q)
			b.Selected++
			return save(b)
		}
	case "BracketLeft":
		if q, ok := b.Photos.Get(b.Selected - 1); ok {
			b.Photos.Set(b.Selected-1, p)
			b.Photos.Set(b.Selected, q)
			b.Selected--
			return save(b)
		}
	case "ArrowLeft":
		p.X -= step
	case "ArrowRight":
		p.X += step
	case "ArrowUp":
		p.Y -= step
	case "ArrowDown":
		p.Y += step
	default:
		return msg.New[messages.Save]()
	}
	b.Photos.Set(b.Selected, p)
	return save(b)
}

// Frame draws the double page.
func Frame(b *Book, t float64) {
	scale, ox, oy := view(b)
	canvas.FillColor(ctx, 0xe9e6e1ff)
	canvas.FillRect(ctx, 0, 0, b.Width, b.Height)

	canvas.Save(ctx)
	canvas.Translate(ctx, ox, oy)
	canvas.Scale(ctx, scale, scale)

	// the pages with a shadow and the fold in the middle
	canvas.Shadow(ctx, 0x00000040, 18, 0, 4)
	canvas.FillColor(ctx, 0xffffffff)
	canvas.FillRect(ctx, 0, 0, spreadW, pageH)
	canvas.Shadow(ctx, 0x00000000, 0, 0, 0)
	fold := canvas.NewLinearGradient(ctx, pageW-12, 0, pageW+12, 0)
	canvas.AddColorStop(fold, 0, 0x00000000)
	canvas.AddColorStop(fold, 0.5, 0x00000022)
	canvas.AddColorStop(fold, 1, 0x00000000)
	canvas.FillPaint(ctx, fold)
	canvas.FillRect(ctx, pageW-12, 0, 24, pageH)

	// the photos do not reach beyond the double page
	canvas.Save(ctx)
	canvas.BeginPath(ctx)
	canvas.Rect(ctx, 0, 0, spreadW, pageH)
	canvas.Clip(ctx)
	for _, p := range b.Photos.All() {
		canvas.Save(ctx)
		canvas.Translate(ctx, p.X, p.Y)
		canvas.Rotate(ctx, p.Rotation)
		img := image(b, p.URL)
		if canvas.ImageWidth(img) > 0 {
			canvas.DrawImageScaled(ctx, img, -p.W/2, -p.H/2, p.W, p.H)
		} else {
			canvas.FillColor(ctx, 0xccd5e0ff)
			canvas.FillRect(ctx, -p.W/2, -p.H/2, p.W, p.H)
		}
		canvas.Restore(ctx)
	}
	canvas.Restore(ctx)

	// the guides of the snapping
	canvas.StrokeColor(ctx, 0xff3b8bff)
	canvas.LineWidth(ctx, 1/scale)
	for _, x := range b.GuideX.All() {
		canvas.BeginPath(ctx)
		canvas.MoveTo(ctx, x, -margin)
		canvas.LineTo(ctx, x, pageH+margin)
		canvas.Stroke(ctx)
	}
	for _, y := range b.GuideY.All() {
		canvas.BeginPath(ctx)
		canvas.MoveTo(ctx, -margin, y)
		canvas.LineTo(ctx, spreadW+margin, y)
		canvas.Stroke(ctx)
	}

	if p, ok := b.Photos.Get(b.Selected); ok {
		selection(p, scale)
	}
	canvas.Restore(ctx)
}

// selection draws the frame and the handles of the selected photo.
func selection(p messages.Photo, scale float64) {
	r := handle / scale
	canvas.Save(ctx)
	canvas.Translate(ctx, p.X, p.Y)
	canvas.Rotate(ctx, p.Rotation)
	canvas.StrokeColor(ctx, 0x1f6fffff)
	canvas.LineWidth(ctx, 2/scale)
	canvas.StrokeRect(ctx, -p.W/2, -p.H/2, p.W, p.H)
	canvas.BeginPath(ctx)
	canvas.MoveTo(ctx, 0, -p.H/2)
	canvas.LineTo(ctx, 0, -p.H/2-3*r)
	canvas.Stroke(ctx)
	canvas.FillColor(ctx, 0xffffffff)
	for _, at := range list.Of(p.W/2, 0.0).All() {
		canvas.BeginPath(ctx)
		if at == 0 {
			canvas.Arc(ctx, 0, -p.H/2-4*r, r, 0, 2*math.Pi, false)
		} else {
			canvas.Arc(ctx, p.W/2, p.H/2, r, 0, 2*math.Pi, false)
		}
		canvas.Fill(ctx)
		canvas.Stroke(ctx)
	}
	canvas.Restore(ctx)
}
