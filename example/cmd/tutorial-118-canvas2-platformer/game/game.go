// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package game is a platformer of the gostrict subset, which runs in the browser and draws on a 2D canvas. The
// player runs with the arrow keys or A and D, jumps with the space bar, W or the up key and collects the coins on
// the way to the flag. On a touch screen, the left and the right third of the canvas run and the middle jumps.
package game

import (
	"math"

	"go.wdy.de/nago/example/cmd/tutorial-118-canvas2-platformer/game/proto"
	"go.wdy.de/nago/pkg/gostrict/lib/asset"
	"go.wdy.de/nago/pkg/gostrict/lib/canvas"
	"go.wdy.de/nago/pkg/gostrict/lib/input"
	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/msg"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

// Tile is the content of a cell of the level.
type Tile int32

const (
	Empty Tile = iota
	Ground
	Coin
	Flag
)

const (
	size     = 32.0
	gravity  = 1800.0
	jumpV    = -640.0
	runV     = 220.0
	maxFall  = 900.0
	playerW  = 22.0
	playerH  = 28.0
	spriteSz = 16.0
)

// The level: # is ground, o a coin, F the flag and P the start of the player.
var level = list.Freeze(list.Of(
	"..............................................................",
	"..............................................................",
	"...........................o.o................................",
	"..........................#####..........o.o.o.......F........",
	"..................o.o...................#######.....###.......",
	"..........o.....#######.......................................",
	".........###..................o...........................####",
	"....P.....................#########....o.o......####..........",
	".##########.......###..................###....................",
	"#######..######.........o.o.......................o.o.o.......",
	"#######..######.......#######...#####........##########.......",
	"##############################################################",
))

var ctx = canvas.Get()

// Game is the state of the game, which the component holds.
type Game struct {
	Player   string
	Sprites  canvas.Image
	Ready    bool
	Tiles    list.List[Tile]
	Cols     int32
	Rows     int32
	StartX   float64
	StartY   float64
	X, Y     float64
	VX, VY   float64
	OnGround bool
	Facing   float64
	Left     bool
	Right    bool
	Jump     bool
	Coins    int32
	Time     float64
	Last     float64
	Done     bool
	Width    float64
	Height   float64
	Best     float64
	Holder   string
	// pointers maps the touch pointers to the direction, which they hold: -1, 0 for a jump or 1
	Pointers list.List[Touch]
}

// Touch is a pointer, which runs or jumps.
type Touch struct {
	ID  int32
	Dir int32
}

// Init loads the level and the sprite sheet at the URL and asks the backend for the highscore.
func Init(player, sprites string) (*Game, msg.Out[proto.Hello]) {
	g := &Game{Player: player, Sprites: canvas.LoadImage(sprites), Tiles: list.New[Tile](), Facing: 1, Pointers: list.New[Touch]()}
	for y, row := range level.All() {
		chars := text.Chars(row)
		g.Cols = chars.Len()
		for x, c := range chars.All() {
			t := Empty
			switch c {
			case "#":
				t = Ground
			case "o":
				t = Coin
			case "F":
				t = Flag
			case "P":
				g.StartX = float64(x) * size
				g.StartY = float64(y) * size
			}
			g.Tiles.Append(t)
		}
		g.Rows = y + 1
	}
	restart(g)

	out := msg.New[proto.Hello]()
	out.Send(proto.Hello{Player: player})
	return g, out
}

func restart(g *Game) {
	g.X, g.Y = g.StartX, g.StartY
	g.VX, g.VY = 0, 0
	g.Time = 0
	g.Done = false
	g.Coins = 0
	// the coins come back
	for y, row := range level.All() {
		for x, c := range text.Chars(row).All() {
			if c == "o" {
				g.Tiles.Set(y*g.Cols+x, Coin)
			}
		}
	}
}

// Resize takes the size of the canvas.
func Resize(g *Game, w, h float64) {
	g.Width, g.Height = w, h
}

func tileAt(g *Game, cx, cy int32) Tile {
	if cx < 0 || cx >= g.Cols {
		return Ground
	}
	if cy < 0 || cy >= g.Rows {
		return Empty
	}
	t, ok := g.Tiles.Get(cy*g.Cols + cx)
	if !ok {
		return Empty
	}
	return t
}

func cell(v float64) int32 {
	return num.Trunc32(math.Floor(v / size))
}

// solid tells, whether the rectangle overlaps ground.
func solid(g *Game, x, y, w, h float64) bool {
	for cy := cell(y); cy <= cell(y+h-0.01); cy++ {
		for cx := cell(x); cx <= cell(x+w-0.01); cx++ {
			if tileAt(g, cx, cy) == Ground {
				return true
			}
		}
	}
	return false
}

// Frame advances the game to the time t in seconds and draws it.
func Frame(g *Game, t float64) msg.Out[proto.Finished] {
	out := msg.New[proto.Finished]()
	dt := t - g.Last
	g.Last = t
	if dt > 0.05 || dt < 0 {
		dt = 0.05
	}
	if !g.Done {
		step(g, dt, out)
	}
	draw(g)
	return out
}

func step(g *Game, dt float64, out msg.Out[proto.Finished]) {
	g.Time += dt
	g.VX = 0
	if g.Left {
		g.VX -= runV
		g.Facing = -1
	}
	if g.Right {
		g.VX += runV
		g.Facing = 1
	}
	if g.Jump && g.OnGround {
		g.VY = jumpV
	}
	g.VY = min(g.VY+gravity*dt, maxFall)

	// move and resolve each axis on its own, so that the player slides along walls
	nx := g.X + g.VX*dt
	if solid(g, nx, g.Y, playerW, playerH) {
		if g.VX > 0 {
			nx = float64(cell(nx+playerW))*size - playerW - 0.001
		} else if g.VX < 0 {
			nx = float64(cell(nx)+1) * size
		}
	}
	g.X = nx

	ny := g.Y + g.VY*dt
	g.OnGround = false
	if solid(g, g.X, ny, playerW, playerH) {
		if g.VY > 0 {
			ny = float64(cell(ny+playerH))*size - playerH - 0.001
			g.OnGround = true
		} else {
			ny = float64(cell(ny)+1) * size
		}
		g.VY = 0
	}
	g.Y = ny

	if g.Y > float64(g.Rows)*size {
		// fell out of the level
		restart(g)
		return
	}

	for cy := cell(g.Y); cy <= cell(g.Y+playerH); cy++ {
		for cx := cell(g.X); cx <= cell(g.X+playerW); cx++ {
			switch tileAt(g, cx, cy) {
			case Coin:
				g.Tiles.Set(cy*g.Cols+cx, Empty)
				g.Coins++
			case Flag:
				g.Done = true
				out.Send(proto.Finished{Player: g.Player, Seconds: g.Time, Coins: g.Coins})
			case Empty, Ground:
			}
		}
	}
}

func draw(g *Game) {
	w, h := g.Width, g.Height
	sky := canvas.NewLinearGradient(ctx, 0, 0, 0, h)
	canvas.AddColorStop(sky, 0, 0x7ec8f0ff)
	canvas.AddColorStop(sky, 1, 0xd8f0ffff)
	canvas.FillPaint(ctx, sky)
	canvas.FillRect(ctx, 0, 0, w, h)

	// the camera follows the player and stays within the level
	camX := min(max(g.X-w/2, 0), max(float64(g.Cols)*size-w, 0))
	camY := min(max(g.Y-h/2, 0), max(float64(g.Rows)*size-h, 0))
	canvas.Save(ctx)
	canvas.Translate(ctx, -math.Round(camX), -math.Round(camY))
	canvas.ImageSmoothing(ctx, false)

	for cy := cell(camY); cy <= cell(camY+h); cy++ {
		for cx := cell(camX); cx <= cell(camX+w); cx++ {
			x, y := float64(cx)*size, float64(cy)*size
			switch tileAt(g, cx, cy) {
			case Ground:
				sprite(g, 0, x, y, 0x8b5a2bff)
			case Coin:
				sprite(g, 1, x, y, 0xffd93dff)
			case Flag:
				sprite(g, 3, x, y, 0xff4d4dff)
			case Empty:
			}
		}
	}

	canvas.Save(ctx)
	canvas.Translate(ctx, g.X+playerW/2, g.Y+playerH/2)
	canvas.Scale(ctx, g.Facing, 1)
	sprite(g, 2, -size/2, -size/2-2, 0x4d96ffff)
	canvas.Restore(ctx)
	canvas.Restore(ctx)

	hud(g)
}

// sprite draws the frame of the sprite sheet or a rectangle of the color, while the sheet is loading.
func sprite(g *Game, frame int32, x, y float64, color uint32) {
	if !g.Ready {
		canvas.FillColor(ctx, color)
		canvas.FillRect(ctx, x, y, size, size)
		return
	}
	canvas.DrawImageRegion(ctx, g.Sprites, float64(frame)*spriteSz, 0, spriteSz, spriteSz, x, y, size, size)
}

func hud(g *Game) {
	canvas.Font(ctx, "bold 16px sans-serif")
	canvas.TextBaseline(ctx, "top")
	canvas.FillColor(ctx, 0x00000099)
	canvas.FillRect(ctx, 8, 8, 230, 52)
	canvas.FillColor(ctx, 0xffffffff)
	canvas.FillText(ctx, "Coins "+text.FromInt(g.Coins)+"   Time "+text.FromFloat(g.Time, 1)+" s", 16, 14)
	best := "no highscore yet"
	if g.Holder != "" {
		best = "Best " + text.FromFloat(g.Best, 2) + " s by " + g.Holder
	}
	canvas.FillText(ctx, best, 16, 36)

	if g.Done {
		canvas.FillColor(ctx, 0x000000aa)
		canvas.FillRect(ctx, 0, g.Height/2-40, g.Width, 80)
		canvas.FillColor(ctx, 0xffffffff)
		canvas.TextAlign(ctx, "center")
		canvas.FillText(ctx, "Finished in "+text.FromFloat(g.Time, 2)+" s. Press R or tap to restart.", g.Width/2, g.Height/2-8)
		canvas.TextAlign(ctx, "left")
	}
}

// OnInput controls the player with the keys or the pointers.
func OnInput(g *Game, e input.Event) {
	if e.Type == input.KeyDown || e.Type == input.KeyUp {
		down := e.Type == input.KeyDown
		switch e.Code {
		case "ArrowLeft", "KeyA":
			g.Left = down
		case "ArrowRight", "KeyD":
			g.Right = down
		case "Space", "ArrowUp", "KeyW":
			g.Jump = down
		case "KeyR":
			if down && g.Done {
				restart(g)
			}
		}
		return
	}

	if e.Type == input.PointerDown {
		if g.Done {
			restart(g)
			return
		}
		dir := int32(0)
		if e.X < g.Width/3 {
			dir = -1
		} else if e.X > g.Width*2/3 {
			dir = 1
		}
		g.Pointers.Append(Touch{ID: e.PointerID, Dir: dir})
		touches(g)
		return
	}
	if e.Type == input.PointerUp || e.Type == input.PointerCancel {
		for i, p := range g.Pointers.All() {
			if p.ID == e.PointerID {
				g.Pointers.Delete(i)
				break
			}
		}
		touches(g)
	}
}

// touches derives the controls from the pointers, which are down.
func touches(g *Game) {
	g.Left, g.Right, g.Jump = false, false, false
	for _, p := range g.Pointers.All() {
		switch p.Dir {
		case -1:
			g.Left = true
		case 1:
			g.Right = true
		default:
			g.Jump = true
		}
	}
}

// OnLoaded starts to draw with the sprite sheet.
func OnLoaded(g *Game, e asset.Loaded) {
	g.Ready = true
}

// OnHighscore takes the best time, which the backend sends.
func OnHighscore(g *Game, h proto.Highscore) {
	g.Best = h.Seconds
	g.Holder = h.Holder
}
