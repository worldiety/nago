// Package paint is a small paint app for the canvas2 component of nago: freehand strokes with several pointers,
// a palette in a toolbar, a brush width, which the wheel changes, undo and a paper texture as background.
package paint

import (
	"go.wdy.de/nago/pkg/gostrict/lib/asset"
	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/pkg/gostrict/std/dict"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
	"go.wdy.de/nago/pkg/gostrict/std/text"

	"go.wdy.de/nago/pkg/gostrict/testdata/llm2/paint/proto"
)

const (
	paperURL      = "/paper.png"
	toolbarHeight = 48.0
	swatchSize    = 32.0
	swatchGap     = 8.0
	minWidth      = 1.0
	maxWidth      = 64.0
	wheelFactor   = 0.02
	paperColor    = uint32(0xf4efe4ff)
	toolbarColor  = uint32(0x2b2b2bff)
	selectColor   = uint32(0xffffffff)
)

// palette holds the colors, which the digits 1 to 5 and the swatches select.
var palette = list.Freeze(list.Of[uint32](
	0x1a1a1aff, // black
	0xd62828ff, // red
	0x2a9d8fff, // green
	0x1d4ed8ff, // blue
	0xf4a261ff, // orange
))

// Stroke is a stroke on the canvas, finished or still drawn.
type Stroke struct {
	Points list.List[proto.Point]
	Color  uint32
	Width  float64
}

// App is the state of the paint app.
type App struct {
	W, H        float64
	Strokes     list.List[Stroke]
	Active      dict.Map[int32, Stroke]
	ColorIndex  int32
	BrushWidth  float64
	Paper       canvas.Image
	PaperLoaded bool
	Time        float64
}

// Init returns the state and starts loading the paper texture.
func Init() *App {
	return &App{
		W:          0,
		H:          0,
		Strokes:    list.New[Stroke](),
		Active:     dict.New[int32, Stroke](),
		ColorIndex: 0,
		BrushWidth: 4,
		Paper:      canvas.LoadImage(paperURL),
	}
}

// Resize stores the size of the canvas in CSS pixels.
func Resize(a *App, w, h float64) {
	a.W = w
	a.H = h
}

// Frame redraws the background, all strokes and the toolbar.
func Frame(a *App, t float64) {
	a.Time = t
	c := canvas.Get()
	drawBackground(a, c)
	for s := range a.Strokes.Values() {
		drawStroke(c, s)
	}
	for _, s := range a.Active.All() {
		drawStroke(c, s)
	}
	drawToolbar(a, c)
}

// OnInput draws strokes with the pointers, changes the brush with the wheel and handles the keys. It sends each
// finished stroke to the backend.
func OnInput(a *App, e input.Event) msg.Out[proto.Stroke] {
	out := msg.New[proto.Stroke]()
	switch e.Type {
	case input.PointerDown:
		if e.Y < toolbarHeight {
			if i, ok := swatchAt(e.X, e.Y); ok {
				a.ColorIndex = i
			}
			return out
		}
		pts := list.New[proto.Point]()
		pts.Append(proto.Point{X: e.X, Y: e.Y})
		a.Active.Set(e.PointerID, Stroke{Points: pts, Color: currentColor(a), Width: a.BrushWidth})
	case input.PointerMove:
		if s, ok := a.Active.Get(e.PointerID); ok {
			s.Points.Append(proto.Point{X: e.X, Y: e.Y})
		}
	case input.PointerUp, input.PointerCancel:
		if s, ok := a.Active.Get(e.PointerID); ok {
			if e.Type == input.PointerUp {
				s.Points.Append(proto.Point{X: e.X, Y: e.Y})
			}
			a.Active.Delete(e.PointerID)
			a.Strokes.Append(s)
			out.Send(proto.Stroke{Points: s.Points.Clone(), Color: s.Color, Width: s.Width})
		}
	case input.MouseWheel:
		a.BrushWidth = clamp(a.BrushWidth-e.DeltaY*wheelFactor, minWidth, maxWidth)
	case input.KeyDown:
		handleKey(a, e)
	default:
	}
	return out
}

// OnLoaded marks the paper texture as loaded.
func OnLoaded(a *App, e asset.Loaded) {
	if e.URL == paperURL {
		a.PaperLoaded = true
	}
}

// OnClear removes all strokes, when the backend asks for it.
func OnClear(a *App, m proto.Clear) {
	clearAll(a)
}

func handleKey(a *App, e input.Event) {
	if e.Code == "KeyZ" && (e.Ctrl || e.Meta) {
		a.Strokes.Pop()
		return
	}
	switch e.Code {
	case "Digit1":
		a.ColorIndex = 0
	case "Digit2":
		a.ColorIndex = 1
	case "Digit3":
		a.ColorIndex = 2
	case "Digit4":
		a.ColorIndex = 3
	case "Digit5":
		a.ColorIndex = 4
	case "Delete":
		clearAll(a)
	}
}

func clearAll(a *App) {
	a.Strokes = list.New[Stroke]()
	a.Active = dict.New[int32, Stroke]()
}

func currentColor(a *App) uint32 {
	c, ok := palette.Get(a.ColorIndex)
	if !ok {
		return 0x000000ff
	}
	return c
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func swatchX(i int32) float64 {
	return swatchGap + float64(i)*(swatchSize+swatchGap)
}

func swatchY() float64 {
	return (toolbarHeight - swatchSize) / 2
}

// swatchAt returns the index of the swatch at (x, y).
func swatchAt(x, y float64) (int32, bool) {
	for i := range palette.Len() {
		sx := swatchX(i)
		sy := swatchY()
		if x >= sx && x <= sx+swatchSize && y >= sy && y <= sy+swatchSize {
			return i, true
		}
	}
	return 0, false
}

func drawBackground(a *App, c canvas.Context) {
	if a.PaperLoaded {
		p := canvas.NewPattern(c, a.Paper, "repeat")
		canvas.FillPaint(c, p)
	} else {
		canvas.FillColor(c, paperColor)
	}
	canvas.FillRect(c, 0, 0, a.W, a.H)
}

func drawStroke(c canvas.Context, s Stroke) {
	first, ok := s.Points.Get(0)
	if !ok {
		return
	}
	if s.Points.Len() == 1 {
		canvas.BeginPath(c)
		canvas.Arc(c, first.X, first.Y, s.Width/2, 0, 6.283185307179586, false)
		canvas.FillColor(c, s.Color)
		canvas.Fill(c)
		return
	}
	canvas.BeginPath(c)
	canvas.LineCap(c, "round")
	canvas.LineJoin(c, "round")
	canvas.LineWidth(c, s.Width)
	canvas.StrokeColor(c, s.Color)
	for i, p := range s.Points.All() {
		if i == 0 {
			canvas.MoveTo(c, p.X, p.Y)
		} else {
			canvas.LineTo(c, p.X, p.Y)
		}
	}
	canvas.Stroke(c)
}

func drawToolbar(a *App, c canvas.Context) {
	canvas.FillColor(c, toolbarColor)
	canvas.FillRect(c, 0, 0, a.W, toolbarHeight)
	for i, col := range palette.All() {
		x := swatchX(i)
		y := swatchY()
		canvas.FillColor(c, col)
		canvas.FillRect(c, x, y, swatchSize, swatchSize)
		if i == a.ColorIndex {
			canvas.LineWidth(c, 3)
			canvas.StrokeColor(c, selectColor)
			canvas.StrokeRect(c, x-2, y-2, swatchSize+4, swatchSize+4)
		}
	}
	// the brush width as text and a preview dot in the current color
	tx := swatchX(palette.Len()) + swatchGap
	canvas.Font(c, "14px sans-serif")
	canvas.TextBaseline(c, "middle")
	canvas.TextAlign(c, "left")
	canvas.FillColor(c, selectColor)
	canvas.FillText(c, "Width: "+text.FromFloat(a.BrushWidth, 1)+" px", tx, toolbarHeight/2)
	w, _, _ := canvas.MeasureText(c, "Width: 00.0 px")
	r := clamp(a.BrushWidth/2, 1, toolbarHeight/2-4)
	canvas.BeginPath(c)
	canvas.Arc(c, tx+w+swatchGap+toolbarHeight/2, toolbarHeight/2, r, 0, 6.283185307179586, false)
	canvas.FillColor(c, currentColor(a))
	canvas.Fill(c)
}
