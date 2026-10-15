// Package grid simulates a player falling and jumping on a tile grid.
package grid

import (
	"math"

	"go.wdy.de/nago/pkg/gostrict/std/list"
	"go.wdy.de/nago/pkg/gostrict/std/num"
	"go.wdy.de/nago/pkg/gostrict/std/text"
)

const (
	tileSize = 16.0
	gravity  = 0.5
	maxFall  = 8.0
	jumpVel  = -7.0
	runSpeed = 1.5
)

type Number interface {
	int32 | float64
}

// Clamp limits v to the range [lo, hi].
func Clamp[T Number](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type Grid struct {
	Width  int32
	Height int32
	Tiles  list.List[bool]
}

func NewGrid(w, h int32) Grid {
	return Grid{Width: w, Height: h, Tiles: list.Make[bool](w * h)}
}

func (g Grid) Set(x, y int32, solid bool) {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return
	}
	g.Tiles.Set(y*g.Width+x, solid)
}

// Solid reports whether the tile is solid; everything outside the grid is solid.
func (g Grid) Solid(x, y int32) bool {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return true
	}
	v, ok := g.Tiles.Get(y*g.Width + x)
	return ok && v
}

func (g Grid) SolidAt(px, py float64) bool {
	return g.Solid(num.Trunc32(math.Floor(px/tileSize)), num.Trunc32(math.Floor(py/tileSize)))
}

type Player struct {
	X, Y     float64
	VX, VY   float64
	W, H     float64
	OnGround bool
	Landings int32
}

func (p *Player) collides(g Grid, x, y float64) bool {
	return g.SolidAt(x, y) || g.SolidAt(x+p.W-0.01, y) ||
		g.SolidAt(x, y+p.H-0.01) || g.SolidAt(x+p.W-0.01, y+p.H-0.01)
}

func (p *Player) Step(g Grid) {
	p.VY = Clamp(p.VY+gravity, jumpVel, maxFall)
	if p.OnGround {
		p.VY = jumpVel
		p.OnGround = false
	}

	// horizontal
	nx := p.X + p.VX
	if p.collides(g, nx, p.Y) {
		p.VX = -p.VX
	} else {
		p.X = nx
	}

	// vertical
	ny := p.Y + p.VY
	if p.collides(g, p.X, ny) {
		if p.VY > 0 {
			// snap onto the tile below
			tile := num.Trunc32(math.Floor((ny + p.H) / tileSize))
			p.Y = float64(tile)*tileSize - p.H
			p.OnGround = true
			p.Landings++
		} else {
			tile := num.Trunc32(math.Floor(ny / tileSize))
			p.Y = float64(tile+1) * tileSize
		}
		p.VY = 0
	} else {
		p.Y = ny
	}
	maxX := float64(g.Width)*tileSize - p.W
	p.X = Clamp(p.X, 0, maxX)
}

func Simulate(steps int32) string {
	g := NewGrid(20, 12)
	for x := range int32(20) {
		g.Set(x, 11, true)
	}
	for x := int32(5); x < 9; x++ {
		g.Set(x, 7, true)
	}
	g.Set(14, 10, true)
	g.Set(14, 9, true)

	p := &Player{X: 32, Y: 16, VX: runSpeed, W: 12, H: 14}
	for range steps {
		p.Step(g)
	}
	return "x=" + text.FromInt(num.Trunc32(p.X)) + " y=" + text.FromInt(num.Trunc32(p.Y)) +
		" landings=" + text.FromInt(p.Landings)
}

func Demo() string {
	return Simulate(300)
}
