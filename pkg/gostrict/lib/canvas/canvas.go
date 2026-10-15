// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package canvas binds the 2D context of a canvas to the gostrict subset, see the component canvas2 of nago. Its
// handles are the objects of the browser itself, so a call costs a JavaScript call and nothing else, e.g.
// (c, x, y, w, h) => c.fillRect(x, y, w, h), which the engine inlines.
//
// The program gets the context of its component with [Get], e.g. once in a package variable, and draws in CSS
// pixels: the host scales the canvas to the pixels of the device and starts each frame with that scale and no
// other transformation, so [SetTransform] and [ResetTransform] are relative to it. The host neither clears the
// canvas nor resets the other properties, e.g. the fill style or the font, so a frame usually starts with [Clear]
// or draws its whole area. Gradients, patterns and paths are cheap, but a program may keep them in its state. Colors are uint32 values 0xRRGGBBAA, e.g. 0xff0000ff
// for an opaque red, or CSS colors as strings with [FillStyle] and [StrokeStyle].
//
// Images are loaded with [LoadImage], which returns a handle right away. The host reports the result with the
// events of package lib/asset, whose URL is the string, which the program passed, and drawing an image, which is
// not loaded yet, does nothing.
//
// The native implementation records the commands, so that a test compares them, see [Commands].
//
//gostrict:header
package canvas

import "go.wdy.de/nago/pkg/gostrict/std/js"

// Context is the 2D context of the canvas of the component.
type Context struct{ ref js.Ref }

// Path is a Path2D, which may be drawn many times and tested for hits.
type Path struct{ ref js.Ref }

// Gradient is a linear, radial or conic gradient.
type Gradient struct{ ref js.Ref }

// Pattern repeats an image.
type Pattern struct{ ref js.Ref }

// Image is an image, which the host loads, see [LoadImage].
type Image struct{ ref js.Ref }

// Pather is anything, which builds a path: the current path of the context or a [Path].
type Pather interface {
	Context | Path
}

// Paint is a style other than a color.
type Paint interface {
	Gradient | Pattern
}

// Get returns the context of the canvas of the component.
//
//gostrict:js () => $host.context
func Get() Context {
	return Context{newContext()}
}

// Width returns the width of the canvas in CSS pixels.
//
//gostrict:js (c) => $host.width
func Width(c Context) float64 {
	return ctxOf(c.ref).width
}

// Height returns the height of the canvas in CSS pixels.
//
//gostrict:js (c) => $host.height
func Height(c Context) float64 {
	return ctxOf(c.ref).height
}

// Save pushes the state of the context, e.g. its transformation, styles and clip.
//
//gostrict:js (c) => { c.save(); }
func Save(c Context) {
	rec(c.ref, "save")
}

// Restore pops the state of the context, which Save has pushed.
//
//gostrict:js (c) => { c.restore(); }
func Restore(c Context) {
	rec(c.ref, "restore")
}

// Translate moves the origin.
//
//gostrict:js (c, x, y) => { c.translate(x, y); }
func Translate(c Context, x, y float64) {
	rec(c.ref, "translate", x, y)
}

// Rotate rotates by the angle in radians clockwise.
//
//gostrict:js (c, a) => { c.rotate(a); }
func Rotate(c Context, angle float64) {
	rec(c.ref, "rotate", angle)
}

// Scale scales the axes.
//
//gostrict:js (c, x, y) => { c.scale(x, y); }
func Scale(c Context, x, y float64) {
	rec(c.ref, "scale", x, y)
}

// Transform multiplies the transformation with the matrix [a c e; b d f; 0 0 1].
//
//gostrict:js (c, a, b, cc, d, e, f) => { c.transform(a, b, cc, d, e, f); }
func Transform(c Context, a, b, cc, d, e, f float64) {
	rec(c.ref, "transform", a, b, cc, d, e, f)
}

// SetTransform replaces the transformation with the matrix [a c e; b d f; 0 0 1] in CSS pixels.
//
//gostrict:js (c, a, b, cc, d, e, f) => { const r = $host.dpr; c.setTransform(a * r, b * r, cc * r, d * r, e * r, f * r); }
func SetTransform(c Context, a, b, cc, d, e, f float64) {
	rec(c.ref, "setTransform", a, b, cc, d, e, f)
}

// ResetTransform removes all transformations.
//
//gostrict:js (c) => { const r = $host.dpr; c.setTransform(r, 0, 0, r, 0, 0); }
func ResetTransform(c Context) {
	rec(c.ref, "setTransform", 1.0, 0.0, 0.0, 1.0, 0.0, 0.0)
}

// GlobalAlpha sets the opacity of everything, which is drawn, from 0 to 1.
//
//gostrict:js (c, a) => { c.globalAlpha = a; }
func GlobalAlpha(c Context, alpha float64) {
	rec(c.ref, "globalAlpha", alpha)
}

// CompositeOperation sets how drawing combines with the canvas, e.g. source-over, multiply or destination-out.
//
//gostrict:js (c, op) => { c.globalCompositeOperation = op; }
func CompositeOperation(c Context, op string) {
	rec(c.ref, "globalCompositeOperation", op)
}

// ImageSmoothing turns the smoothing of scaled images on or off, e.g. off for pixel art.
//
//gostrict:js (c, on) => { c.imageSmoothingEnabled = on; }
func ImageSmoothing(c Context, on bool) {
	rec(c.ref, "imageSmoothingEnabled", on)
}

// FillColor sets the color 0xRRGGBBAA for filling.
//
//gostrict:js (c, v) => { c.fillStyle = '#' + (v >>> 0).toString(16).padStart(8, '0'); }
func FillColor(c Context, rgba uint32) {
	rec(c.ref, "fillStyle", hex(rgba))
}

// StrokeColor sets the color 0xRRGGBBAA for lines.
//
//gostrict:js (c, v) => { c.strokeStyle = '#' + (v >>> 0).toString(16).padStart(8, '0'); }
func StrokeColor(c Context, rgba uint32) {
	rec(c.ref, "strokeStyle", hex(rgba))
}

// FillStyle sets a CSS color for filling, e.g. rebeccapurple.
//
//gostrict:js (c, s) => { c.fillStyle = s; }
func FillStyle(c Context, css string) {
	rec(c.ref, "fillStyle", css)
}

// StrokeStyle sets a CSS color for lines.
//
//gostrict:js (c, s) => { c.strokeStyle = s; }
func StrokeStyle(c Context, css string) {
	rec(c.ref, "strokeStyle", css)
}

// FillPaint fills with a gradient or a pattern.
//
//gostrict:js (c, p) => { c.fillStyle = p; }
func FillPaint[T Paint](c Context, p T) {
	rec(c.ref, "fillStyle", ctxName(refOf(p)))
}

// StrokePaint draws lines with a gradient or a pattern.
//
//gostrict:js (c, p) => { c.strokeStyle = p; }
func StrokePaint[T Paint](c Context, p T) {
	rec(c.ref, "strokeStyle", ctxName(refOf(p)))
}

// NewLinearGradient returns a gradient along the line from (x0, y0) to (x1, y1).
//
//gostrict:js (c, x0, y0, x1, y1) => c.createLinearGradient(x0, y0, x1, y1)
func NewLinearGradient(c Context, x0, y0, x1, y1 float64) Gradient {
	return Gradient{newObject("linear")}
}

// NewRadialGradient returns a gradient between two circles.
//
//gostrict:js (c, x0, y0, r0, x1, y1, r1) => c.createRadialGradient(x0, y0, r0, x1, y1, r1)
func NewRadialGradient(c Context, x0, y0, r0, x1, y1, r1 float64) Gradient {
	return Gradient{newObject("radial")}
}

// NewConicGradient returns a gradient around the point (x, y), which starts at the angle in radians.
//
//gostrict:js (c, a, x, y) => c.createConicGradient(a, x, y)
func NewConicGradient(c Context, angle, x, y float64) Gradient {
	return Gradient{newObject("conic")}
}

// AddColorStop adds the color 0xRRGGBBAA at the offset from 0 to 1.
//
//gostrict:js (g, o, v) => { g.addColorStop(Math.min(Math.max(o, 0), 1), '#' + (v >>> 0).toString(16).padStart(8, '0')); }
func AddColorStop(g Gradient, offset float64, rgba uint32) {
	rec(g.ref, "addColorStop", offset, hex(rgba))
}

// NewPattern repeats the image: repeat, repeat-x, repeat-y or no-repeat. An image, which is not loaded yet, is transparent.
//
//gostrict:js (c, i, r) => i.bitmap ? c.createPattern(i.bitmap, r) : c.createPattern(new OffscreenCanvas(1, 1), r)
func NewPattern(c Context, img Image, repetition string) Pattern {
	return Pattern{newObject("pattern")}
}

// LineWidth sets the width of lines.
//
//gostrict:js (c, w) => { c.lineWidth = w; }
func LineWidth(c Context, w float64) {
	rec(c.ref, "lineWidth", w)
}

// LineCap sets the end of lines: butt, round or square.
//
//gostrict:js (c, s) => { c.lineCap = s; }
func LineCap(c Context, cap string) {
	rec(c.ref, "lineCap", cap)
}

// LineJoin sets the corners of lines: miter, round or bevel.
//
//gostrict:js (c, s) => { c.lineJoin = s; }
func LineJoin(c Context, join string) {
	rec(c.ref, "lineJoin", join)
}

// MiterLimit limits the length of mitered corners.
//
//gostrict:js (c, l) => { c.miterLimit = l; }
func MiterLimit(c Context, limit float64) {
	rec(c.ref, "miterLimit", limit)
}

// SetLineDash sets the lengths of dashes and gaps separated by spaces, e.g. "4 2", or a solid line for "".
//
//gostrict:js (c, p) => { c.setLineDash(p.split(' ').filter((s) => s !== '').map(Number).filter((n) => n >= 0 && isFinite(n))); }
func SetLineDash(c Context, pattern string) {
	rec(c.ref, "setLineDash", dashes(pattern))
}

// LineDashOffset shifts the dashes.
//
//gostrict:js (c, o) => { c.lineDashOffset = o; }
func LineDashOffset(c Context, offset float64) {
	rec(c.ref, "lineDashOffset", offset)
}

// Shadow sets the shadow of everything, which is drawn: its color 0xRRGGBBAA, its blur and its offset. A transparent color turns it off.
//
//gostrict:js (c, v, b, x, y) => { c.shadowColor = '#' + (v >>> 0).toString(16).padStart(8, '0'); c.shadowBlur = b; c.shadowOffsetX = x; c.shadowOffsetY = y; }
func Shadow(c Context, rgba uint32, blur, dx, dy float64) {
	rec(c.ref, "shadowColor", hex(rgba))
	rec(c.ref, "shadowBlur", blur)
	rec(c.ref, "shadowOffsetX", dx)
	rec(c.ref, "shadowOffsetY", dy)
}

// NewPath returns an empty path.
//
//gostrict:js () => new Path2D()
func NewPath() Path {
	return Path{newObject("path")}
}

// NewPathSVG returns the path of SVG path data, e.g. "M 0 0 L 10 10".
//
//gostrict:js (d) => new Path2D(d)
func NewPathSVG(d string) Path {
	return Path{newObject("path " + d)}
}

// BeginPath starts a new current path.
//
//gostrict:js (c) => { c.beginPath(); }
func BeginPath(c Context) {
	rec(c.ref, "beginPath")
}

// ClosePath connects the end of the path with its start.
//
//gostrict:js (p) => { p.closePath(); }
func ClosePath[T Pather](p T) {
	rec(refOf(p), "closePath")
}

// MoveTo starts a sub path at (x, y).
//
//gostrict:js (p, x, y) => { p.moveTo(x, y); }
func MoveTo[T Pather](p T, x, y float64) {
	rec(refOf(p), "moveTo", x, y)
}

// LineTo adds a line to (x, y).
//
//gostrict:js (p, x, y) => { p.lineTo(x, y); }
func LineTo[T Pather](p T, x, y float64) {
	rec(refOf(p), "lineTo", x, y)
}

// BezierCurveTo adds a cubic Bézier curve with the control points (c1x, c1y) and (c2x, c2y) to (x, y).
//
//gostrict:js (p, a, b, c, d, x, y) => { p.bezierCurveTo(a, b, c, d, x, y); }
func BezierCurveTo[T Pather](p T, c1x, c1y, c2x, c2y, x, y float64) {
	rec(refOf(p), "bezierCurveTo", c1x, c1y, c2x, c2y, x, y)
}

// QuadraticCurveTo adds a quadratic Bézier curve with the control point (cx, cy) to (x, y).
//
//gostrict:js (p, a, b, x, y) => { p.quadraticCurveTo(a, b, x, y); }
func QuadraticCurveTo[T Pather](p T, cx, cy, x, y float64) {
	rec(refOf(p), "quadraticCurveTo", cx, cy, x, y)
}

// Arc adds an arc of the circle around (x, y) from the angle start to end in radians.
//
//gostrict:js (p, x, y, r, s, e, ccw) => { p.arc(x, y, Math.max(r, 0), s, e, ccw); }
func Arc[T Pather](p T, x, y, r, start, end float64, counterclockwise bool) {
	rec(refOf(p), "arc", x, y, r, start, end, counterclockwise)
}

// ArcTo adds an arc with the radius, which touches the lines to (x1, y1) and from there to (x2, y2).
//
//gostrict:js (p, a, b, c, d, r) => { p.arcTo(a, b, c, d, Math.max(r, 0)); }
func ArcTo[T Pather](p T, x1, y1, x2, y2, r float64) {
	rec(refOf(p), "arcTo", x1, y1, x2, y2, r)
}

// Ellipse adds an arc of the ellipse around (x, y) with the radii, which is rotated by rotation.
//
//gostrict:js (p, x, y, rx, ry, rot, s, e, ccw) => { p.ellipse(x, y, Math.max(rx, 0), Math.max(ry, 0), rot, s, e, ccw); }
func Ellipse[T Pather](p T, x, y, rx, ry, rotation, start, end float64, counterclockwise bool) {
	rec(refOf(p), "ellipse", x, y, rx, ry, rotation, start, end, counterclockwise)
}

// Rect adds a rectangle.
//
//gostrict:js (p, x, y, w, h) => { p.rect(x, y, w, h); }
func Rect[T Pather](p T, x, y, w, h float64) {
	rec(refOf(p), "rect", x, y, w, h)
}

// RoundRect adds a rectangle with rounded corners of the radius.
//
//gostrict:js (p, x, y, w, h, r) => { p.roundRect(x, y, w, h, Math.max(r, 0)); }
func RoundRect[T Pather](p T, x, y, w, h, r float64) {
	rec(refOf(p), "roundRect", x, y, w, h, r)
}

// AddPath adds the path q to p.
//
//gostrict:js (p, q) => { p.addPath(q); }
func AddPath(p Path, q Path) {
	rec(p.ref, "addPath", ctxName(q.ref))
}

// Fill fills the current path with the nonzero rule.
//
//gostrict:js (c) => { c.fill(); }
func Fill(c Context) {
	rec(c.ref, "fill")
}

// FillEvenOdd fills the current path with the evenodd rule.
//
//gostrict:js (c) => { c.fill('evenodd'); }
func FillEvenOdd(c Context) {
	rec(c.ref, "fill", "evenodd")
}

// Stroke draws the lines of the current path.
//
//gostrict:js (c) => { c.stroke(); }
func Stroke(c Context) {
	rec(c.ref, "stroke")
}

// Clip restricts the drawing to the current path.
//
//gostrict:js (c) => { c.clip(); }
func Clip(c Context) {
	rec(c.ref, "clip")
}

// FillPath fills the path with the nonzero rule.
//
//gostrict:js (c, p) => { c.fill(p); }
func FillPath(c Context, p Path) {
	rec(c.ref, "fill", ctxName(p.ref))
}

// StrokePath draws the lines of the path.
//
//gostrict:js (c, p) => { c.stroke(p); }
func StrokePath(c Context, p Path) {
	rec(c.ref, "stroke", ctxName(p.ref))
}

// ClipPath restricts the drawing to the path.
//
//gostrict:js (c, p) => { c.clip(p); }
func ClipPath(c Context, p Path) {
	rec(c.ref, "clip", ctxName(p.ref))
}

// IsPointInPath tells, whether the point (x, y) in CSS pixels is inside the path, which is drawn with the current transformation.
//
//gostrict:js (c, p, x, y) => c.isPointInPath(p, x * $host.dpr, y * $host.dpr)
func IsPointInPath(c Context, p Path, x, y float64) bool {
	return false
}

// IsPointInStroke tells, whether the point (x, y) in CSS pixels is on the lines of the path with the current line width.
//
//gostrict:js (c, p, x, y) => c.isPointInStroke(p, x * $host.dpr, y * $host.dpr)
func IsPointInStroke(c Context, p Path, x, y float64) bool {
	return false
}

// FillRect fills a rectangle.
//
//gostrict:js (c, x, y, w, h) => { c.fillRect(x, y, w, h); }
func FillRect(c Context, x, y, w, h float64) {
	rec(c.ref, "fillRect", x, y, w, h)
}

// StrokeRect draws the lines of a rectangle.
//
//gostrict:js (c, x, y, w, h) => { c.strokeRect(x, y, w, h); }
func StrokeRect(c Context, x, y, w, h float64) {
	rec(c.ref, "strokeRect", x, y, w, h)
}

// ClearRect makes a rectangle transparent.
//
//gostrict:js (c, x, y, w, h) => { c.clearRect(x, y, w, h); }
func ClearRect(c Context, x, y, w, h float64) {
	rec(c.ref, "clearRect", x, y, w, h)
}

// Clear makes the whole canvas transparent, whatever the transformation is.
//
//gostrict:js (c) => { c.save(); c.setTransform(1, 0, 0, 1, 0, 0); c.clearRect(0, 0, c.canvas.width, c.canvas.height); c.restore(); }
func Clear(c Context) {
	clearAll(c.ref)
}

// Font sets the font as in CSS, e.g. "bold 16px sans-serif".
//
//gostrict:js (c, f) => { c.font = f; }
func Font(c Context, font string) {
	rec(c.ref, "font", font)
}

// TextAlign sets the horizontal alignment of text: start, end, left, right or center.
//
//gostrict:js (c, a) => { c.textAlign = a; }
func TextAlign(c Context, align string) {
	rec(c.ref, "textAlign", align)
}

// TextBaseline sets the vertical alignment of text: alphabetic, top, hanging, middle, ideographic or bottom.
//
//gostrict:js (c, b) => { c.textBaseline = b; }
func TextBaseline(c Context, baseline string) {
	rec(c.ref, "textBaseline", baseline)
}

// Direction sets the direction of text: ltr, rtl or inherit.
//
//gostrict:js (c, d) => { c.direction = d; }
func Direction(c Context, dir string) {
	rec(c.ref, "direction", dir)
}

// FillText fills the text at (x, y).
//
//gostrict:js (c, t, x, y) => { c.fillText(t, x, y); }
func FillText(c Context, text string, x, y float64) {
	rec(c.ref, "fillText", text, x, y)
}

// FillTextMax fills the text at (x, y) and condenses it to the width.
//
//gostrict:js (c, t, x, y, w) => { c.fillText(t, x, y, Math.max(w, 0)); }
func FillTextMax(c Context, text string, x, y, maxWidth float64) {
	rec(c.ref, "fillText", text, x, y, maxWidth)
}

// StrokeText draws the outline of the text at (x, y).
//
//gostrict:js (c, t, x, y) => { c.strokeText(t, x, y); }
func StrokeText(c Context, text string, x, y float64) {
	rec(c.ref, "strokeText", text, x, y)
}

// StrokeTextMax draws the outline of the text at (x, y) and condenses it to the width.
//
//gostrict:js (c, t, x, y, w) => { c.strokeText(t, x, y, Math.max(w, 0)); }
func StrokeTextMax(c Context, text string, x, y, maxWidth float64) {
	rec(c.ref, "strokeText", text, x, y, maxWidth)
}

// MeasureText returns the width of the text with the current font and how far it extends above and below the baseline.
//
//gostrict:js (c, t) => { const m = c.measureText(t); return [m.width, m.actualBoundingBoxAscent, m.actualBoundingBoxDescent]; }
func MeasureText(c Context, text string) (width, ascent, descent float64) {
	return measure(ctxOf(c.ref), text)
}

// LoadImage loads the image at the URL of the application, e.g. a resource or an upload, and returns its handle right away. The host reports the result with asset.Loaded or asset.Failed.
//
//gostrict:js (url) => $host.loadImage(url)
func LoadImage(url string) Image {
	return Image{newObject("image " + url)}
}

// ImageWidth returns the width of the image in pixels or 0, if it is not loaded.
//
//gostrict:js (i) => i.width
func ImageWidth(img Image) float64 {
	return 0
}

// ImageHeight returns the height of the image in pixels or 0, if it is not loaded.
//
//gostrict:js (i) => i.height
func ImageHeight(img Image) float64 {
	return 0
}

// DrawImage draws the image at (x, y) in its size. An image, which is not loaded yet, is not drawn.
//
//gostrict:js (c, i, x, y) => { if (i.bitmap) c.drawImage(i.bitmap, x, y); }
func DrawImage(c Context, img Image, x, y float64) {
	rec(c.ref, "drawImage", ctxName(img.ref), x, y)
}

// DrawImageScaled draws the image into the rectangle.
//
//gostrict:js (c, i, x, y, w, h) => { if (i.bitmap) c.drawImage(i.bitmap, x, y, w, h); }
func DrawImageScaled(c Context, img Image, x, y, w, h float64) {
	rec(c.ref, "drawImage", ctxName(img.ref), x, y, w, h)
}

// DrawImageRegion draws the region (sx, sy, sw, sh) of the image into the rectangle (dx, dy, dw, dh), e.g. a frame of a sprite sheet.
//
//gostrict:js (c, i, sx, sy, sw, sh, dx, dy, dw, dh) => { if (i.bitmap) c.drawImage(i.bitmap, sx, sy, sw, sh, dx, dy, dw, dh); }
func DrawImageRegion(c Context, img Image, sx, sy, sw, sh, dx, dy, dw, dh float64) {
	rec(c.ref, "drawImage", ctxName(img.ref), sx, sy, sw, sh, dx, dy, dw, dh)
}
