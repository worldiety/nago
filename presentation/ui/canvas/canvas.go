// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package canvas

import (
	"go.wdy.de/nago/application/color"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/proto"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/internal"
)

type TCanvas struct {
	id    string
	frame ui.Frame
}

func Canvas(id string) TCanvas {
	return TCanvas{id: id}
}

func (c TCanvas) Frame(frame ui.Frame) TCanvas {
	c.frame = frame
	return c
}

func (c TCanvas) Render(ctx core.RenderContext) core.RenderNode {
	return &proto.Canvas{Frame: internal.FrameToOra(c.frame), Id: proto.Str(c.id)}
}

type ImgHnd proto.Uint

// TContext2D is a context for drawing on a canvas.
type TContext2D struct {
	wnd core.Window
	id  string
}

func Context2D(wnd core.Window, id string) TContext2D {
	return TContext2D{wnd: wnd, id: id}
}

// FillColor sets the FillStyle as an absolute color value for subsequent drawing operations.
func (c TContext2D) FillColor(color color.Color) TContext2D {
	return c.FillStyle(string(color))
}

func (c TContext2D) LoadImage(hnd ImgHnd, url core.URI) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasLoadImage{Id: proto.Str(c.id), Hnd: proto.Uint(hnd), Url: proto.Str(url)}, nil)
	return c
}

// FillStyle specifies the color, gradient,
// or pattern to use inside shapes. The default style is black.
func (c TContext2D) FillStyle(style string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasFillStyle{Id: proto.Str(c.id), Style: proto.Str(style)}, nil)
	return c
}

// FillRect draws a filled rectangle whose starting point is at (x, y) and whose size is specified by
// width and height. The fill style is determined by the current fillStyle attribute.
func (c TContext2D) FillRect(x, y, width, height float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasFillRect{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
		W:  proto.Float(width),
		H:  proto.Float(height),
	}, nil)
	return c
}

// Arc adds a circular arc to the current sub-path. x, y are the coordinates of the arc's center,
// radius is the arc's radius, startAngle and endAngle are the angles (in radians) at which the arc
// starts and ends. antiClockwise indicates whether the arc is drawn counter-clockwise.
func (c TContext2D) Arc(x, y, radius, startAngle, endAngle float64, antiClockwise bool) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasArc{
		Id:            proto.Str(c.id),
		X:             proto.Float(x),
		Y:             proto.Float(y),
		R:             proto.Float(radius),
		Start:         proto.Float(startAngle),
		End:           proto.Float(endAngle),
		AntiClockwise: proto.Bool(antiClockwise),
	}, nil)
	return c
}

// ArcTo adds a circular arc to the current sub-path using the given control points and radius.
func (c TContext2D) ArcTo(x1, y1, x2, y2, radius float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasArcTo{
		Id:     proto.Str(c.id),
		X1:     proto.Float(x1),
		Y1:     proto.Float(y1),
		X2:     proto.Float(x2),
		Y2:     proto.Float(y2),
		Radius: proto.Float(radius),
	}, nil)
	return c
}

// BeginPath starts a new path by emptying the list of sub-paths.
func (c TContext2D) BeginPath() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasBeginPath{Id: proto.Str(c.id)}, nil)
	return c
}

// BezierCurveTo adds a cubic Bézier curve to the current sub-path.
// cp1x/cp1y is the first control point, cp2x/cp2y the second, x/y the end point.
func (c TContext2D) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasBezierCurveTo{
		Id:   proto.Str(c.id),
		Cp1x: proto.Float(cp1x),
		Cp1y: proto.Float(cp1y),
		Cp2x: proto.Float(cp2x),
		Cp2y: proto.Float(cp2y),
		X:    proto.Float(x),
		Y:    proto.Float(y),
	}, nil)
	return c
}

// CallList replays a previously recorded display list identified by handle.
func (c TContext2D) CallList(handle uint64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasCallList{
		Id:     proto.Str(c.id),
		Handle: proto.Uint(handle),
	}, nil)
	return c
}

// Clear clears the entire canvas.
func (c TContext2D) Clear() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasClear{Id: proto.Str(c.id)}, nil)
	return c
}

// ClearRect erases the pixels in a rectangular area, making it fully transparent.
func (c TContext2D) ClearRect(x, y, width, height float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasClearRect{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
		W:  proto.Float(width),
		H:  proto.Float(height),
	}, nil)
	return c
}

// Clip turns the current path into the current clipping region.
func (c TContext2D) Clip() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasClip{Id: proto.Str(c.id)}, nil)
	return c
}

// ClosePath adds a straight line from the current point to the start of the current sub-path.
func (c TContext2D) ClosePath() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasClosePath{Id: proto.Str(c.id)}, nil)
	return c
}

// DrawImage2 draws an image onto the canvas. hnd is the image handle,
// dx/dy/dw/dh define the destination rectangle, sx/sy/sw/sh the source rectangle.
func (c TContext2D) DrawImage2(hnd ImgHnd, dx, dy, dw, dh, sx, sy, sw, sh float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasDrawImage{
		Id:  proto.Str(c.id),
		Hnd: proto.Uint(hnd),
		Dx:  proto.Float(dx),
		Dy:  proto.Float(dy),
		Dw:  proto.Float(dw),
		Dh:  proto.Float(dh),
		Sx:  proto.Float(sx),
		Sy:  proto.Float(sy),
		Sw:  proto.Float(sw),
		Sh:  proto.Float(sh),
	}, nil)
	return c
}

// DrawImage is a simplified version of DrawImage2 that only takes destination coordinates.
func (c TContext2D) DrawImage(hnd ImgHnd, dx, dy float64) TContext2D {
	c.DrawImage2(hnd, dx, dy, 0, 0, 0, 0, 0, 0)
	return c
}

// EndList ends a display list recording previously started with NewList.
func (c TContext2D) EndList() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasEndList{Id: proto.Str(c.id)}, nil)
	return c
}

// Fill fills the current path with the current fill style.
func (c TContext2D) Fill() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasFill{Id: proto.Str(c.id)}, nil)
	return c
}

// FillText draws a text string at the specified coordinates using the current fillStyle.
// maxWidth limits the rendered width; the text is scaled down to fit if needed.
func (c TContext2D) FillText(text string, x, y, maxWidth float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasFillText{
		Id:       proto.Str(c.id),
		Text:     proto.Str(text),
		X:        proto.Float(x),
		Y:        proto.Float(y),
		MaxWidth: proto.Float(maxWidth),
	}, nil)
	return c
}

// Font sets the font used for text operations, e.g. "16px sans-serif".
func (c TContext2D) Font(font string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasFont{Id: proto.Str(c.id), Font: proto.Str(font)}, nil)
	return c
}

// LineTo adds a straight line to the current sub-path from the last point to (x, y).
func (c TContext2D) LineTo(x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasLineTo{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
	}, nil)
	return c
}

// MoveTo begins a new sub-path at the point (x, y).
func (c TContext2D) MoveTo(x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasMoveTo{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
	}, nil)
	return c
}

// NewList begins recording a new display list identified by handle.
// Subsequent drawing commands are recorded until EndList is called.
// The list can later be replayed with CallList.
func (c TContext2D) NewList(handle uint64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasNewList{
		Id:     proto.Str(c.id),
		Handle: proto.Uint(handle),
	}, nil)
	return c
}

// QuadraticCurveTo adds a quadratic Bézier curve; cpx/cpy is the control point, x/y the end point.
func (c TContext2D) QuadraticCurveTo(cpx, cpy, x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasQuadraticCurveTo{
		Id:  proto.Str(c.id),
		Cpx: proto.Float(cpx),
		Cpy: proto.Float(cpy),
		X:   proto.Float(x),
		Y:   proto.Float(y),
	}, nil)
	return c
}

// Rect adds a rectangle to the current path.
func (c TContext2D) Rect(x, y, width, height float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasRect{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
		W:  proto.Float(width),
		H:  proto.Float(height),
	}, nil)
	return c
}

// Restore restores the most recently saved canvas state.
func (c TContext2D) Restore() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasRestore{Id: proto.Str(c.id)}, nil)
	return c
}

// Rotate adds a rotation to the transformation matrix.
// angle is in radians; to convert degrees: deg * math.Pi / 180.
func (c TContext2D) Rotate(angle float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasRotate{
		Id:    proto.Str(c.id),
		Angle: proto.Float(angle),
	}, nil)
	return c
}

// Save saves the entire state of the canvas by pushing it onto a stack.
func (c TContext2D) Save() TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasSave{Id: proto.Str(c.id)}, nil)
	return c
}

// Scale adds a scaling transformation to the canvas units.
func (c TContext2D) Scale(x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasScale{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
	}, nil)
	return c
}

// SetTransform resets the current transformation to the identity matrix and applies the given matrix.
// The matrix has the form: [a c e / b d f / 0 0 1].
func (c TContext2D) SetTransform(a, b, cc, d, e, f float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasSetTransform{
		Id: proto.Str(c.id),
		A:  proto.Float(a),
		B:  proto.Float(b),
		C:  proto.Float(cc),
		D:  proto.Float(d),
		E:  proto.Float(e),
		F:  proto.Float(f),
	}, nil)
	return c
}

// StrokeColor sets the stroke style as an absolute color value.
func (c TContext2D) StrokeColor(clr color.Color) TContext2D {
	return c.StrokeStyle(string(clr))
}

// StrokeStyle sets the color, gradient, or pattern used for strokes around shapes.
func (c TContext2D) StrokeStyle(style string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasStrokeStyle{Id: proto.Str(c.id), Style: proto.Str(style)}, nil)
	return c
}

// StrokeRect draws a stroked rectangle according to the current strokeStyle.
func (c TContext2D) StrokeRect(x, y, width, height float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasStrokeRect{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
		W:  proto.Float(width),
		H:  proto.Float(height),
	}, nil)
	return c
}

// StrokeText draws the outlines of a text string at the specified coordinates.
// maxWidth limits the rendered width.
func (c TContext2D) StrokeText(text string, x, y, maxWidth float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasStrokeText{
		Id:       proto.Str(c.id),
		Text:     proto.Str(text),
		X:        proto.Float(x),
		Y:        proto.Float(y),
		MaxWidth: proto.Float(maxWidth),
	}, nil)
	return c
}

// TextAlign sets the text alignment. Valid values: "left", "right", "center", "start", "end".
func (c TContext2D) TextAlign(align string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasTextAlign{Id: proto.Str(c.id), TextAlign: proto.Str(align)}, nil)
	return c
}

// TextBaseline sets the text baseline. Valid values: "top", "hanging", "middle", "alphabetic", "ideographic", "bottom".
func (c TContext2D) TextBaseline(baseline string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasTextBaseline{Id: proto.Str(c.id), Baseline: proto.Str(baseline)}, nil)
	return c
}

// Translate adds a translation transformation, moving the canvas origin to (x, y).
func (c TContext2D) Translate(x, y float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasTranslate{
		Id: proto.Str(c.id),
		X:  proto.Float(x),
		Y:  proto.Float(y),
	}, nil)
	return c
}

func (c TContext2D) LineWidth(width float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasLineWidth{Id: proto.Str(c.id), Width: proto.Float(width)}, nil)
	return c
}

func (c TContext2D) LineCap(cap string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasLineCap{Id: proto.Str(c.id), Cap: proto.Str(cap)}, nil)
	return c
}

func (c TContext2D) LineJoin(join string) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasLineJoin{Id: proto.Str(c.id), Join: proto.Str(join)}, nil)
	return c
}

func (c TContext2D) MiterLimit(limit float64) TContext2D {
	core.AsyncCall(c.wnd, &proto.CanvasMiterLimit{Id: proto.Str(c.id), Limit: proto.Float(limit)}, nil)
	return c
}
